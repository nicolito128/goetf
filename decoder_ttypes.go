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

	case EttLargeBig:
		n, b, err = dec.readLargeBig()

	case EttBinary:
		n, b, err = dec.readBinary()

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

func (dec *Decoder) decodeVariadicType(src reflect.Value, ettflag ExternalTagType) error {
	switch ettflag {
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

			switch src.Type().Kind() {
			case reflect.Slice, reflect.Array:
				if src.Len() > 0 {
					elem := src.Index(i)
					switch elem.Type().Kind() {
					case reflect.Slice, reflect.Array, reflect.Map:
						if err := dec.decodeVariadicType(elem, flag); err != nil {
							return err
						}

					default:
						item := dec.parseType(src, flag, b)
						elem.Set(reflect.ValueOf(item))
					}
				}

			case reflect.Map:
				if err := dec.decodeVariadicType(src, flag); err != nil {
					return err
				}
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
			fg := ExternalTagType(bflag)

			_, b, err := dec.readType(fg)
			if err != nil {
				return ErrMalformedLargeTuple
			}

			switch src.Type().Kind() {
			case reflect.Slice, reflect.Array:
				if src.Len() > 0 {
					elem := src.Index(i)
					switch elem.Type().Kind() {
					case reflect.Slice, reflect.Array, reflect.Map:
						if err := dec.decodeVariadicType(elem, fg); err != nil {
							return err
						}

					default:
						item := dec.parseType(src, fg, b)
						elem.Set(reflect.ValueOf(item))
					}
				}

			case reflect.Map:
				if err := dec.decodeVariadicType(src, fg); err != nil {
					return err
				}
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
			value := dec.parseType(src, flag, b)

			switch src.Elem().Type().Kind() {
			case reflect.Array, reflect.Slice:
				elem := src.Elem().Index(i)
				elem.Set(valueOf(value))
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

			src.SetMapIndex(valueOf(key), valueOf(value))
		}
	}

	return nil
}

func (dec *Decoder) parseType(src reflect.Value, tag ExternalTagType, data []byte) Term {
	switch tag {
	case EttNil:
		return nil

	case EttAtom, EttAtomUTF8, EttString:
		return dec.parseString(data)

	case EttSmallAtom, EttSmallAtomUTF8:
		return dec.parseString(data)

	case EttSmallInteger:
		kind := deepTypeOf(src.Type()).Kind()

		if kind == reflect.Int {
			return int(dec.parseSmallInteger(data))
		} else if kind == reflect.Uint {
			return uint(dec.parseSmallInteger(data))
		} else {
			return dec.parseSmallInteger(data)
		}

	case EttInteger:
		kind := src.Type().Kind()

		if kind == reflect.Int {
			return int(dec.parseInteger(data))
		} else {
			return dec.parseInteger(data)
		}

	case EttNewFloat:
		return dec.parseNewFloat(data)

	case EttFloat:
		return dec.parseFloat(data)

	case EttSmallBig:
		kind := src.Type().Kind()

		if kind == reflect.Int {
			return int(dec.parseSmallBig(data))
		} else {
			return dec.parseSmallBig(data)
		}

	case EttLargeBig:
		kind := src.Type().Kind()

		if kind == reflect.Int {
			return int(dec.parseLargeBig(data))
		} else {
			return dec.parseLargeBig(data)
		}

	case EttBinary:
		kind := src.Type().Kind()

		if kind == reflect.String {
			return string(data)
		} else {
			return data
		}

	case EttBitBinary:
		kind := src.Type().Kind()

		if kind == reflect.String {
			return string(data)
		} else {
			return data
		}

	case EttSmallTuple:
		return dec.parseSmallTuple(src)

	case EttLargeTuple:
		return dec.parseLargeTuple(src)

	case EttList:
		return dec.parseList(src)

	case EttMap:
		return dec.parseMap(src)
	}

	return nil
}

func (dec *Decoder) parseSmallTuple(src reflect.Value) any {
	sb, err := dec.rd.ReadByte()
	if err != nil {
		return err
	}

	arity := int(sb)

	sliceType := reflect.SliceOf(deepTypeOf(src.Type()))
	slice := reflect.MakeSlice(sliceType, 0, arity)

	for i := 0; i < arity; i++ {
		bflag, err := dec.rd.ReadByte()
		if err != nil {
			panic(ErrMalformedSmallTuple)
		}
		flag := ExternalTagType(bflag)

		_, b, err := dec.readType(flag)
		if err != nil {
			panic(ErrMalformedSmallTuple)
		}

		item := dec.parseType(src, flag, b)
		switch v := item.(type) {
		case reflect.Value:
			slice = reflect.Append(slice, v)

		default:
			if item != nil {
				itemOf := reflect.ValueOf(item)
				slice = reflect.Append(slice, itemOf)
			}
		}
	}

	return slice
}

func (dec *Decoder) parseLargeTuple(src reflect.Value) any {
	b := make([]byte, SizeLargeTupleArity)
	_, err := dec.rd.Read(b)
	if err != nil {
		panic(ErrMalformedLargeTuple)
	}

	arity := int(dec.parseInteger(b))

	slice := reflect.MakeSlice(reflect.TypeOf([]any{}), 0, arity)
	for i := 0; i < arity; i++ {
		bflag, err := dec.rd.ReadByte()
		if err != nil {
			panic(ErrMalformedLargeTuple)
		}
		fg := ExternalTagType(bflag)

		_, b, err := dec.readType(fg)
		if err != nil {
			panic(ErrMalformedLargeTuple)
		}

		item := dec.parseType(src, fg, b)
		switch v := item.(type) {
		case reflect.Value:
			slice = reflect.Append(slice, v)

		default:
			if item != nil {
				itemOf := reflect.ValueOf(item)
				slice = reflect.Append(slice, itemOf)
			}
		}
	}

	return slice
}

func (dec *Decoder) parseMap(src reflect.Value) any {
	var newMap reflect.Value

	switch src.Type().Kind() {
	case reflect.Map:
		bsize := make([]byte, SizeMapArity)
		n, err := dec.rd.Read(bsize)
		if err != nil {
			panic(ErrMalformedMap)
		}

		if n < 4 {
			panic(ErrMalformedMap)
		}

		arity := int(dec.parseInteger(bsize))

		var newMapType reflect.Type
		for i := 0; i < arity; i++ {
			// Key
			bflag, err := dec.rd.ReadByte()
			if err != nil {
				panic(ErrMalformedMap)
			}
			keyFlag := ExternalTagType(bflag)

			_, b, err := dec.readType(keyFlag)
			if err != nil {
				panic(ErrMalformedMap)
			}
			key := dec.parseType(src, keyFlag, b)

			// Value
			bflag, err = dec.rd.ReadByte()
			if err != nil {
				panic(ErrMalformedMap)
			}
			valueFlag := ExternalTagType(bflag)

			_, b, err = dec.readType(valueFlag)
			if err != nil {
				panic(ErrMalformedMap)
			}
			value := dec.parseType(src, valueFlag, b)

			keyOf := reflect.ValueOf(key)
			switch v := value.(type) {
			case reflect.Value:
				src.SetMapIndex(keyOf, v)
			default:
				if v != nil {
					newMapType = reflect.MapOf(keyOf.Type(), reflect.TypeOf(v))
					newMap = reflect.MakeMap(newMapType)
				} else {
					newMap = reflect.MakeMap(reflect.TypeOf(map[any]any{}))
				}

				newMap.SetMapIndex(keyOf, reflect.ValueOf(value))
			}
		}
	}
	return newMap
}

func (dec *Decoder) parseList(src reflect.Value) any {
	b := make([]byte, SizeListLength)
	_, err := dec.rd.Read(b)
	if err != nil {
		return ErrMalformedList
	}

	length := int(dec.parseInteger(b))

	slice := reflect.MakeSlice(reflect.TypeOf([]any{}), 0, length)
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

		item := dec.parseType(src, flag, b)
		switch v := item.(type) {
		case reflect.Value:
			slice.Set(v)

		default:
			if item != nil {
				slice = reflect.Append(slice, valueOf(item))
			}
		}
	}
	// read the 106 tail byte ([])
	dec.rd.ReadByte()

	return slice
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
