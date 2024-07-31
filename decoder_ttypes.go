package goetf

import (
	"bytes"
	"encoding/binary"
	"math"
)

// readType reads a specific tag type from the underlying buffer,
// then returns the number of bytes read, a byte slice and an error.
func (dec *Decoder) readType(typ ExternalTagType) (n int, b []byte, err error) {
	switch typ {
	case EttSmallInteger:
		n, b, err = dec.readSmallInteger()

	case EttInteger:
		n, b, err = dec.readInteger()

	case EttFloat:
		n, b, err = dec.readFloat()

	case EttNewFloat:
		n, b, err = dec.readNewFloat()

	case EttAtom, EttAtomUTF8:
		n, b, err = dec.readAtomUTF8()

	case EttString:
		n, b, err = dec.readString()

	case EttSmallBig:
		n, b, err = dec.readSmallBig()
	}

	return
}

func (dec *Decoder) readString() (int, []byte, error) {
	length := make([]byte, SizeStringLength)
	n, err := dec.rd.Read(length)
	if err != nil {
		return n, length, ErrMalformedString
	}

	return n, length, nil
}

func (dec *Decoder) readSmallBig() (int, []byte, error) {
	// 'n' is the amount of bytes that are used for the small big
	n, err := dec.rd.ReadByte()
	if err != nil {
		return 0, nil, ErrMalformedSmallBig
	}
	size := int(n)

	if dec.rd.Size() < (size + 1) {
		return 0, nil, ErrMalformedSmallBig
	}

	// positive or negative sign
	sign, err := dec.rd.ReadByte()
	if err != nil {
		return 0, nil, ErrMalformedSmallBig
	}

	// fill with 0 to allow parsing
	if size < 8 {
		size += 8 - size
	}

	num := make([]byte, size+1) // size+1 to store internaly the sign
	num[0] = sign               // positive or negative

	_, err = dec.rd.Read(num[1:])
	if err != nil {
		return 0, num, ErrMalformedSmallBig
	}

	return size, num, nil
}

func (dec *Decoder) readAtomUTF8() (int, []byte, error) {
	if dec.rd.Size() < 2 {
		return 0, nil, ErrMalformedAtomUTF8
	}

	b := make([]byte, SizeAtomUTF8)
	n, err := dec.rd.Read(b)
	if err != nil {
		return n, b, ErrMalformedAtomUTF8
	}

	size := int(binary.BigEndian.Uint16(b))
	// {..., 118, 0, 0, ...}
	if size == 0 {
		return n, b, ErrMalformedAtomUTF8
	}

	// {..., 118, 0, 1, 10, 11, 13}
	if dec.rd.Size() < size {
		return n, b, ErrMalformedAtomUTF8
	}

	b = make([]byte, size)
	n, err = dec.rd.Read(b)
	if err != nil {
		return n, b, ErrMalformedAtomUTF8
	}

	return n, b, nil
}

func (dec *Decoder) readSmallInteger() (int, []byte, error) {
	sb, err := dec.rd.ReadByte()
	if err != nil {
		return 0, nil, ErrMalformedSmallInteger
	}

	return 1, []byte{sb}, nil
}

func (dec *Decoder) readInteger() (int, []byte, error) {
	b := make([]byte, SizeInteger)
	n, err := dec.rd.Read(b)
	if err != nil {
		return n, b, ErrMalformedInteger
	}

	return n, b, nil
}

func (dec *Decoder) readNewFloat() (int, []byte, error) {
	if dec.rd.Size() < 8 {
		return 0, nil, ErrMalformedNewFloat
	}

	b := make([]byte, SizeNewFloat)
	n, err := dec.rd.Read(b)
	if err != nil {
		return n, b, ErrMalformedNewFloat
	}

	if n < int(SizeNewFloat) || n > int(SizeNewFloat) {
		return n, b, ErrMalformedNewFloat
	}

	return n, b, nil
}

func (dec *Decoder) readFloat() (int, []byte, error) {
	if dec.rd.Size() < 31 {
		return 0, nil, ErrMalformedFloat
	}

	b := make([]byte, SizeFloat)
	n, err := dec.rd.Read(b)
	if err != nil {
		return n, b, ErrMalformedFloat
	}

	return n, b, nil
}

func (dec *Decoder) parseType(flag ExternalTagType, data []byte) Term {
	switch flag {
	case EttAtom, EttAtomUTF8, EttString:
		return dec.parseString(data)

	case EttSmallInteger:
		return dec.parseSmallInteger(data)

	case EttInteger:
		return dec.parseInteger(data)

	case EttNewFloat:
		return dec.parseNewFloat(data)

	case EttFloat:
		return dec.parseFloat(data)

	case EttSmallBig:
		return dec.parseSmallBig(data)
	}

	return nil
}

func (dec *Decoder) parseString(b []byte) string {
	return string(b)
}

func (dec *Decoder) parseSmallInteger(b []byte) uint8 {
	return uint8(b[0])
}

func (dec *Decoder) parseInteger(b []byte) int32 {
	return int32(binary.BigEndian.Uint32(b))
}

func (dec *Decoder) parseNewFloat(b []byte) float64 {
	bits := binary.BigEndian.Uint64(b)
	float := math.Float64frombits(bits)
	return float
}

func (dec *Decoder) parseFloat(b []byte) float64 {
	// todo: change to use string float IEEE format
	bits := binary.LittleEndian.Uint64(b)
	float := math.Float64frombits(bits)
	return float
}

func (dec *Decoder) parseSmallBig(b []byte) int64 {
	sign := b[0]
	rest := bytes.Clone(b[1:])

	bits := binary.LittleEndian.Uint64(rest)
	smallBig := int64(bits)

	if sign == 1 {
		smallBig *= -1
	}

	return smallBig
}
