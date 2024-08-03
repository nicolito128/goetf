package goetf

import (
	"maps"
	"slices"
	"testing"
)

func TestDecodeSmallInteger(t *testing.T) {
	b := []byte{131, 97, 128}

	var out uint8
	if err := Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}

	want := uint8(128)
	if want != out {
		t.Errorf("unmarshal error: want = %v got = %v", want, out)
	}
}

func TestDecodeInteger(t *testing.T) {
	b := []byte{131, 98, 0, 2, 0, 1}

	var out int32
	if err := Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}

	var want int32 = 131073
	if want != out {
		t.Errorf("want = %v, got = %v", want, out)
	}
}

func TestDecodeSmallBig(t *testing.T) {
	b := []byte{131, 110, 5, 1, 199, 25, 70, 150, 2}

	var out int64
	if err := Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}

	var want int64 = -11111111111
	if want != out {
		t.Errorf("want = %v, got = %v", want, out)
	}
}

func TestDecodeLargeBig(t *testing.T) {
	b := []byte{131, 111, 0, 0, 0, 6, 0, 100, 101, 97, 33, 75, 128}

	var out int64
	if err := Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}

	var want int64 = 141060170933604
	if want != out {
		t.Errorf("want = %v, got = %v", want, out)
	}
}

func TestDecodeSmallAtom(t *testing.T) {
	{
		b := []byte{131, 119, 4, 116, 114, 117, 101}
		var out bool
		if err := Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		want := true
		if want != out {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{
		b := []byte{131, 119, 2, 111, 107}
		var out string
		if err := Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		want := "ok"
		if want != out {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
}

func TestDecodeString(t *testing.T) {
	b := []byte{131, 107, 0, 12, 104, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100}

	var out string
	if err := Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}

	want := "hello, world"
	if want != out {
		t.Errorf("unmarshal error: want = %v got = %v", want, out)
	}
}

func TestDecodeBinary(t *testing.T) {
	{
		b := []byte{131, 109, 0, 0, 0, 4, 101, 111, 112, 107}

		out := make([]byte, 0)
		if err := Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		want := []byte{101, 111, 112, 107}
		if n := slices.Compare(want, out); n != 0 {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{
		b := []byte{131, 109, 0, 0, 0, 14, 91, 34, 116, 101, 115, 116, 34, 93, 44, 123, 34, 116, 34, 125}

		out := make([]byte, 0)
		if err := Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		want := []byte{91, 34, 116, 101, 115, 116, 34, 93, 44, 123, 34, 116, 34, 125}
		if n := slices.Compare(want, out); n != 0 {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
}

func TestDecodeBitBinary(t *testing.T) {
	b := []byte{131, 77, 0, 0, 0, 3, 8, 128, 100, 99}

	out := make([]byte, 0)
	if err := Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}

	want := []byte{128, 100, 99}
	if n := slices.Compare(want, out); n != 0 {
		t.Errorf("unmarshal error: want = %v got = %v", want, out)
	}
}

func TestDecodeTuple(t *testing.T) {
	{ // small tuple
		b := []byte{131, 104, 5, 97, 1, 97, 2, 97, 3, 97, 4, 97, 5}

		out := make([]int, 0)
		if err := Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		want := []int{1, 2, 3, 4, 5}
		if !slices.Equal(want, out) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{ // large tuple
		b := []byte{131, 105, 0, 0, 0, 3, 97, 110, 97, 94, 98, 0, 0, 1, 1}

		out := make([]int, 0)
		if err := Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		want := []int{110, 94, 257}
		if !slices.Equal(want, out) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
}

func TestDecodeList(t *testing.T) {
	{
		b := []byte{131, 108, 0, 0, 0, 3, 97, 101, 97, 201, 97, 255, 106}

		var out [3]int
		if err := Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		want := [3]int{101, 201, 255}
		if !slices.Equal(want[:], out[:]) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{
		b := []byte{131, 108, 0, 0, 0, 2, 70, 64, 9, 33, 249, 240, 27, 134, 110, 70, 64, 5, 191, 9, 149, 170, 247, 144, 106}

		var out [2]float64
		if err := Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		want := [2]float64{3.14159, 2.71828}
		if !slices.Equal(want[:], out[:]) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
}

func TestDecodeMap(t *testing.T) {
	{
		b := []byte{
			131, 116, 0, 0, 0, 2,
			119, 4, 110, 97, 109, 101, // k1
			107, 0, 4, 74, 111, 104, 110, // v1
			119, 8, 112, 111, 115, 105, 116, 105, 111, 110, // k2
			107, 0, 8, 115, 121, 115, 97, 100, 109, 105, 110, // v2
		}

		out := map[string]string{}
		if err := Unmarshal(b, out); err != nil {
			t.Fatal(err)
		}

		want := map[string]string{"name": "John", "position": "sysadmin"}
		if !maps.Equal(want, out) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{
		b := []byte{131, 116, 0, 0, 0, 2, 119, 1, 97, 119, 4, 116, 114, 117, 101, 119, 1, 98, 70, 192, 9, 30, 184, 81, 235, 133, 31}

		out := map[string]any{}
		if err := Unmarshal(b, out); err != nil {
			t.Fatal(err)
		}

		want := map[string]any{"a": true, "b": -3.14}
		outa := out["a"].(bool)
		outb := out["b"].(float64)
		if want["a"] != outa || want["b"] != outb {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
}

func TestDecodeStruct(t *testing.T) {
	type user struct {
		Name string `etf:"name"`
		Age  uint8  `etf:"age"`
	}

	{
		b := []byte{
			131, 116, 0, 0, 0, 2,
			119, 4, 110, 97, 109, 101, 107, 0, 4, 77, 105, 108, 101,
			119, 3, 97, 103, 101, 97, 22,
		}

		var out user
		if err := Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		want := user{"Mile", 22}
		if want.Name != out.Name || want.Age != out.Age {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}

	type C struct {
		B string `etf:"b"`
		A []int  `etf:"a"`
	}

	{
		b := []byte{131, 116, 0, 0, 0, 2, 119, 1, 97, 104, 2, 97, 1, 97, 2, 119, 1, 98, 107, 0, 1, 98}

		var out C
		if err := Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		want := C{A: []int{1, 2}, B: "b"}
		if !slices.Equal(want.A, out.A) || want.B != out.B {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
}
