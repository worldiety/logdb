package logdb

const (
	offsetSize       = 0
	offsetFieldCount = 4
	offsetFieldList  = 8
)

// An Object contains all field data and fits into memory.
//
// Format specification:
//  - size                uint32, including size (4 byte)
//  - fieldCount          uint32
//  - []                  variable, depending on count
//     {
//       - name           uint32
//       - fieldType      uint8
//       - value          variable, depending on type
//     }
type Object struct {
	buf *LEBuffer
}

func newObject(maxSize int) *Object {
	o := &Object{buf: &LEBuffer{
		Buf: make([]byte, maxSize),
	}}
	o.Reset()
	return o
}

func (d *Object) Size() uint32 {
	d.buf.Pos = offsetSize
	return d.buf.ReadUint32()
}

func (d *Object) setSize(s uint32) {
	d.buf.Pos = offsetSize
	d.buf.WriteUint32(s)
}

func (d *Object) FieldCount() uint32 {
	d.buf.Pos = offsetFieldCount
	return d.buf.ReadUint32()
}

func (d *Object) setFieldCount(v uint32) {
	d.buf.Pos = offsetFieldCount
	d.buf.WriteUint32(v)
}

func (d *Object) Bytes() []byte {
	return d.buf.Buf[:d.Size()]
}

// WithFields iterates over each available field. This is the fastest
// thing we can do.
func (d *Object) WithFields(f func(name uint32, kind DataType, f *LEBuffer)) {
	count := int(d.FieldCount())
	d.buf.Pos = offsetFieldList
	for i := 0; i < count; i++ {
		name := d.buf.ReadUint32()
		kind := DataType(d.buf.ReadUint8())
		myDrainPos := d.buf.Pos
		f(name, kind, d.buf)

		// reset the pos to ensure we are correct, independently what f has done
		d.buf.Pos = myDrainPos
		d.buf.Drain(kind)
	}
}

// AddField appends another field.
func (d *Object) AddField(name uint32, kind DataType, f func(f *LEBuffer)) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint32(name)
	d.buf.WriteUint8(uint8(kind))
	f(d.buf)
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}

// Reset sets the length to the minimum length
func (d *Object) Reset() {
	d.setSize(offsetFieldList)
	d.setFieldCount(0)
}
