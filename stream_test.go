package goetf

import (
	"bytes"
	"slices"
	"testing"
)

func TestStreamer(t *testing.T) {
	buf := make([]byte, 0)
	stream := newStreamer(bytes.NewBuffer(buf))

	var err error
	var n int
	{
		err = stream.writeByte(131)
		if err != nil {
			t.Fatal(err)
		}

		err = stream.writeByte(98)
		if err != nil {
			t.Fatal(err)
		}

		n, err = stream.write([]byte{0, 0, 1, 1})
		if err != nil {
			t.Fatal(err)
		}
		if n < 4 {
			t.Errorf("streamer shuld write")
		}

		got, err := stream.readAll()
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{131, 98, 0, 0, 1, 1}
		if !slices.Equal(want, got) {
			t.Errorf("streamer error: want = %v got = %v", want, got)
		}
	}
}
