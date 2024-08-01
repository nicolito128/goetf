package goetf

import (
	"encoding/binary"
	"math"
	"reflect"
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

	case EttSmallAtom, EttSmallAtomUTF8:
		n, b, err = dec.readSmallAtomUTF8()

	case EttString:
		n, b, err = dec.readString()

	case EttSmallBig:
		n, b, err = dec.readSmallBig()

	case EttBinary:
		n, b, err = dec.readBinary()

	case EttLargeBig:
		n, b, err = dec.readLargeBig()

	case EttBitBinary:
		n, b, err = dec.readBitBinary()
	}

	return
}

func (dec *Decoder) readBitBinary() (int, []byte, error) {
	blength := make([]byte, SizeBitBinaryLen)
	n, err := dec.rd.Read(blength)
	if err != nil {
		return n, blength, ErrMalformedBitBinary
	}

	if n < int(SizeBitBinaryLen) || dec.rd.Size() < (n+1) {
		return n, blength, ErrMalformedBitBinary
	}

	length := int(binary.BigEndian.Uint32(blength))

	_, err = dec.rd.ReadByte()
	if err != nil {
		return n + 1, blength, ErrMalformedBitBinary
	}

	b := make([]byte, length)
	n, err = dec.rd.Read(b)
	if err != nil {
		return n, b, ErrMalformedBitBinary
	}

	return n, b, nil
}

func (dec *Decoder) readSmallAtomUTF8() (int, []byte, error) {
	sb, err := dec.rd.ReadByte()
	if err != nil {
		return 1, nil, ErrMalformedSmallAtomUTF8
	}

	length := int(sb)
	if length == 0 || dec.rd.Size() < length {
		return 1, nil, ErrMalformedSmallAtomUTF8
	}

	b := make([]byte, length)
	n, err := dec.rd.Read(b)
	if err != nil {
		return n, b, ErrMalformedSmallAtomUTF8
	}

	return n, b, nil
}

func (dec *Decoder) readLargeBig() (int, []byte, error) {
	bn := make([]byte, SizeLargeBigN)
	n, err := dec.rd.Read(bn)
	if err != nil {
		return n, bn, ErrMalformedLargeBig
	}

	if n < int(SizeLargeBigN) || dec.rd.Size() < (n+1) {
		return n, bn, ErrMalformedLargeBig
	}

	size := int(dec.parseInteger(bn))

	// fill with 0 to allow parsing
	if size < 8 {
		size += 8 - size
	}

	sign, err := dec.rd.ReadByte()
	if err != nil {
		return n, nil, ErrMalformedLargeBig
	}

	num := make([]byte, size+1) // size+1 to store internaly the sign
	num[0] = sign               // positive or negative

	_, err = dec.rd.Read(num[1:])
	if err != nil {
		return 0, num, ErrMalformedSmallBig
	}

	return n, num, nil
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

func (dec *Decoder) readBinary() (int, []byte, error) {
	if dec.rd.Size() < 4 {
		return 0, nil, ErrMalformedBinary
	}

	blength := make([]byte, SizeBinaryLen)
	n, err := dec.rd.Read(blength)
	if err != nil {
		return n, blength, ErrMalformedBinary
	}

	length := int(dec.parseInteger(blength))

	binary := make([]byte, length)
	n, err = dec.rd.Read(binary)
	if err != nil {
		return n, binary, ErrMalformedBinary
	}

	if n < length {
		return n, binary, ErrMalformedBinary
	}

	return n, binary, nil
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

func (dec *Decoder) readString() (int, []byte, error) {
	if dec.rd.Size() < 2 {
		return 0, nil, ErrMalformedString
	}

	blength := make([]byte, 2)
	n, err := dec.rd.Read(blength)
	if err != nil {
		return n, blength, ErrMalformedString
	}

	length := int(binary.BigEndian.Uint16(blength))
	bstr := make([]byte, length)
	n, err = dec.rd.Read(bstr)
	if err != nil {
		return n, bstr, ErrMalformedString
	}

	if n < length {
		return n, bstr, ErrMalformedString
	}

	return n, bstr, nil
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

func (dec *Decoder) decodeVariadicType(src reflect.Value, flag ExternalTagType) error {
	switch flag {
	case EttSmallTuple:
		sb, err := dec.rd.ReadByte()
		if err != nil {
			return err
		}

		arity := int(sb)

		for i := 0; i < arity; i++ {
			bflag, err := dec.rd.ReadByte()
			if err != nil {
				return ErrMalformedSmallTuple
			}
			flag := ExternalTagType(bflag)

			_, b, err := dec.readType(flag)
			if err != nil {
				return ErrMalformedSmallTuple
			}

			elem := src.Index(i)
			switch elem.Type().Kind() {
			case reflect.Slice, reflect.Array:
				if err := dec.decodeVariadicType(elem, EttSmallTuple); err != nil {
					return err
				}
			default:
				item := dec.parseType(reflect.ValueOf(nil), flag, b)
				elem.Set(reflect.ValueOf(item))
			}
		}

	case EttLargeTuple:
		b := make([]byte, SizeLargeTupleArity)
		_, err := dec.rd.Read(b)
		if err != nil {
			return ErrMalformedLargeTuple
		}

		arity := int(dec.parseInteger(b))

		for i := 0; i < arity; i++ {
			bflag, err := dec.rd.ReadByte()
			if err != nil {
				return ErrMalformedLargeTuple
			}
			flag := ExternalTagType(bflag)

			_, b, err := dec.readType(flag)
			if err != nil {
				return ErrMalformedLargeTuple
			}

			elem := src.Index(i)
			switch elem.Type().Kind() {
			case reflect.Slice, reflect.Array, reflect.Map:
				dec.decodeVariadicType(elem, EttLargeTuple)
			default:
				item := dec.parseType(reflect.ValueOf(nil), flag, b)
				elem.Set(reflect.ValueOf(item))
			}
		}

	case EttList:
		b := make([]byte, SizeListLength)
		_, err := dec.rd.Read(b)
		if err != nil {
			return ErrMalformedList
		}

		length := int(dec.parseInteger(b))

		for i := 0; i < length; i++ {
			bflag, err := dec.rd.ReadByte()
			if err != nil {
				return ErrMalformedList
			}
			flag := ExternalTagType(bflag)

			_, b, err := dec.readType(flag)
			if err != nil {
				return ErrMalformedList
			}

			elem := src.Index(i)
			switch elem.Type().Kind() {
			case reflect.Slice, reflect.Array, reflect.Map:
				if err := dec.decodeVariadicType(elem, EttList); err != nil {
					return err
				}
			default:
				item := dec.parseType(reflect.ValueOf(nil), flag, b)
				elem.Set(reflect.ValueOf(item))
			}
		}

		// read the 106 tail byte ([])
		dec.rd.ReadByte()

	case EttMap:
		bsize := make([]byte, SizeMapArity)
		n, err := dec.rd.Read(bsize)
		if err != nil {
			return ErrMalformedMap
		}

		if n < 4 {
			return ErrMalformedMap
		}

		arity := int(dec.parseInteger(bsize))

		for i := 0; i < arity; i++ {
			// Key
			bflag, err := dec.rd.ReadByte()
			if err != nil {
				return ErrMalformedMap
			}
			keyFlag := ExternalTagType(bflag)

			_, b, err := dec.readType(keyFlag)
			if err != nil {
				return ErrMalformedMap
			}
			key := dec.parseType(src, keyFlag, b)

			// Value
			bflag, err = dec.rd.ReadByte()
			if err != nil {
				return ErrMalformedMap
			}
			valueFlag := ExternalTagType(bflag)

			_, b, err = dec.readType(valueFlag)
			if err != nil {
				return ErrMalformedMap
			}
			value := dec.parseType(src, valueFlag, b)

			switch v := value.(type) {
			case reflect.Value:
				src.SetMapIndex(reflect.ValueOf(key), v)
			default:
				src.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
			}
		}
	}

	return nil
}

func (dec *Decoder) parseType(src reflect.Value, tag ExternalTagType, data []byte) Term {
	switch tag {
	case EttAtom, EttAtomUTF8, EttString:
		return dec.parseString(data)

	case EttSmallAtom, EttSmallAtomUTF8:
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

	case EttLargeBig:
		return dec.parseLargeBig(data)

	case EttBinary:
		return dec.parseBinary(data)

	case EttBitBinary:
		return dec.parseBinary(data)

	case EttMap:
		switch src.Type().Kind() {
		case reflect.Map:
			bsize := make([]byte, SizeMapArity)
			n, err := dec.rd.Read(bsize)
			if err != nil {
				return ErrMalformedMap
			}

			if n < 4 {
				return ErrMalformedMap
			}

			arity := int(dec.parseInteger(bsize))

			var newMap reflect.Value
			for i := 0; i < arity; i++ {
				// Key
				bflag, err := dec.rd.ReadByte()
				if err != nil {
					return ErrMalformedMap
				}
				keyFlag := ExternalTagType(bflag)

				_, b, err := dec.readType(keyFlag)
				if err != nil {
					return ErrMalformedMap
				}
				key := dec.parseType(src, keyFlag, b)

				// Value
				bflag, err = dec.rd.ReadByte()
				if err != nil {
					return ErrMalformedMap
				}
				valueFlag := ExternalTagType(bflag)

				_, b, err = dec.readType(valueFlag)
				if err != nil {
					return ErrMalformedMap
				}
				value := dec.parseType(src, valueFlag, b)

				newMap = reflect.MakeMap(src.Type().Elem())
				newMap.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
			}

			return newMap
		}
	}

	return nil
}

func (dec *Decoder) parseBinary(b []byte) []byte {
	return b
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
	rest := b[1:]

	bits := binary.LittleEndian.Uint64(rest)
	smallBig := int64(bits)

	if sign == 1 {
		smallBig *= -1
	}

	return smallBig
}

func (dec *Decoder) parseLargeBig(b []byte) int64 {
	sign := b[0]
	rest := b[1:]

	bits := binary.LittleEndian.Uint64(rest)
	largeBig := int64(bits)

	if sign == 1 {
		largeBig *= -1
	}

	return largeBig
}
