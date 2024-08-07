package goetf_test

import (
	"bytes"
	"maps"
	"math"
	"math/big"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/nicolito128/goetf"
)

func TestEncodeSmallInteger(t *testing.T) {
	var data uint8 = 255

	got, err := goetf.Marshal(data)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	want := []byte{131, 97, 255}
	if !slices.Equal(want, got) {
		t.Errorf("encode error: want = %v got = %v", want, got)
	}
}

func TestEncodeInteger(t *testing.T) {
	{
		var data uint16 = math.MaxUint16

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 98, 0, 0, 255, 255}
		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
	{
		var data int16 = math.MaxInt16

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 98, 0, 0, 127, 255}
		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
	{
		var data int32 = math.MaxInt32

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 98, 127, 255, 255, 255}
		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
}

func TestEncodeBig(t *testing.T) {
	{
		var data int64 = 314159265359

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 110, 8, 0, 79, 246, 89, 37, 73, 0, 0, 0}
		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
	{
		data := big.NewInt(16777216)

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 111, 0, 0, 0, 4, 0, 0, 0, 0, 1}
		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}

	}
}

func TestEncodeFloat(t *testing.T) {
	{
		var data float64 = 3.14159265359

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 70, 64, 9, 33, 251, 84, 68, 46, 234}
		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
	{
		var data float32 = 340.28234663852885

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 70, 64, 117, 68, 132, 128, 0, 0, 0}
		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
}

func TestEncodeString(t *testing.T) {
	{
		data := "hello, etf world"

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 107, 0, 16, 104, 101, 108, 108, 111, 44, 32, 101, 116, 102, 32,
			119, 111, 114, 108, 100}
		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
	{
		data := []byte{'p', 'h', 'o', 'n', 'e', ' ', 'n', 'u', 'm', 'b', 'e', 'e', 'r'}

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 77, 0, 0, 0, 13, 8, 112, 104, 111, 110, 101, 32, 110, 117, 109, 98, 101, 101, 114}

		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
	{
		var data strings.Builder
		data.Grow(256)
		for range 256 {
			_, err := data.WriteString("n")
			if err != nil {
				t.Fatal(err)
			}
		}

		got, err := goetf.Marshal(data.String())
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 107, 1, 0, 110}
		for range 255 {
			want = append(want, 110)
		}

		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
}

func TestEncodeBool(t *testing.T) {
	{
		data := true

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 119, 4, 116, 114, 117, 101}
		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
	{
		data := false

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 119, 5, 102, 97, 108, 115, 101}
		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
}

func TestEncodeNil(t *testing.T) {
	var data *int

	got, err := goetf.Marshal(data)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	want := []byte{131, 119, 3, 110, 105, 108}
	if !slices.Equal(want, got) {
		t.Errorf("encode error: want = %v got = %v", want, got)
	}
}

func TestEncodeTuples(t *testing.T) {
	{
		data := []int{1, 2, 3, 4, 5}

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 104, 5, 98, 0, 0, 0, 1, 98, 0, 0, 0, 2, 98, 0, 0, 0, 3, 98, 0, 0, 0, 4, 98, 0, 0, 0, 5}
		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
	{
		data := [][]int{{1, 2}, {3, 4}, {5, 6}}

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		want := []byte{131, 104, 3, 104, 2, 98, 0, 0, 0, 1, 98, 0, 0, 0, 2, 104, 2, 98, 0, 0, 0, 3, 98, 0, 0, 0, 4, 104, 2, 98, 0, 0, 0, 5, 98, 0, 0, 0, 6}
		if !slices.Equal(want, got) {
			t.Errorf("encode error: want = %v got = %v", want, got)
		}
	}
}

func TestEncodeList(t *testing.T) {
	data := [3]string{"a", "b", "c"}

	got, err := goetf.Marshal(data)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	want := []byte{131, 108, 0, 0, 0, 3, 119, 1, 97, 119, 1, 98, 119, 1, 99, 106}
	if !slices.Equal(want, got) {
		t.Errorf("encode error: want = %v got = %v", want, got)
	}
}

func TestEncodeMap(t *testing.T) {
	want := map[string]int{"a": 97, "b": 98, "c": 99}

	got, err := goetf.Marshal(want)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	out := map[string]int{}
	if err := goetf.Unmarshal(got, out); err != nil {
		t.Fatal("unmarshal error:", err)
	}

	if !maps.Equal(out, want) {
		t.Errorf("encode error: want = %v got = %v", want, out)
	}
}

func TestEncodeStruct(t *testing.T) {
	type Profile struct {
		Active    bool  `etf:"active"`
		LastLogin int64 `etf:"last_login"`
	}

	type User struct {
		Name string  `etf:"name"`
		Age  uint8   `etf:"age"`
		P    Profile `etf:"profile"`
	}

	data := User{"John", 37, Profile{true, 24000}}
	got, err := goetf.Marshal(data)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	var out User
	if err := goetf.Unmarshal(got, &out); err != nil {
		t.Fatal("unmarshal error:", err)
	}

	if !reflect.ValueOf(out).Equal(reflect.ValueOf(data)) {
		t.Errorf("encode error: want = %v got = %v", data, out)
	}
}

func TestEncodeStructWithNil(t *testing.T) {
	type Account struct {
		Username string  `etf:"name"`
		Status   int     `etf:"status"`
		Items    *[]int  `etf:"items"`
		Thing    *string `etf:"thing"`
	}

	data := Account{"username", 15, nil, nil}
	got, err := goetf.Marshal(data)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	var out Account
	if err := goetf.Unmarshal(got, &out); err != nil {
		t.Fatal("unmarshal error:", err)
	}

	want := Account{Username: "username", Status: 15, Items: nil, Thing: nil}
	if want.Username != out.Username || want.Status != out.Status || out.Items != nil || out.Thing != nil {
		t.Errorf("encode error: want = %v got = %v", want, out)
	}
}

func TestEncodeOptions(t *testing.T) {
	data := map[string]int{"1": 1, "2": 2}

	buf := bytes.NewBuffer(make([]byte, 0))
	eng := goetf.NewEncoder(buf, goetf.WithStringOverAtom)

	if err := eng.Encode(data); err != nil {
		t.Fatal(err)
	}

	got, err := eng.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	want := []byte{131, 116, 0, 0, 0, 2, 107, 0, 1, 49, 98, 0, 0, 0, 1, 107, 0, 1, 50, 98, 0, 0, 0, 2}
	if !slices.Equal(want, got) {
		t.Errorf("encode error: want = %v got = %v", want, got)
	}
}
