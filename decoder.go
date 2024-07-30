package goetf

import (
	"bufio"
	"bytes"
	"io"

	"github.com/philpearl/intern"
)

// Decoder allows embedding ETF byte slices into valid Go code.
type Decoder[E Term] struct {
	rd        *bufio.Reader
	atomCache *intern.Intern
	// todo: mu        sync.Mutex
}

func NewDecoder[E Term](rd io.Reader) *Decoder[E] {
	return &Decoder[E]{
		rd:        bufio.NewReader(rd),
		atomCache: intern.New(2048),
	}
}

// Decode reads the next ETF-encoded value from its input and stores it in the value pointed to by v.
func (dec *Decoder[E]) Decode(v any) error {
	if v == nil {
		return ErrNilDecodeValue
	}

	return dec.decode(v)
}

// DecodePacket reads the next ETF-encoded packet and stores it in the value pointed to by v.
func (dec *Decoder[E]) DecodePacket(packet []byte, v any) error {
	if v == nil {
		return ErrNilDecodeValue
	}

	dec.rd = bufio.NewReader(bytes.NewReader(packet))
	return dec.decode(v)
}

func (dec *Decoder[E]) decode(v any) error {
	if dec.rd.Size() == 0 {
		return ErrMalformed
	}

	var err error
	var b byte

	b, err = dec.rd.ReadByte()
	if err != nil {
		return ErrMalformed
	}

	if b != EtVersion {
		return ErrMalformed
	}

	for {
		if dec.rd.Size() == 0 {
			return ErrMalformed
		}

		b, err = dec.rd.ReadByte()
		if err != nil && err != io.EOF {
			return ErrMalformed
		}

		if err == io.EOF {
			return nil
		}

		if !IsValidEtt(b) {
			return ErrMalformed
		}

		err = dec.decodeStatic(ExternalTagType(b), v)
		if err != nil {
			return err
		}
	}
}

func (dec *Decoder[E]) decodeStatic(b ExternalTagType, v any) error {
	switch b {
	case EttAtom, EttAtomUTF8:
		_, b, err := dec.readAtomUTF8()
		if err != nil {
			return err
		}

		parsed := dec.parseAtom(b)

		value := (v).(*E)
		(*value) = Term(parsed).(E)

	case EttSmallInteger:
		_, b, err := dec.readSmallInteger()
		if err != nil {
			return err
		}

		parsed := dec.parseSmallInteger(b)

		value := (v).(*E)
		(*value) = Term(parsed).(E)

	case EttInteger:
		_, b, err := dec.readInteger()
		if err != nil {
			return err
		}

		parsed := dec.parseInteger(b)

		value := (v).(*E)
		(*value) = Term(parsed).(E)

	case EttFloat:
		_, b, err := dec.readFloat()
		if err != nil {
			return err
		}

		parsed := dec.parseFloat(b)

		value := (v).(*E)
		(*value) = Term(parsed).(E)

	case EttNewFloat:
		_, b, err := dec.readNewFloat()
		if err != nil {
			return err
		}

		parsed := dec.parseNewFloat(b)

		value := (v).(*E)
		(*value) = Term(parsed).(E)

	case EttSmallBig:
		_, b, err := dec.readSmallBig()
		if err != nil {
			return err
		}

		parsed := dec.parseSmallBig(b)

		value := (v).(*E)
		(*value) = Term(parsed).(E)

	case EttSmallTuple:
		sb, err := dec.rd.ReadByte()
		if err != nil {
			return err
		}

		arity := int(sb)
		for i := 0; i < arity; i++ {
			br, err := dec.rd.ReadByte()
			if err != nil {
				return err
			}

			flag := ExternalTagType(br)
			dist := (v).(*[]E)
			_, b, err := dec.readType(flag)
			if err != nil {
				return err
			}

			(*dist)[i] = Term(dec.parseType(flag, b)).(E)
		}
	}

	return nil
}
