package logdb

func (d *Object) AddFloat32(name uint32, v float32) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint32(name)
	d.buf.WriteUint8(uint8(TFloat32))
	d.buf.WriteFloat32(v)
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}

func (d *Object) AddFloat64(name uint32, v float64) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint32(name)
	d.buf.WriteUint8(uint8(TFloat64))
	d.buf.WriteFloat64(v)
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}

func (d *Object) AddUint32(name uint32, v uint32) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint32(name)
	d.buf.WriteUint8(uint8(TUint32))
	d.buf.WriteUint32(v)
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}

func (d *Object) AddInt16(name uint32, v int16) {
	d.AddUint16(name, uint16(v))
}

func (d *Object) AddUint16(name uint32, v uint16) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint32(name)
	d.buf.WriteUint8(uint8(TUint16))
	d.buf.WriteUint16(v)
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}

func (d *Object) AddInt8(name uint32, v int8) {
	d.AddUint8(name, uint8(v))
}

func (d *Object) AddUint8(name uint32, v uint8) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint32(name)
	d.buf.WriteUint8(uint8(TUint8))
	d.buf.WriteUint8(v)
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}

func (d *Object) AddInt64(name uint32, v int64) {
	d.AddUint64(name, uint64(v))
}

func (d *Object) AddUint64(name uint32, v uint64) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint32(name)
	d.buf.WriteUint8(uint8(TUint64))
	d.buf.WriteUint64(v)
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}

func (d *Object) AddTinyString(name uint32, s string) {
	count := d.FieldCount()
	d.buf.Pos = int(d.Size())
	d.buf.WriteUint32(name)
	d.buf.WriteUint8(uint8(TTinyBlob))
	d.buf.WriteTinyString(s)
	d.setSize(uint32(d.buf.Pos))
	d.setFieldCount(count + 1)
}
