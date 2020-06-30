package logdb

import "github.com/worldiety/ioutil"

const (
	offsetSize       = 0
	offsetFieldCount = 3
	offsetFieldList  = 7
)

// An Object contains all field data and fits into memory.
//
// Format specification:
//  - size                uint24, including size (3 byte), at most 16MiB
//  - fieldCount          uint16, at most 65.536
//  - []                  variable, depending on count
//     {
//       - name           uint16, at most 65.536 per object file
//       - fieldType      uint8
//       - value          variable, depending on type
//     }
type Object struct {
	buf        *ioutil.LittleEndianBuffer
	size       uint32
	fieldCount uint16
}

func newObject(maxSize int) *Object {
	o := &Object{buf: &ioutil.LittleEndianBuffer{
		Bytes: make([]byte, maxSize),
	}}
	o.resetWrite()
	return o
}

func (d *Object) Size() uint32 {
	return d.size
}

func (d *Object) setSize(s uint32) {
	d.size = s
}

func (d *Object) FieldCount() uint16 {
	return d.fieldCount
}

func (d *Object) setFieldCount(v uint16) {
	d.fieldCount = v
}

// flush writes some numbers into the buffer, like size and field count
func (d *Object) flush() {

	d.buf.Pos = offsetFieldCount
	d.buf.WriteUint16(d.fieldCount)

	d.buf.Pos = offsetSize
	d.buf.WriteUint24(d.size)
}

func (d *Object) Bytes() []byte {
	return d.buf.Bytes[:d.Size()]
}

// WithFields iterates over each available field. This is the fastest
// thing we can do.
func (d *Object) WithFields(f func(name uint32, kind ioutil.Type, f *FieldReader)) {
	count := int(d.FieldCount())
	d.buf.Pos = offsetFieldList
	for i := 0; i < count; i++ {
		name := d.buf.ReadUint16()
		kind := d.buf.ReadType()
		myDrainPos := d.buf.Pos
		d.buf.Pos--
		f(uint32(name), kind, (*FieldReader)(d.buf))

		// reset the pos to ensure we are correct, independently what f has done
		d.buf.Pos = myDrainPos
		d.buf.Drain(kind)
	}
}

// AddField appends another field.
func (d *Object) AddField(name uint32, f func(f *FieldWriter)) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint16(uint16(name))
	f((*FieldWriter)(d.buf))
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}

// Reset sets the length to the minimum length
func (d *Object) resetWrite() {
	d.setSize(offsetFieldList)
	d.setFieldCount(0)
}

func (d *Object) reverseFlush() {
	d.buf.Pos = offsetSize
	d.size = d.buf.ReadUint24()

	d.buf.Pos = offsetFieldCount
	d.fieldCount = d.buf.ReadUint16()
}
