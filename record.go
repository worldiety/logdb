package logdb

import (
	"github.com/worldiety/ioutil"
)

const (
	offsetRecMagic    = 0
	offsetRecSize     = 8
	offsetRecObjCount = 12
	offsetRecObjList  = 16
)

var recordMagic = [8]byte{'w', 'd', 'y', 'r', 'e', 'c', '0', '1'}

type Record struct {
	magic    [8]byte
	buf      *ioutil.LittleEndianBuffer
	size     uint32
	objCount uint32
}

func newRecord(maxSize int) *Record {
	r := &Record{buf: &ioutil.LittleEndianBuffer{
		Bytes: make([]byte, maxSize)},
		magic: recordMagic}
	r.Reset()
	return r
}

func (d *Record) MaxSize() int {
	return len(d.buf.Bytes)
}

func (d *Record) Size() uint32 {
	return d.size
}

func (d *Record) setSize(s uint32) {
	d.size = s
}

func (d *Record) ObjectCount() uint32 {
	return d.objCount
}

func (d *Record) setObjectCount(v uint32) {
	d.objCount = v
}

func (d *Record) Bytes() []byte {
	return d.buf.Bytes[:d.Size()]
}

func (d *Record) Reset() {
	d.setSize(offsetRecObjList)
	d.setObjectCount(0)
}

func (d *Record) Add(obj *Object) {
	count := d.ObjectCount()
	size := d.Size()
	tmp := obj.Bytes()
	d.buf.Pos = int(size)
	d.buf.WriteSlice(tmp)
	d.setSize(size + uint32(len(tmp)))
	d.setObjectCount(count + 1)
}

func (d *Record) At(offset int, tmp *Object, f func(offset int, object *Object) error) error {
	size := d.ReadUint24At(offset)
	objBuf := d.buf.Bytes[offset : offset+int(size)]
	tmp.buf.Bytes = objBuf
	tmp.reverseFlush()

	if err := f(offset, tmp); err != nil {
		return err
	}

	return nil
}

func (d *Record) ReadUint24At(pos int) uint32 {
	b := d.buf.Bytes

	_ = b[pos+2] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[pos]) | uint32(b[pos+1])<<8 | uint32(b[pos+2])<<16
}

func (d *Record) ObjOffsets(dst []int) int {
	count := int(d.ObjectCount())
	d.buf.Pos = offsetRecObjList
	for i := 0; i < count; i++ {
		dst[i] = d.buf.Pos
		size := int(d.buf.ReadUint24())
		d.buf.Pos += size-3
	}

	return count
}

func (d *Record) ForEach(tmp *Object, f func(offset int, object *Object) error) error {
	// we don't want another memcpy, so we slice into
	/*tmpBuf := tmp.buf.Bytes
	defer func() {
		tmp.buf.Bytes = tmpBuf
	}()*/

	count := int(d.ObjectCount())
	d.buf.Pos = offsetRecObjList
	_ = offsetSize // for documentation only
	for i := 0; i < count; i++ {



		size := d.buf.ReadUint24()
		d.buf.Pos -= 3



		// just slice it
		objBuf := d.buf.Bytes[d.buf.Pos : d.buf.Pos+int(size)]
		tmp.buf.Bytes = objBuf

		tmp.reverseFlush()

		if err := f(d.buf.Pos, tmp); err != nil {
			return err
		}

		d.buf.Pos += int(size)
	}

	return nil
}

func (d *Record) flush() {
	d.buf.Pos = offsetRecMagic
	d.buf.WriteSlice(d.magic[:])

	d.buf.Pos = offsetRecSize
	d.buf.WriteUint32(d.size)

	d.buf.Pos = offsetRecObjCount
	d.buf.WriteUint32(d.objCount)
}

func (d *Record) reverseFlush() {
	d.buf.Pos = offsetRecSize
	d.size = d.buf.ReadUint32()

	d.buf.Pos = offsetRecObjCount
	d.objCount = d.buf.ReadUint32()
}
