package logdb

import (
	"fmt"
	"io"
	"os"
)

type DB struct {
	file               *os.File
	eof                int64
	tmpWriteObj        *Object
	maxObjSize         int
	pendingWriteRecord *Record
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

	db := &DB{file: file, eof: stat.Size()}
	db.maxObjSize = 1024 * 64                               // 64k max object size
	db.pendingWriteRecord = newRecord(db.maxObjSize * 1000) // 64MB max batch size
	db.tmpWriteObj = newObject(db.maxObjSize)
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
	obj.Reset()
	err := f(obj)
	if err != nil {
		return err
	}
	record.Add(obj)
	return nil
}

func (db *DB) Flush() error {
	record := db.pendingWriteRecord

	tmp := record.Bytes()
	if len(tmp) == offsetRecObjList{
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

func (db *DB) ForEach(f func(obj *Object) error) error {
	record := newRecord(db.pendingWriteRecord.MaxSize())
	obj := newObject(db.maxObjSize)

	offset := int64(0)
	for offset < db.eof {
		lBuf, err := db.file.ReadAt(record.buf.Buf, offset)
		if err != nil {
			if err != io.EOF {
				return err
			}
		}

		if lBuf < offsetRecObjList {
			return fmt.Errorf("unable to read a record at offset %d", offset)
		}

		offset += int64(record.Size())

		err = record.ForEach(obj, func(object *Object) error {
			return f(object)
		})

		if err != nil {
			return err
		}

	}

	return nil
}
