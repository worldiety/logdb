package logdb

func (d *Object) AddFloat(name uint16, v float64) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint16(name)
	(*FieldWriter)(d.buf).WriteFloat(v)
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}

func (d *Object) AddInt(name uint16, v int64) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint16(name)
	(*FieldWriter)(d.buf).WriteInt(v)
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}

func (d *Object) AddInt8(name uint16, v int8) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint16(name)
	(*FieldWriter)(d.buf).WriteUint8(uint8(v))
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}

func (d *Object) AddUint32(name uint16, v uint32) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint16(name)
	(*FieldWriter)(d.buf).WriteUint32(v)
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}

func (d *Object) AddString(name uint16, v string) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint16(name)
	(*FieldWriter)(d.buf).WriteString(v)
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}
