package goetf

import (
	"bytes"
	"slices"
	"testing"
)

func TestDecodeSmallInteger(t *testing.T) {
	var data uint8

	b := []byte{131, 97, 1}
	dec := NewDecoder[uint8](bytes.NewReader(b))

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	var want uint8 = 1
	if want != data {
		t.Errorf("want = %v, got = %v", want, data)
	}
}

func TestDecodeFloat(t *testing.T) {
	var data float64

	b := []byte{131, 70, 64, 9, 30, 184, 81, 235, 133, 31}
	dec := NewDecoder[float64](bytes.NewReader(b))

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	var want float64 = 3.14
	if want != data {
		t.Errorf("want = %v, got = %v", want, data)
	}
}

func TestDecodeAtomUTF8(t *testing.T) {
	var data Atom

	b := []byte{131, 118, 0, 3, 10, 11, 12}
	dec := NewDecoder[Atom](bytes.NewReader(b))

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	var want Atom = "\n\v\f"
	if want != data {
		t.Errorf("want = %v, got = %v", want, data)
	}

	b2 := []byte{131, 118, 0, 0, 10} // malformed
	dec2 := NewDecoder[Atom](bytes.NewReader(b2))

	if err := dec2.Decode(&data); err != ErrMalformedAtomUTF8 {
		t.Errorf("decode expected malformed atom UTF8 data")
	}
}

func TestDecodeInteger(t *testing.T) {
	var data int32

	b := []byte{131, 98, 0, 2, 0, 1}
	dec := NewDecoder[int32](bytes.NewReader(b))

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

	b := []byte{131, 110, 5, 1, 199, 25, 70, 150, 2}
	dec := NewDecoder[int64](bytes.NewReader(b))

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	var want int64 = -11111111111
	if want != data {
		t.Errorf("want = %v, got = %v", want, data)
	}
}

func TestDecodeSmallTuple(t *testing.T) {
	data := make([]uint8, 2)

	b := []byte{131, 104, 2, 97, 1, 97, 2}
	dec := NewDecoder[uint8](bytes.NewReader(b))

	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}

	var want = []uint8{1, 2}
	if slices.Compare(want, data) != 0 {
		t.Errorf("want = %v, got = %v", want, data)
	}
}
