package logdb

import (
	"fmt"
	"github.com/worldiety/ioutil"
	"io"
	"os"
	"runtime"
	"sync"
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
	header             *Header
}

func Open(fname string) (*DB, error) {
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
	if err != nil {
		return nil, err
	}
	db.objPool = sync.Pool{
		New: func() interface{} { return newObject(db.maxObjSize) },
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

	n, err := db.file.WriteAt(tmp, db.eof)
	if err != nil {
		return err
	}

	if n != len(tmp) {
		return fmt.Errorf("file did not accept full buffer")
	}

	db.eof += int64(len(tmp))
	record.Reset()

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

func (db *DB) ForEachP(f func(id uint64, obj *Object) error) error {
	records, err := db.findRecords()
	if err != nil {
		return err
	}

	fmt.Printf("found %d records\n", len(records))

	recordQueue := make(chan int64, len(records))
	for _, r := range records {
		recordQueue <- r
	}

	wg := sync.WaitGroup{}
	wg.Add(runtime.NumCPU())

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			defer wg.Done()

			record := newRecord(db.pendingWriteRecord.MaxSize())
			obj := newObject(db.maxObjSize)

			for {

				var offset int64
				select {
				case offset = <-recordQueue:
				default:
					return
				}




				lBuf, err := db.file.ReadAt(record.buf.Bytes, offset)
				if err != nil {
					if err != io.EOF {
						panic(err)
					}
				}

				if lBuf < offsetRecObjList {
					panic(fmt.Errorf("unable to read a record at offset %d", offset))
				}

				record.reverseFlush()
				err = record.ForEach(obj, func(recOffset int, object *Object) error {
					return f(uint64(offset)+uint64(recOffset), object)
				})

				offset += int64(record.Size())

				if err != nil {
					panic(err)
				}
			}

		}()
	}

	wg.Wait()

	return nil
}
