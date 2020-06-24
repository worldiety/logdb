package store

import (
	"fmt"
	"github.com/worldiety/ioutil"
	"io"
	ioutil2 "io/ioutil"
	"testing"
)

func TestParseHeader(t *testing.T) {
	args := []Header{
		{
			magic:      magic,
			version:    version,
			entries:    0,
			fieldCount: 0,
		},
		{
			magic:      magic,
			version:    version,
			entries:    1,
			fieldCount: 1,
			fieldNames: []string{"a"},
		},
		{
			magic:           magic,
			version:         version,
			entries:         1,
			lastTransaction: 1,
			fieldCount:      2,
			fieldNames:      []string{"a", "b"},
		},
		{
			magic:           magic,
			version:         version,
			entries:         2,
			lastTransaction: 3,
			fieldCount:      3,
			fieldNames:      []string{"a", "bbbb", "ccccc"},
		},
	}

	for num, arg := range args {
		fmt.Printf("test %d \n", num)
		f, err := ioutil2.TempFile("", "test")
		assertNil(t, err)

		dout := ioutil.NewDataOutput(ioutil.LittleEndian, f)
		din := ioutil.NewDataInput(ioutil.LittleEndian, f)

		assertNil(t, arg.Write(dout))

		// reset to start
		_, err = f.Seek(0, io.SeekStart)
		assertNil(t, err)

		header, err := ParseHeader(din)
		assertNil(t, err)

		if !header.Equals(arg) {
			t.Fatalf("expected %+v but got %+v", arg, header)
		}

		assertNil(t, f.Close())

	}

	header := NewHeader()
	if i := header.AddField("a"); i != 0 {
		t.Fatal(i)
	}

	if i := header.AddField("a"); i != 0 {
		t.Fatal(i)
	}

	if i := header.AddField("b"); i != 1 {
		t.Fatal(i)
	}
}

func assertNil(t *testing.T, i interface{}) {
	t.Helper()
	if i != nil {
		t.Fatalf("expected nil but got '%v'", i)
	}
}
