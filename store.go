package logdb

import (
	"fmt"
	"io"
	"os"
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
	db.maxObjSize = 1024 * 64            // 64k max object size
	db.maxRecSize = db.maxObjSize * 1000 // 64MB max batch size
	db.pendingWriteRecord = newRecord(db.maxRecSize)
	db.tmpWriteObj = newObject(db.maxObjSize)
	db.reader, err = newConcurrentCachedReader(db.file, db.maxRecSize)
	if err != nil {
		return nil, err
	}
	db.objPool = sync.Pool{
		New: func() interface{} { return newObject(db.maxObjSize) },
	}
	return db, nil
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

	return nil
}

func (db *DB) Close() error {
	if err := db.Flush(); err != nil {
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

	offset := int64(0)
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
