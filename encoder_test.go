package goetf_test

import (
	"math"
	"math/big"
	"slices"
	"strings"
	"testing"

	"github.com/nicolito128/goetf"
)

func TestEncodeSmallInteger(t *testing.T) {
	var data uint8 = 255

	got, err := goetf.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	want := []byte{131, 97, 255}
	if !slices.Equal(want, got) {
		t.Errorf("marshal error: want = %v got = %v", want, got)
	}
}

func TestEncodeInteger(t *testing.T) {
	{
		var data uint16 = math.MaxUint16

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 98, 0, 0, 255, 255}
		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}
	}
	{
		var data int16 = math.MaxInt16

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 98, 0, 0, 127, 255}
		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}
	}
	{
		var data int32 = math.MaxInt32

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 98, 127, 255, 255, 255}
		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}
	}
}

func TestEncodeBig(t *testing.T) {
	{
		var data int64 = 314159265359

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 110, 8, 0, 79, 246, 89, 37, 73, 0, 0, 0}
		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}
	}
	{
		data := big.NewInt(16777216)

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 111, 0, 0, 0, 4, 0, 0, 0, 0, 1}
		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}

	}
}

func TestEncodeFloat(t *testing.T) {
	{
		var data float64 = 3.14159265359

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 70, 64, 9, 33, 251, 84, 68, 46, 234}
		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}
	}
	{
		var data float32 = 340.28234663852885

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 70, 64, 117, 68, 132, 128, 0, 0, 0}
		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}
	}
}

func TestEncodeString(t *testing.T) {
	{
		data := "hello, etf world"

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 107, 0, 16, 104, 101, 108, 108, 111, 44, 32, 101, 116, 102, 32,
			119, 111, 114, 108, 100}
		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}
	}
	{
		data := []byte{'p', 'h', 'o', 'n', 'e', ' ', 'n', 'u', 'm', 'b', 'e', 'e', 'r'}

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 77, 0, 0, 0, 13, 8, 112, 104, 111, 110, 101, 32, 110, 117, 109, 98, 101, 101, 114}

		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
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
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}
	}
}

func TestEncodeBool(t *testing.T) {
	{
		data := true

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 119, 4, 116, 114, 117, 101}
		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}
	}
	{
		data := false

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 119, 5, 102, 97, 108, 115, 101}
		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}
	}
}

func TestEncodeNil(t *testing.T) {
	var data *int

	got, err := goetf.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	want := []byte{131, 119, 3, 110, 105, 108}
	if !slices.Equal(want, got) {
		t.Errorf("marshal error: want = %v got = %v", want, got)
	}
}

func TestEncodeTuples(t *testing.T) {
	{
		data := []int{1, 2, 3, 4, 5}

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 104, 5, 97, 1, 97, 2, 97, 3, 97, 4, 97, 5}
		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}
	}
	{
		data := [][]int{{1, 2}, {3, 4}, {5, 6}}

		got, err := goetf.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 104, 3, 104, 2, 97, 1, 97, 2, 104, 2, 97, 3, 97, 4, 104, 2, 97, 5, 97, 6}
		if !slices.Equal(want, got) {
			t.Errorf("marshal error: want = %v got = %v", want, got)
		}
	}
}

func TestEncodeList(t *testing.T) {
	data := [3]string{"a", "b", "c"}

	got, err := goetf.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	want := []byte{131, 108, 0, 0, 0, 3, 119, 1, 97, 119, 1, 98, 119, 1, 99, 106}
	if !slices.Equal(want, got) {
		t.Errorf("marshal error: want = %v got = %v", want, got)
	}
}
