package goetf

import (
	"encoding/binary"
)

// readType reads a specific tag type from the underlying buffer,
// then returns the number of bytes read, a byte slice and an error, if any.
func (d *Decoder) readStaticType(tag ExternalTagType) (n int, b []byte, err error) {
	switch tag {
	default:
		n, b, err = 0, nil, ErrMalformed
	case EttNil:
		n, b, err = 0, []byte{EttNil}, nil
	case EttSmallInteger:
		n, b, err = d.readSmallInteger()
	case EttInteger:
		n, b, err = d.readInteger()
	case EttFloat:
		n, b, err = d.readFloat()
	case EttNewFloat:
		n, b, err = d.readNewFloat()
	case EttAtom, EttAtomUTF8:
		n, b, err = d.readAtomUTF8()
	case EttSmallAtom, EttSmallAtomUTF8:
		n, b, err = d.readSmallAtomUTF8()
	case EttString:
		n, b, err = d.readString()
	case EttSmallBig:
		n, b, err = d.readSmallBig()
	case EttLargeBig:
		n, b, err = d.readLargeBig()
	case EttBinary:
		n, b, err = d.readBinary()
	case EttBitBinary:
		n, b, err = d.readBitBinary()
	}

	return
}

func (d *Decoder) readBitBinary() (int, []byte, error) {
	n, bLen, err := d.scan.readN(SizeBitBinaryLen)
	if err != nil {
		return n, bLen, ErrMalformedBitBinary
	}

	if n < SizeBitBinaryLen {
		return n, bLen, ErrMalformedBitBinary
	}

	length := int(binary.BigEndian.Uint32(bLen))

	_, err = d.scan.readByte()
	if err != nil {
		return n + 1, bLen, ErrMalformedBitBinary
	}

	n, data, err := d.scan.readN(length)
	if err != nil {
		return n, data, ErrMalformedBitBinary
	}

	return n, data, nil
}

func (d *Decoder) readAtomUTF8() (int, []byte, error) {
	n, bLen, err := d.scan.readN(SizeAtomUTF8)
	if err != nil {
		return n, bLen, ErrMalformedAtomUTF8
	}
	length := int(binary.BigEndian.Uint16(bLen))

	// {..., 118, 0, 0, ...}
	if length == 0 {
		return n, bLen, ErrMalformedAtomUTF8
	}

	n, data, err := d.scan.readN(length)
	if err != nil {
		return n, data, ErrMalformedAtomUTF8
	}

	return n, data, nil
}

func (d *Decoder) readSmallAtomUTF8() (int, []byte, error) {
	bLen, err := d.scan.readByte()
	if err != nil {
		return 1, nil, ErrMalformedSmallAtomUTF8
	}

	length := int(bLen)

	if length == 0 {
		return 1, nil, ErrMalformedSmallAtomUTF8
	}

	n, data, err := d.scan.readN(length)
	if err != nil {
		return n, data, ErrMalformedSmallAtomUTF8
	}

	return n, data, nil
}

func (d *Decoder) readLargeBig() (int, []byte, error) {
	n, bN, err := d.scan.readN(SizeLargeBigN)
	if err != nil {
		return n, bN, ErrMalformedLargeBig
	}

	if n < SizeLargeBigN {
		return n, bN, ErrMalformedLargeBig
	}
	N := int(binary.BigEndian.Uint32(bN))

	// fill with 0 to allow parsing
	if N < 8 {
		N += 8 - N
	}

	sign, err := d.scan.readByte()
	if err != nil {
		return n, nil, ErrMalformedLargeBig
	}

	n, data, err := d.scan.readN(N + 1) // N+1 to store internaly the sign
	if err != nil {
		return n, data, ErrMalformedSmallBig
	}
	data[N] = sign // positive or negative at N

	return n, data, nil
}

func (d *Decoder) readSmallBig() (int, []byte, error) {
	bN, err := d.scan.readByte()
	if err != nil {
		return 1, nil, ErrMalformedSmallBig
	}
	// 'N' is the amount of bytes that are used for the small big
	N := int(bN)

	// positive or negative sign
	sign, err := d.scan.readByte()
	if err != nil {
		return 1, nil, ErrMalformedSmallBig
	}

	// fill with 0 to allow parsing
	if N < 8 {
		N += 8 - N
	}

	n, data, err := d.scan.readN(N + 1) // N+1 to store internaly the sign
	if err != nil {
		return n, data, ErrMalformedSmallBig
	}
	data[N] = sign // positivo or negative

	return N + 1, data, nil
}

func (d *Decoder) readBinary() (int, []byte, error) {
	n, bLen, err := d.scan.readN(SizeBinaryLen)
	if err != nil {
		return n, bLen, ErrMalformedBinary
	}
	length := int(binary.BigEndian.Uint32(bLen))

	n, binary, err := d.scan.readN(length)
	if err != nil {
		return n, binary, ErrMalformedBinary
	}

	if n < length {
		return n, binary, ErrMalformedBinary
	}

	return n, binary, nil
}

func (d *Decoder) readString() (int, []byte, error) {
	n, bLen, err := d.scan.readN(SizeStringLength)
	if err != nil {
		return n, bLen, ErrMalformedString
	}
	length := int(binary.BigEndian.Uint16(bLen))

	if length == 0 {
		return n, bLen, ErrMalformedString
	}

	n, bStr, err := d.scan.readN(length)
	if err != nil {
		return n, bStr, ErrMalformedString
	}

	return n, bStr, nil
}

func (d *Decoder) readSmallInteger() (int, []byte, error) {
	num, err := d.scan.readByte()
	if err != nil {
		return 1, []byte{num}, ErrMalformedSmallInteger
	}

	return 1, []byte{num}, nil
}

func (d *Decoder) readInteger() (int, []byte, error) {
	n, num, err := d.scan.readN(SizeInteger)
	if err != nil {
		return n, num, ErrMalformedInteger
	}

	return n, num, nil
}

func (d *Decoder) readNewFloat() (int, []byte, error) {
	n, num, err := d.scan.readN(SizeNewFloat)
	if err != nil {
		return n, num, ErrMalformedNewFloat
	}

	if n < (SizeNewFloat) {
		return n, num, ErrMalformedNewFloat
	}

	return n, num, nil
}

func (d *Decoder) readFloat() (int, []byte, error) {
	n, num, err := d.scan.readN(SizeFloat)
	if err != nil {
		return n, num, ErrMalformedFloat
	}

	return n, num, nil
}
