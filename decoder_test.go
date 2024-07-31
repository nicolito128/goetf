package goetf_test

import (
	"slices"
	"testing"

	"github.com/nicolito128/goetf"
)

func TestDecodeSmallInteger(t *testing.T) {
	var data uint8

	b := []byte{131, 97, 1}
	dec := goetf.NewDecoder(b)

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	var want uint8 = 1
	if want != data {
		t.Errorf("want = %v, got = %v", want, data)
	}

	// decoding using plain uint
	var data2 uint

	b = []byte{131, 97, 2}
	if err := dec.DecodePacket(b, &data2); err != nil {
		t.Fatal(err)
	}

	want2 := uint(2)
	if want2 != data2 {
		t.Errorf("want = %v, got = %v", want, data)
	}

}

func TestDecodeNewFloat(t *testing.T) {
	var data float64

	b := []byte{131, 70, 64, 9, 30, 184, 81, 235, 133, 31}
	dec := goetf.NewDecoder(b)

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	var want float64 = 3.14
	if want != data {
		t.Errorf("want = %v, got = %v", want, data)
	}

	// negative value parsing
	b = []byte{131, 70, 192, 4, 184, 81, 235, 133, 30, 184}

	if err := dec.DecodePacket(b, &data); err != nil {
		t.Fatal(err)
	}

	want = -2.59
	if want != data {
		t.Errorf("want = %v, got = %v", want, data)
	}
}

func TestDecodeAtomUTF8(t *testing.T) {
	var data goetf.Atom

	b := []byte{131, 118, 0, 3, 10, 11, 12}
	dec := goetf.NewDecoder(b)

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	var want goetf.Atom = "\n\v\f"
	if want != data {
		t.Errorf("want = %v, got = %v", want, data)
	}

	b2 := []byte{131, 118, 0, 0, 10} // malformed
	dec2 := goetf.NewDecoder(b2)

	if err := dec2.Decode(&data); err != goetf.ErrMalformedAtomUTF8 {
		t.Errorf("decode expected malformed atom UTF8 data")
	}
}

func TestDecodeString(t *testing.T) {
	var data string

	b := []byte{131, 107, 0, 12, 104, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100}
	dec := goetf.NewDecoder(b)

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	want := "hello, world"
	if want != data {
		t.Errorf("want = %v, got = %v", want, data)
	}
}

func TestDecodeInteger(t *testing.T) {
	var data int32

	b := []byte{131, 98, 0, 2, 0, 1}
	dec := goetf.NewDecoder(b)

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	var want int32 = 131073
	if want != data {
		t.Errorf("want = %v, got = %v", want, data)
	}
}

func TestDecodeSmallBig(t *testing.T) {
	var data int64

	b := []byte{131, 110, 5, 1, 199, 25, 70, 150, 3}
	dec := goetf.NewDecoder(b)

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	var want int64 = -15406078407
	if want != data {
		t.Errorf("want = %v, got = %v", want, data)
	}
}

func TestDecodeLargeBig(t *testing.T) {
	var data int64

	b := []byte{131, 111, 0, 0, 0, 6, 0, 100, 101, 97, 33, 75, 128}
	dec := goetf.NewDecoder(b)

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	var want int64 = 141060170933604
	if want != data {
		t.Errorf("want = %v, got = %v", want, data)
	}
}

func TestDecodeBinary(t *testing.T) {
	data := make([]byte, 0)

	b := []byte{131, 109, 0, 0, 0, 4, 101, 111, 112, 107}
	dec := goetf.NewDecoder(b)

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	want := []byte{101, 111, 112, 107}
	if n := slices.Compare(data, want); n != 0 {
		t.Errorf("want = %v, got = %v", want, data)
	}
}

func TestDecodeSmallTuple(t *testing.T) {
	data := make([]uint8, 2)

	b := []byte{131, 104, 2, 97, 1, 97, 2}
	dec := goetf.NewDecoder(b)

	if err := dec.Decode(data); err != nil {
		t.Fatal(err)
	}

	var want = []uint8{1, 2}
	if slices.Compare(want, data) != 0 {
		t.Errorf("want = %v, got = %v", want, data)
	}
}

func TestDecodeLargeTuple(t *testing.T) {
	data := make([]uint8, 2)

	b := []byte{131, 105, 0, 0, 0, 2, 97, 128, 97, 255}
	dec := goetf.NewDecoder(b)

	if err := dec.Decode(data); err != nil {
		t.Fatal(err)
	}

	var want = []uint8{128, 255}
	if slices.Compare(want, data) != 0 {
		t.Errorf("want = %v, got = %v", want, data)
	}
}

func TestDecodeList(t *testing.T) {
	data := []int32{0, 0, 0}

	b := []byte{131, 108, 0, 0, 0, 3, 98, 0, 0, 0, 1, 98, 0, 0, 0, 2, 98, 0, 0, 0, 3, 106}
	dec := goetf.NewDecoder(b)

	if err := dec.Decode(data); err != nil {
		t.Fatal(err)
	}

	var want = []int32{1, 2, 3}
	if slices.Compare(want, data) != 0 {
		t.Errorf("want = %v, got = %v", want, data)
	}
}
