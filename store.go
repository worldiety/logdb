package logdb

import (
	"fmt"
	"github.com/pierrec/lz4"
	"github.com/worldiety/ioutil"
	"io"
	"os"
	"runtime"
	"sync"
	"syscall"
)

type DB struct {
	file               *os.File
	eof                int64
	tmpWriteObj        *Object
	maxObjSize         int
	pendingWriteRecord *Record
	maxRecSize         int
	reader             *concurrentCachedReader
	objPool            sync.Pool
	recPool            sync.Pool
	header             *Header
	mmapFile           []byte
	useMmap            bool
	compress           bool
	compressHashtable  []int
}

func Open(fname string, useMmap bool, compression bool) (*DB, error) {
	file, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fmt.Printf("db size is %d (%d MiB)\n", stat.Size(), stat.Size()/1024/1024)

	db := &DB{file: file, eof: stat.Size()}
	db.maxObjSize = 1024 * 64                                           // 64k max object size
	db.maxRecSize = db.maxObjSize * 1000                                // 64MB max batch size
	db.header = newHeader(int(ioutil.MaxUint8) * int(ioutil.MaxUint16)) // 16mb
	db.pendingWriteRecord = newRecord(db.maxRecSize)
	db.tmpWriteObj = newObject(db.maxObjSize)
	db.reader, err = newConcurrentCachedReader(db.file, db.maxRecSize)
	db.useMmap = useMmap
	db.compressHashtable = make([]int, 1<<16)
	db.compress = true

	if useMmap {
		data, err := syscall.Mmap(int(file.Fd()), 0, int(stat.Size()), syscall.PROT_READ, syscall.MAP_PRIVATE)
		if err != nil {
			return nil, fmt.Errorf("error mmap: %w", err)
		}
		db.mmapFile = data
	}

	if err != nil {
		return nil, err
	}
	db.objPool = sync.Pool{
		New: func() interface{} { return newObject(db.maxObjSize) },
	}

	db.recPool = sync.Pool{
		New: func() interface{} { return newRecord(db.maxRecSize) },
	}

	if db.eof == 0 {
		db.header.Flush()
		_, err := db.file.Write(db.header.buf.Bytes)
		if err != nil {
			_ = db.file.Close()
			return nil, fmt.Errorf("unable to create db header: %w", err)
		}
		db.eof = int64(len(db.header.buf.Bytes))

	} else {
		if db.eof < int64(len(db.header.buf.Bytes)) {
			_ = db.file.Close()
			return nil, fmt.Errorf("truncated database file, header to short: %w", err)
		} else {
			_, err := db.file.Read(db.header.buf.Bytes)
			if err != nil {
				_ = db.file.Close()
				return nil, fmt.Errorf("unable to read header: %w", err)
			}
			db.header.reverseFlush()
		}
	}

	fmt.Printf("names: %d\n", db.header.nameCount)
	fmt.Printf("objects: %d\n", db.header.ObjectCount())
	fmt.Printf("last transaction: %d\n", db.header.TxCount())

	return db, nil
}

func (db *DB) ObjectCount() uint64 {
	return db.header.ObjectCount()
}

func (db *DB) NameByIndex(idx int) string {
	return db.header.NameByIndex(idx)
}

func (db *DB) IndexByName(name string) int {
	return db.header.IndexByName(name)
}

func (db *DB) PutName(name string) uint16 {
	return uint16(db.header.AddName(name))
}

func (db *DB) Names() []string {
	return db.header.Names()
}

func (db *DB) Add(f func(obj *Object) error) error {
	record := db.pendingWriteRecord
	if record.MaxSize()-int(record.Size()) < db.maxObjSize {
		if err := db.Flush(); err != nil {
			return err
		}
	}

	obj := db.tmpWriteObj
	obj.resetWrite()
	err := f(obj)
	if err != nil {
		return err
	}

	obj.flush()
	record.Add(obj)
	db.header.AddObjectCount(1)
	return nil
}

func (db *DB) Flush() error {
	record := db.pendingWriteRecord
	record.flush()

	tmp := record.Bytes()
	if len(tmp) == offsetRecObjList {
		return nil
	}

	if db.compress {
		compressedRec := db.recPool.Get().(*Record)
		defer db.recPool.Put(compressedRec)

		n, err := lz4.CompressBlock(tmp, compressedRec.buf.Bytes, db.compressHashtable)
		if err != nil {
			return fmt.Errorf("failed to compress: %w", err)
		}

		length := ioutil.LittleEndianBuffer{
			Bytes: make([]byte, 4),
			Pos:   0,
		}
		length.WriteUint32(uint32(n))
		db.file.WriteAt(length.Bytes, db.eof)
		db.eof += 4

		ctmp := compressedRec.buf.Bytes[:n]
		n, err = db.file.WriteAt(ctmp, db.eof)

		if n != len(ctmp) {
			return fmt.Errorf("file did not accept full cbuffer")
		}

		db.eof += int64(len(ctmp))
		record.Reset()

	} else {
		n, err := db.file.WriteAt(tmp, db.eof)
		if err != nil {
			return err
		}

		if n != len(tmp) {
			return fmt.Errorf("file did not accept full buffer")
		}

		db.eof += int64(len(tmp))
		record.Reset()
	}

	db.header.AddTxCount(1)
	return nil
}

func (db *DB) flushHeader() error {
	header := db.header
	header.Flush()

	_, err := db.file.WriteAt(header.buf.Bytes, 0)
	return err
}

func (db *DB) Close() error {
	if err := db.Flush(); err != nil {
		return err
	}

	if err := db.flushHeader(); err != nil {
		return err
	}

	return db.file.Close()
}

// Read seeks to the id (currently just the offset) and reads the object. It is safe to be used
// concurrently.
func (db *DB) Read(id uint64, f func(obj *Object) error) error {
	obj := db.objPool.Get().(*Object)
	defer db.objPool.Put(obj)

	_, err := db.reader.ReadAt(int64(id), obj.buf.Bytes)
	obj.reverseFlush()
	if err != nil {
		return err
	}
	if err := f(obj); err != nil {
		return err
	}

	return nil
}

// ForEach is safe to be used concurrently. It allocates its own buffer
// on each call, which is an easy design and is negligible for large datasets.
func (db *DB) ForEach(f func(id uint64, obj *Object) error) error {
	record := newRecord(db.pendingWriteRecord.MaxSize())
	obj := newObject(db.maxObjSize)

	offset := int64(len(db.header.buf.Bytes))
	for offset < db.eof {
		lBuf, err := db.file.ReadAt(record.buf.Bytes, offset)
		if err != nil {
			if err != io.EOF {
				return err
			}
		}

		if lBuf < offsetRecObjList {
			return fmt.Errorf("unable to read a record at offset %d", offset)
		}

		record.reverseFlush()
		err = record.ForEach(obj, func(recOffset int, object *Object) error {
			return f(uint64(offset)+uint64(recOffset), object)
		})

		offset += int64(record.Size())

		if err != nil {
			return err
		}

	}

	return nil
}


// findRecords parses over the entire file and returns all record offset
func (db *DB) findRecords() ([]int64, error) {
	res := make([]int64, 0, db.header.txCount)

	if db.compress {
		tmp := make([]byte, 4)
		buf := ioutil.LittleEndianBuffer{
			Bytes: tmp,
			Pos:   0,
		}

		offset := int64(len(db.header.buf.Bytes))
		for offset < db.eof {
			res = append(res, offset)
			_, err := db.file.ReadAt(tmp, offset)
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
			}
			offset+=4

			buf.Pos = 0
			size := buf.ReadUint32()

			offset += int64(size)
		}
		return res, nil

	} else {

		tmp := make([]byte, offsetRecObjCount)
		buf := ioutil.LittleEndianBuffer{
			Bytes: tmp,
			Pos:   0,
		}

		offset := int64(len(db.header.buf.Bytes))
		for offset < db.eof {
			res = append(res, offset)
			lBuf, err := db.file.ReadAt(tmp, offset)
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
			}

			if lBuf < offsetRecObjCount {
				return nil, fmt.Errorf("unable to read a record bound at offset %d", offset)
			}

			buf.Pos = offsetRecSize
			size := buf.ReadUint32()

			offset += int64(size)
		}

		return res, nil
	}
}

// ForEachP walks parallel over records and makes things worse: looks like we are crashing the cpu-memory bandwidth barrier with
// more than 1 routine already. On big cloud machines this effect is even worse and makes everything
// slower (e.g. from 20m to 13m for 2 cores, without any locking effects). Instruments shows
// cache-misses increases on macos linearly.
func (db *DB) ForEachP(routines int, f func(gid int, id uint64, obj *Object) error) (e error) {
	records, err := db.findRecords()
	if err != nil {
		return err
	}

	fmt.Printf("found %d records\n", len(records))

	wg := sync.WaitGroup{}
	wg.Add(routines)

	batchSize := len(records) / routines // 11/2 = 5

	for i := 0; i < routines; i++ {
		fromRec, toRec := i*batchSize, (i+1)*batchSize
		if i == routines-1 {
			toRec = len(records)
		}

		go func(id, fromRec, toRec int) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			defer wg.Done()

			record := newRecord(db.pendingWriteRecord.MaxSize())
			compressedRec := newRecord(db.pendingWriteRecord.MaxSize())

			obj := newObject(db.maxObjSize)

			for r := fromRec; r < toRec; r++ {

				offset := records[r]

				if db.useMmap {
					max := db.maxRecSize
					avail := int(db.eof) - int(offset)
					if max > avail {
						max = avail
					}
					record.buf.Bytes = db.mmapFile[offset : int(offset)+max]
					if db.compress {
						panic("not yet implemented")
					}
				} else {
					if db.compress {
						db.file.ReadAt(compressedRec.buf.Bytes, offset)
						offset += 4
						clen := compressedRec.buf.ReadUint32()
						compressedRec.buf.Pos = 0

						buf := compressedRec.buf.Bytes[:clen]
						_, err := db.file.ReadAt(buf, offset)
						n, err := lz4.UncompressBlock(buf, record.buf.Bytes)
						if err != nil {
							panic(err)
						}

						if n < offsetRecObjList {
							panic(fmt.Errorf("unable to read a record at offset %d", offset))
						}

					} else {

						lBuf, err := db.file.ReadAt(record.buf.Bytes, offset)
						if err != nil {
							if err != io.EOF {
								panic(err)
							}
						}

						if lBuf < offsetRecObjList {
							panic(fmt.Errorf("unable to read a record at offset %d", offset))
						}
					}

				}

				record.reverseFlush()
				err = record.ForEach(obj, func(recOffset int, object *Object) error {
					return f(id, uint64(offset)+uint64(recOffset), object)
				})

				offset += int64(record.Size())

				if err != nil {
					if e != nil {
						e = err
					}
					return
				}
			}

		}(i, fromRec, toRec)
	}

	wg.Wait()

	return
}
