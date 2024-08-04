package goetf

import (
	"fmt"
	"io"
	"sync"
)

type scanner struct {
	// Buffer where readings are stored. Default size is 2048.
	buf []byte
	// Start of unread data in buf.
	scanp int
	// Total bytes consumed.
	scanned int64

	r io.Reader
}

var scannerPool = sync.Pool{
	New: func() any {
		return &scanner{}
	},
}

func newScanner(r io.Reader) *scanner {
	scan := scannerPool.Get().(*scanner)
	scan.scanned = 0
	scan.r = r
	scan.buf = make([]byte, 4096)
	return scan
}

func (s *scanner) reset(bufinit int) {
	newLen := bufinit + len(s.buf)*2
	s.buf = make([]byte, newLen)
	s.scanp = 0
}

func (s *scanner) readByte() (byte, error) {
	_, b, err := s.readN(1)
	return b[0], err
}

func (s *scanner) readN(n int) (int, []byte, error) {
	if s.buf == nil {
		return 0, nil, fmt.Errorf("scanner readN error: invalid buffer")
	}

	if (s.scanp + n) >= len(s.buf) {
		s.reset(n)
	}

	bytes, err := s.r.Read(s.buf[s.scanp : s.scanp+n])
	s.forward(bytes)
	if err != nil {
		return bytes, s.buf, err
	}

	return bytes, s.buf[s.scanp-bytes : s.scanp], nil
}

func (s *scanner) forward(steps int) {
	s.scanp += steps
	s.scanned += int64(steps)
}

func (s *scanner) eof() bool {
	_, _, err := s.readN(0)
	return err == io.EOF
}
