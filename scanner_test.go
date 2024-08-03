package goetf

import (
	"bytes"
	"io"
	"testing"
)

func TestScanner(t *testing.T) {
	b := []byte{131, 98, 0, 0, 1, 0}
	scan := newScanner(bytes.NewReader(b))

	{
		p, err := scan.readByte()
		if err != nil {
			t.Fatal(err)
		}

		want := byte(131)
		got := p
		if got != want {
			t.Errorf("error matching version flag: want = %v got = %v", want, got)
		}
	}
	{
		p, err := scan.readByte()
		if err != nil {
			t.Fatal(err)
		}

		want := byte(98)
		got := p
		if got != want {
			t.Errorf("error matching type tag: want = %v got = %v", want, got)
		}
	}
	{
		_, b, err := scan.readN(4)
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{0, 0, 1, 0}
		got := b
		if !bytes.Equal(want, got) {
			t.Errorf("error matching sequence of bytes: want = %v got = %v", want, got)
		}
	}
	{
		_, _, err := scan.readN(0)
		if err != io.EOF {
			t.Errorf("error should be EOF, but got: %v", err)
		}
	}
}
