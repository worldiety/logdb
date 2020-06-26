package logdb

import (
	"io"
	"os"
	"sync"
)

type concurrentCachedReader struct {
	file             *os.File
	record           *Record
	recordOffset     int64
	mutex            sync.RWMutex
	actualRecordSize int64
}

func newConcurrentCachedReader(file *os.File, maxRecordSize int) (*concurrentCachedReader, error) {
	r := &concurrentCachedReader{
		file:         file,
		record:       newRecord(maxRecordSize),
		recordOffset: 0,
		mutex:        sync.RWMutex{},
	}
	return r, r.pageIn(0)
}

func (r *concurrentCachedReader) inPage(from, to int64) bool {
	return r.recordOffset <= from && r.recordOffset+r.actualRecordSize >= to
}

func (r *concurrentCachedReader) pageIn(offset int64) error {
	r.recordOffset = offset
	n, err := r.file.ReadAt(r.record.buf.Bytes, offset)
	r.actualRecordSize = int64(n)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	return nil
}

func (r *concurrentCachedReader) ReadAt(offset int64, dst []byte) (int, error) {
	r.mutex.RLock()
	from, to := offset, offset+int64(len(dst))
	if !r.inPage(from, to) {
		r.mutex.RUnlock()

		r.mutex.Lock()
		defer r.mutex.Unlock()

		err := r.pageIn(offset)
		if err != nil {
			return 0, err
		}
		return copy(dst, r.record.buf.Bytes[offset-r.recordOffset:r.actualRecordSize]), nil
	}
	defer r.mutex.RUnlock()
	return copy(dst, r.record.buf.Bytes[offset-r.recordOffset:r.actualRecordSize]), nil
}
