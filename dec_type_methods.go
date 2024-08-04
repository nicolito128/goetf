package goetf

import (
	"encoding/binary"
	"math"
	"math/big"
	"reflect"
)

// parseStaticType parses a specific tag type from the input data
// and then returns the parsed value in the form of any.
func (d *Decoder) parseStaticType(kind reflect.Kind, tag ExternalTagType, data []byte) any {
	switch tag {
	case EttNil:
		return nil

	case EttString:
		return d.parseString(data)

	case EttAtom, EttAtomUTF8, EttSmallAtom, EttSmallAtomUTF8:
		s := d.parseString(data)
		switch {
		case s == "true":
			return true
		case s == "false":
			return false
		case s == "nil":
			return nil
		}

		return d.cache.Deduplicate(s)

	case EttSmallInteger:
		switch kind {
		default:
			return d.parseSmallInteger(data)
		case reflect.Int:
			return int(d.parseSmallInteger(data))
		case reflect.Int32:
			return int32(d.parseSmallInteger(data))
		case reflect.Int64:
			return int64(d.parseSmallInteger(data))
		case reflect.Uint:
			return uint(d.parseSmallInteger(data))
		case reflect.Uint16:
			return uint16(d.parseSmallInteger(data))
		case reflect.Uint32:
			return uint32(d.parseSmallInteger(data))
		case reflect.Uint64:
			return uint64(d.parseSmallInteger(data))
		}

	case EttInteger:
		switch kind {
		default:
			return d.parseInteger(data)
		case reflect.Int:
			return int(d.parseInteger(data))
		case reflect.Int64:
			return int64(d.parseInteger(data))
		}

	case EttNewFloat:
		switch kind {
		default:
			return d.parseNewFloat(data)
		case reflect.Float32:
			return float32(d.parseNewFloat(data))
		}

	case EttFloat:
		switch kind {
		default:
			return d.parseFloat(data)
		case reflect.Float32:
			return float32(d.parseFloat(data))
		}

	case EttSmallBig:
		switch kind {
		default:
			return d.parseSmallBig(data)
		case reflect.Int:
			return int(d.parseSmallBig(data))
		}

	case EttLargeBig:
		switch kind {
		default:
			return d.parseLargeBig(data)
		case reflect.Int, reflect.Int64:
			return d.parseLargeBig(data).Int64()
		}

	case EttBinary:
		switch kind {
		default:
			return data
		case reflect.String:
			return string(data)
		}

	case EttBitBinary:
		switch kind {
		default:
			return data
		case reflect.String:
			return string(data)
		}
	}

	return nil
}

func (d *Decoder) parseString(b []byte) string {
	return string(b)
}

func (d *Decoder) parseSmallInteger(b []byte) uint8 {
	return uint8(b[0])
}

func (d *Decoder) parseInteger(b []byte) int32 {
	return int32(binary.BigEndian.Uint32(b))
}

func (d *Decoder) parseNewFloat(b []byte) float64 {
	bits := binary.BigEndian.Uint64(b)
	float := math.Float64frombits(bits)
	return float
}

func (d *Decoder) parseFloat(b []byte) float64 {
	// todo: change to use string float IEEE format
	bits := binary.LittleEndian.Uint64(b)
	float := math.Float64frombits(bits)
	return float
}

func (d *Decoder) parseSmallBig(b []byte) int64 {
	sign := b[0]
	rest := b[1:]

	bits := binary.LittleEndian.Uint64(rest)
	smallBig := int64(bits)

	if sign == 1 {
		smallBig *= -1
	}

	return smallBig
}

func (d *Decoder) parseLargeBig(b []byte) *big.Int {
	sign := b[0]
	rest := b[1:]

	bits := binary.LittleEndian.Uint64(rest)
	largeBig := int64(bits)

	if sign == 1 {
		largeBig *= -1
	}

	return big.NewInt(largeBig)
}

// readStaticType reads a specific tag type from the underlying buffer,
// then returns the number of bytes read, a byte slice and an error, if any.
func (d *Decoder) readStaticType(tag ExternalTagType) (n int, b []byte, err error) {
	switch tag {
	default:
		n, b, err = 0, nil, errMalformed
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
		return n, bLen, errMalformedBitBinary
	}

	if n < SizeBitBinaryLen {
		return n, bLen, errMalformedBitBinary
	}

	length := int(binary.BigEndian.Uint32(bLen))

	_, err = d.scan.readByte()
	if err != nil {
		return n + 1, bLen, errMalformedBitBinary
	}

	n, data, err := d.scan.readN(length)
	if err != nil {
		return n, data, errMalformedBitBinary
	}

	return n, data, nil
}

func (d *Decoder) readAtomUTF8() (int, []byte, error) {
	n, bLen, err := d.scan.readN(SizeAtomUTF8)
	if err != nil {
		return n, bLen, errMalformedAtomUTF8
	}
	length := int(binary.BigEndian.Uint16(bLen))

	// {..., 118, 0, 0, ...}
	if length == 0 {
		return n, bLen, errMalformedAtomUTF8
	}

	n, data, err := d.scan.readN(length)
	if err != nil {
		return n, data, errMalformedAtomUTF8
	}

	return n, data, nil
}

func (d *Decoder) readSmallAtomUTF8() (int, []byte, error) {
	bLen, err := d.scan.readByte()
	if err != nil {
		return 1, nil, errMalformedSmallAtomUTF8
	}

	length := int(bLen)

	if length == 0 {
		return 1, nil, errMalformedSmallAtomUTF8
	}

	n, data, err := d.scan.readN(length)
	if err != nil {
		return n, data, errMalformedSmallAtomUTF8
	}

	return n, data, nil
}

func (d *Decoder) readLargeBig() (int, []byte, error) {
	n, bN, err := d.scan.readN(SizeLargeBigN)
	if err != nil {
		return n, bN, errMalformedLargeBig
	}

	if n < SizeLargeBigN {
		return n, bN, errMalformedLargeBig
	}
	N := int(binary.BigEndian.Uint32(bN))

	sign, err := d.scan.readByte()
	if err != nil {
		return n, nil, errMalformedLargeBig
	}

	n, data, err := d.scan.readN(N) // N+1 to store internaly the sign
	if err != nil {
		return n, data, errMalformedSmallBig
	}

	if N < 8 {
		N += 8 - N
	}

	largeBig := make([]byte, N+1)
	largeBig[0] = sign
	copy(largeBig[1:], data)

	return len(largeBig) - 1, largeBig, nil
}

func (d *Decoder) readSmallBig() (int, []byte, error) {
	bN, err := d.scan.readByte()
	if err != nil {
		return 1, nil, errMalformedSmallBig
	}
	// 'N' is the amount of bytes that are used for the small big
	N := int(bN)

	// positive or negative sign
	sign, err := d.scan.readByte()
	if err != nil {
		return 1, nil, errMalformedSmallBig
	}

	// fill with 0 to allow parsing
	n, data, err := d.scan.readN(N) // N+1 to store internaly the sign
	if err != nil {
		return n, data, errMalformedSmallBig
	}

	if N < 8 {
		N += 8 - N
	}

	smallBig := make([]byte, N+1)
	smallBig[0] = sign
	copy(smallBig[1:], data)

	return len(smallBig) - 1, smallBig, nil
}

func (d *Decoder) readBinary() (int, []byte, error) {
	n, bLen, err := d.scan.readN(SizeBinaryLen)
	if err != nil {
		return n, bLen, errMalformedBinary
	}
	length := int(binary.BigEndian.Uint32(bLen))

	n, binary, err := d.scan.readN(length)
	if err != nil {
		return n, binary, errMalformedBinary
	}

	if n < length {
		return n, binary, errMalformedBinary
	}

	return n, binary, nil
}

func (d *Decoder) readString() (int, []byte, error) {
	n, bLen, err := d.scan.readN(SizeStringLength)
	if err != nil {
		return n, bLen, errMalformedString
	}
	length := int(binary.BigEndian.Uint16(bLen))

	if length == 0 {
		return n, bLen, errMalformedString
	}

	n, bStr, err := d.scan.readN(length)
	if err != nil {
		return n, bStr, errMalformedString
	}

	return n, bStr, nil
}

func (d *Decoder) readSmallInteger() (int, []byte, error) {
	num, err := d.scan.readByte()
	if err != nil {
		return 1, []byte{num}, errMalformedSmallInteger
	}

	return 1, []byte{num}, nil
}

func (d *Decoder) readInteger() (int, []byte, error) {
	n, num, err := d.scan.readN(SizeInteger)
	if err != nil {
		return n, num, errMalformedInteger
	}

	return n, num, nil
}

func (d *Decoder) readNewFloat() (int, []byte, error) {
	n, num, err := d.scan.readN(SizeNewFloat)
	if err != nil {
		return n, num, errMalformedNewFloat
	}

	if n < (SizeNewFloat) {
		return n, num, errMalformedNewFloat
	}

	return n, num, nil
}

func (d *Decoder) readFloat() (int, []byte, error) {
	n, num, err := d.scan.readN(SizeFloat)
	if err != nil {
		return n, num, errMalformedFloat
	}

	return n, num, nil
}
