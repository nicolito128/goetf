package goetf

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

type streamer struct {
	// buffer
	buf []byte
	// start of next data in buf
	written int64
	// writer buffer method
	w io.Writer
}

var streamerPool = sync.Pool{
	New: func() any {
		return &streamer{}
	},
}

func newStreamer(w io.Writer) *streamer {
	stream := streamerPool.Get().(*streamer)
	stream.written = 0
	if w != nil {
		stream.w = w
	} else {
		stream.buf = make([]byte, 0)
		stream.w = bytes.NewBuffer(stream.buf)
	}
	return stream
}

func (s *streamer) writeByte(b byte) error {
	_, err := s.write([]byte{b})
	return err
}

func (s *streamer) write(p []byte) (int, error) {
	n, err := s.w.Write(p)
	s.add(n)
	return n, err
}

func (s *streamer) readAll() ([]byte, error) {
	r, ok := s.w.(io.Reader)
	if !ok {
		return nil, fmt.Errorf("writer should implements the io.Reader interface")
	}
	return io.ReadAll(r)
}

func (s *streamer) add(steps int) {
	s.written += int64(steps)
}
