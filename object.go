package logdb

import "github.com/worldiety/ioutil"

const (
	offsetSize       = 0
	offsetFieldCount = 3
	offsetFieldList  = 5
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

	fieldReaderNum int16
	fieldReaderName uint16
	fieldReaderType ioutil.Type
	fieldReaderDrainPos int
}

func newObject(maxSize int) *Object {
	o := &Object{buf: &ioutil.LittleEndianBuffer{
		Bytes: make([]byte, maxSize),
	}}
	o.resetWrite()
	return o
}

func (d *Object)DeepClone()*Object{
	tmp := make([]byte,len(d.buf.Bytes))
	copy(tmp,d.buf.Bytes)

	return &Object{
		buf:                 &ioutil.LittleEndianBuffer{
			Bytes: tmp,
			Pos:   0,
		},
		size:                d.size,
		fieldCount:          d.fieldCount,
		fieldReaderNum:      d.fieldReaderNum,
		fieldReaderName:     d.fieldReaderName,
		fieldReaderType:     d.fieldReaderType,
		fieldReaderDrainPos: d.fieldReaderDrainPos,
	}
}

func (d *Object)Clone()*Object{
	return &Object{
		buf:                 &ioutil.LittleEndianBuffer{
			Bytes: d.buf.Bytes,
			Pos:   0,
		},
		size:                d.size,
		fieldCount:          d.fieldCount,
		fieldReaderNum:      d.fieldReaderNum,
		fieldReaderName:     d.fieldReaderName,
		fieldReaderType:     d.fieldReaderType,
		fieldReaderDrainPos: d.fieldReaderDrainPos,
	}
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

func (d *Object) FieldReaderReset(){
	d.fieldReaderNum = -1
	d.buf.Pos = offsetFieldList
	d.fieldReaderType = 0
}

func (d *Object) FieldReaderName()uint16{
	return d.fieldReaderName
}

func (d *Object) FieldReaderNext()bool{
	d.fieldReaderNum++
	if uint16(d.fieldReaderNum) < d.fieldCount{

		d.fieldReaderName = d.buf.ReadUint16()
		//d.fieldReaderType = d.buf.ReadType()
		//d.fieldReaderDrainPos = d.buf.Pos
		//d.buf.Pos--

		return true
	}

	return false
}

func(d *Object) FieldReader()*FieldReader{
	return (*FieldReader)(d.buf)
}




// WithFields iterates over each available field. This is the fastest
// thing we can do.
func (d *Object) WithFields(f func(name uint16, kind ioutil.Type, f *FieldReader)) {
	count := int(d.FieldCount())
	d.buf.Pos = offsetFieldList
	for i := 0; i < count; i++ {
		name := d.buf.ReadUint16()
		kind := d.buf.ReadType()
		myDrainPos := d.buf.Pos
		d.buf.Pos--
		f(name, kind, (*FieldReader)(d.buf))

		// reset the pos to ensure we are correct, independently what f has done
		d.buf.Pos = myDrainPos
		if i := d.buf.DrainFast(kind); i == -1 {
			d.buf.Drain(kind)
		}
	}
}

// AddField appends another field.
func (d *Object) AddField(name uint16, f func(f *FieldWriter)) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint16(name)
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
