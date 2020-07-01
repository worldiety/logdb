package logdb

import (
	"bytes"
	"fmt"
	"github.com/worldiety/ioutil"
	ioutil2 "io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

const defaultTableSize = 10_000

func TestObj(t *testing.T) {
	obj := newObject(1024 * 64)
	obj.AddField(1, func(f *FieldWriter) {
		f.WriteFloat(3)
	})

	if obj.FieldCount() != 1 {
		t.Fatalf("expected 1 but got %d", obj.FieldCount())
	}
}

func TestOpen(t *testing.T) {
	dir, err := ioutil2.TempDir("", "test")
	assertNil(t, err)

	fname := filepath.Join(dir, "mydb.bin")
	_ = os.Remove(fname)
	db, err := Open(fname)
	assertNil(t, err)
	err = db.Add(func(obj *Object) error {
		obj.AddField(1, func(f *FieldWriter) {
			f.WriteFloat(3)
		})
		return nil
	})
	assertNil(t, err)

	err = db.Close()
	assertNil(t, err)

	db, err = Open(fname)
	assertNil(t, err)

	is3 := false
	err = db.ForEach(func(id uint64, obj *Object) error {
		obj.WithFields(func(name uint32, kind ioutil.Type, f *FieldReader) {
			v := f.ReadFloat()
			if name == 1 && kind.IsNumber() && v == 3 {
				is3 = true
			}
		})
		return nil
	})
	assertNil(t, err)

	if !is3 {
		t.Fatalf("entry has been lost")
	}

	defer db.Close()
}

func assertNil(t interface {
	Helper()
	Fatalf(string, ...interface{})
}, i interface{}) {
	t.Helper()
	if i != nil {
		t.Fatalf("expected nil but got '%v'", i)
	}
}

func BenchmarkWrite(b *testing.B) {
	testSet := createTestTable(10)

	dir, err := ioutil2.TempDir("", "tableTest")
	assertNil(b, err)

	fmt.Println(dir)

	fname := filepath.Join(dir, "tabledb.bin")
	_ = os.Remove(fname)
	db, err := Open(fname)
	assertNil(b, err)

	b.Run("table", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, testObj := range testSet {
				err = db.Add(func(obj *Object) error {
					for _, field := range testObj.Fields {
						obj.AddField(field.Name, field.Write)
					}
					return nil
				})
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	err = db.Close()

}

func BenchmarkRead(b *testing.B) {
	testSet := createTestTable(defaultTableSize)

	dir, err := ioutil2.TempDir("", "tableTest")
	assertNil(b, err)

	fmt.Println(dir)

	fname := filepath.Join(dir, "tabledb.bin")
	_ = os.Remove(fname)
	db, err := Open(fname)
	assertNil(b, err)

	start := time.Now()
	for _, testObj := range testSet {
		err = db.Add(func(obj *Object) error {
			for _, field := range testObj.Fields {
				obj.AddField(field.Name, field.Write)
			}
			return nil
		})
		assertNil(b, err)
	}

	err = db.Close()
	assertNil(b, err)
	eps := float64(len(testSet)) / float64(time.Now().Sub(start)) * float64(time.Second)
	fmt.Printf("needed %v to insert %d entries (%2.f entries/second)\n", time.Now().Sub(start), len(testSet), eps)

	db, err = Open(fname)
	assertNil(b, err)
	defer db.Close()

	b.Run("table", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			objIdx := 0
			err = db.ForEach(func(id uint64, obj *Object) error {
				testObj := testSet[objIdx]

				fieldIdx := 0
				obj.WithFields(func(name uint32, kind ioutil.Type, f *FieldReader) {
					err = testObj.Fields[fieldIdx].Read(name, kind, f)
					if err != nil {
						b.Fatalf("failed at object %d at field %d: %v", objIdx, fieldIdx, err)
					}
					fieldIdx++
				})

				objIdx++
				return nil
			})

			if err != nil {
				b.Fatal(err)
			}
		}

	})

}

func TestTable(t *testing.T) {
	fields := []string{"", "A", "a", "jaksndkasjd\n#+\t"}

	testSet := createTestTable(defaultTableSize)

	dir, err := ioutil2.TempDir("", "tableTest")
	assertNil(t, err)

	fmt.Println(dir)

	fname := filepath.Join(dir, "tabledb.bin")
	_ = os.Remove(fname)
	db, err := Open(fname)
	assertNil(t, err)

	for _, field := range fields {
		if db.IndexByName(field) != -1 {
			t.Fatal()
		}
	}

	for _, field := range fields {
		db.PutName(field)
	}

	for i, field := range fields {
		if db.IndexByName(field) != i {
			t.Fatal()
		}
	}

	if db.PutName("A") != 1 {
		t.Fatal()
	}

	start := time.Now()
	for _, testObj := range testSet {
		err = db.Add(func(obj *Object) error {
			for _, field := range testObj.Fields {
				obj.AddField(field.Name, field.Write)
			}
			return nil
		})
		assertNil(t, err)
	}

	err = db.Close()
	assertNil(t, err)
	eps := float64(len(testSet)) / float64(time.Now().Sub(start)) * float64(time.Second)
	fmt.Printf("needed %v to insert %d entries (%2.f entries/second)\n", time.Now().Sub(start), len(testSet), eps)

	start = time.Now()
	db, err = Open(fname)
	assertNil(t, err)
	defer db.Close()

	for i, field := range fields {
		if db.IndexByName(field) != i {
			t.Fatal()
		}
	}

	tableOffsets := make([]uint64, len(testSet))
	tableCheck := make([]bool, len(testSet))
	objIdx := 0
	err = db.ForEach(func(id uint64, obj *Object) error {
		testObj := testSet[objIdx]

		fieldIdx := 0
		obj.WithFields(func(name uint32, kind ioutil.Type, f *FieldReader) {
			err = testObj.Fields[fieldIdx].Read(name, kind, f)
			if err != nil {
				t.Fatalf("failed at object %d at field %d: %v", objIdx, fieldIdx, err)
			}
			fieldIdx++
		})

		if fieldIdx != len(testObj.Fields) {
			t.Fatalf("lost fields for object %d, expected %d but got %d", objIdx, len(testObj.Fields), fieldIdx)
		}
		tableCheck[objIdx] = true
		tableOffsets[objIdx] = id

		objIdx++
		return nil
	})
	assertNil(t, err)

	for i, check := range tableCheck {
		if !check {
			t.Fatalf("object %d has not been visited", i)
		}
	}

	eps = float64(len(testSet)) / float64(time.Now().Sub(start)) * float64(time.Second)
	fmt.Printf("needed %v to read %d entries (%2.f entries/second)\n", time.Now().Sub(start), len(testSet), eps)

	// now, check single read commands
	start = time.Now()
	for objIdx, id := range tableOffsets {
		err := db.Read(id, func(obj *Object) error {
			testObj := testSet[objIdx]

			fieldIdx := 0
			obj.WithFields(func(name uint32, kind ioutil.Type, f *FieldReader) {
				err = testObj.Fields[fieldIdx].Read(name, kind, f)
				if err != nil {
					t.Fatalf("failed at object %d at field %d: %v", objIdx, fieldIdx, err)
				}
				fieldIdx++
			})

			if fieldIdx != len(testObj.Fields) {
				t.Fatalf("lost fields for object %d, expected %d but got %d", objIdx, len(testObj.Fields), fieldIdx)
			}
			return nil
		})
		assertNil(t, err)
	}
	eps = float64(len(testSet)) / float64(time.Now().Sub(start)) * float64(time.Second)
	fmt.Printf("needed %v to read single %d entries (%2.f entries/second)\n", time.Now().Sub(start), len(testSet), eps)

}

type TestObject struct {
	Fields []TestField
}

type TestField struct {
	Name  uint32
	Kind  ioutil.Type
	Value interface{}
}

var tmp64k = make([]byte, 1024*64)

func (t TestField) Read(name uint32, kind ioutil.Type, f *FieldReader) error {
	if t.Name != name {
		return fmt.Errorf("expected name %d but got %d", t.Name, name)
	}

	if t.Kind != kind {
		return fmt.Errorf("expected kind %d but got %d", t.Kind, kind)
	}

	switch t.Kind {
	case ioutil.TUint8:
		v := t.Value.(uint8)
		v1 := f.ReadUint8()
		if v != v1 {
			return fmt.Errorf("expected uint8 %v but got %v", v, v1)
		}
	case ioutil.TUint16:
		v := t.Value.(uint16)
		v1 := f.ReadUint16()
		if v != v1 {
			return fmt.Errorf("expected uint16 %v but got %v", v, v1)
		}
	case ioutil.TUint24:
		v := t.Value.(uint32)
		v1 := f.ReadUint24()
		if v != v1 {
			return fmt.Errorf("expected uint24 %v but got %v", v, v1)
		}
	case ioutil.TUint32:
		v := t.Value.(uint32)
		v1 := f.ReadUint32()
		if v != v1 {
			return fmt.Errorf("expected uint32 %v but got %v", v, v1)
		}
	case ioutil.TUint64:
		v := t.Value.(uint64)
		v1 := f.ReadUint64()
		if v != v1 {
			return fmt.Errorf("expected uint64 %v but got %v", v, v1)
		}
	case ioutil.TBlob8:
		v := t.Value.([]byte)

		lTmp := f.ReadBlob8(tmp64k)
		if len(v) != lTmp {
			return fmt.Errorf("expected tinyBlob length %d but got %d", len(v), lTmp)
		}

		if !bytes.Equal(v, tmp64k[:len(v)]) {
			return fmt.Errorf("expected tinyBlob equal \n%v but got \n%v", v, tmp64k[:len(v)])
		}
	case ioutil.TBlob16:
		v := t.Value.([]byte)
		lTmp := f.ReadBlob16(tmp64k)
		if len(v) != lTmp {
			return fmt.Errorf("expected Blob length %d but got %d", len(v), lTmp)
		}

		if !bytes.Equal(v, tmp64k[:len(v)]) {
			return fmt.Errorf("expected Blob equal %v but got %v", v, tmp64k[:len(v)])
		}
	case ioutil.TBlob24:
		v := t.Value.([]byte)
		lTmp := f.ReadBlob24(tmp64k)
		if len(v) != lTmp {
			return fmt.Errorf("expected mediumBlob length %d but got %d", len(v), lTmp)
		}

		if !bytes.Equal(v, tmp64k[:len(v)]) {
			return fmt.Errorf("expected mediumBlob equal %v but got %v", v, tmp64k[:len(v)])
		}
	case ioutil.TBlob32:
		v := t.Value.([]byte)
		lTmp := f.ReadBlob32(tmp64k)
		if len(v) != lTmp {
			return fmt.Errorf("expected longBlob length %d but got %d", len(v), lTmp)
		}

		if !bytes.Equal(v, tmp64k[:len(v)]) {
			return fmt.Errorf("expected longBlob equal \n%v but got \n%v", v, tmp64k[:len(v)])
		}
	case ioutil.TFloat32:
		v := t.Value.(float32)
		v1 := f.ReadFloat32()
		if v != v1 {
			return fmt.Errorf("expected float32 %v but got %v", v, v1)
		}
	case ioutil.TFloat64:
		v := t.Value.(float64)
		v1 := f.ReadFloat64()
		if v != v1 {
			return fmt.Errorf("expected float64 %v but got %v", v, v1)
		}
	default:
		panic("not implemented " + strconv.Itoa(int(t.Kind)))
	}

	return nil
}

func (t TestField) Write(f *FieldWriter) {
	switch t.Kind {
	case ioutil.TUint8:
		f.WriteUint8(t.Value.(uint8))
	case ioutil.TUint16:
		f.WriteUint16(t.Value.(uint16))
	case ioutil.TUint24:
		f.WriteUint24(t.Value.(uint32))
	case ioutil.TUint32:
		f.WriteUint32(t.Value.(uint32))
	case ioutil.TUint64:
		f.WriteUint64(t.Value.(uint64))
	case ioutil.TBlob8:
		f.WriteBlob8(t.Value.([]byte))
	case ioutil.TBlob16:
		f.WriteBlob16(t.Value.([]byte))
	case ioutil.TBlob24:
		f.WriteBlob24(t.Value.([]byte))
	case ioutil.TBlob32:
		f.WriteBlob32(t.Value.([]byte))
	case ioutil.TFloat32:
		f.WriteFloat32(t.Value.(float32))
	case ioutil.TFloat64:
		f.WriteFloat64(t.Value.(float64))
	default:
		panic("not implemented " + strconv.Itoa(int(t.Kind)))
	}
}

func createTestTable(max int) []TestObject {
	fmt.Println("creating table...")
	var r []TestObject
	for i := 0; i < max; i++ {
		myObj := TestObject{}
		fields := generateFields()
		for _, f := range fields {
			myObj.Fields = append(myObj.Fields, f)
		}
		r = append(r, myObj)
	}
	fmt.Println("table done")
	return r
}

func generateFields() []TestField {
	var r []TestField
	fieldCount := random.Intn(700)
	for i := 0; i < fieldCount; i++ {
		kind := randomKind()
		r = append(r, TestField{
			Name:  uint32(uint16(random.Uint32())),
			Kind:  kind,
			Value: generateValue(kind),
		})
	}
	return r
}

var testedTypes = []ioutil.Type{
	ioutil.TUint8, ioutil.TUint16, ioutil.TUint24, ioutil.TUint32, ioutil.TUint64, ioutil.TBlob8, ioutil.TBlob24, ioutil.TBlob16, ioutil.TBlob32, ioutil.TFloat64, ioutil.TFloat32,
}

func randomKind() ioutil.Type {

	n := random.Intn(len(testedTypes))
	return testedTypes[n]
}

var random = rand.New(rand.NewSource(1234))

func generateValue(kind ioutil.Type) interface{} {
	switch kind {
	case ioutil.TUint8:
		return uint8(random.Uint32())
	case ioutil.TUint16:
		return uint16(random.Uint32())
	case ioutil.TUint24:
		b := make([]byte, 3)
		random.Read(b)
		return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
	case ioutil.TUint32:
		return random.Uint32()
	case ioutil.TUint64:
		return random.Uint64()
	case ioutil.TBlob8:
		maxLen := byte(random.Uint32())
		tmp := make([]byte, maxLen)
		random.Read(tmp)
		return tmp
	case ioutil.TBlob24:
		fallthrough //TODO 64k max size limit?
	case ioutil.TBlob32:
		fallthrough //TODO 64k max size limit?
	case ioutil.TBlob16:
		maxLen := byte(random.Uint32()) * 2
		tmp := make([]byte, maxLen)
		random.Read(tmp)
		return tmp

	case ioutil.TFloat32:
		return random.Float32()
	case ioutil.TFloat64:
		return random.Float64()
	default:
		panic("not implemented " + strconv.Itoa(int(kind)))
	}
}
