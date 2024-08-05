package goetf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"
	"reflect"
	"strings"
	"unicode/utf8"
)

var (
	typeOfBytes  = reflect.TypeOf([]byte(nil))
	typeOfBigInt = reflect.TypeOf(*big.NewInt(0))
)

// Marshaler is the interface implemented by types that can marshal themselves into valid ETF.
type Marshaler interface {
	MarshalETF(data []byte, src any) (err error)
}

// Marshal returns the ETF encoding of v.
func Marshal(v any) ([]byte, error) {
	en := NewEncoder(bytes.NewBuffer(make([]byte, 0)))
	if err := en.Encode(v); err != nil {
		return nil, err
	}
	return en.stream.readAll()
}

// An Encoder writes ETF values to an output stream.
type Encoder struct {
	w io.Writer

	stream *streamer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the ETF encoding of v to the stream.
func (e *Encoder) Encode(v any) error {
	e.init()
	return e.encode(v)
}

func (e *Encoder) init() {
	if e.stream == nil {
		e.stream = newStreamer(e.w)
	}
}

func (e *Encoder) encode(v any) error {
	vOf := valueOf(v)
	e.stream.writeByte(131)

	err := e.parseType(vOf)
	if err != nil {
		return err
	}

	return nil
}

// parseType parses the reflected value of src and writes its representation in bytes.
func (e *Encoder) parseType(src reflect.Value) error {
	var kind reflect.Kind
	if src.IsValid() {
		kind = src.Type().Kind()
	}

	switch kind {
	case reflect.Int:
		integer := src.Int()
		typ := e.assertNumericType(integer)
		e.parseType(valueOf(typ))

	case reflect.Uint:
		unsigned := src.Uint()
		typ := e.assertUintType(unsigned)
		e.parseType(valueOf(typ))

	case reflect.Uint8:
		unsigned := uint8(src.Uint())
		e.writeBytes([]byte{EttSmallInteger, unsigned})

	case reflect.Uint16:
		unsigned := uint16(src.Uint())
		data := binary.BigEndian.AppendUint16([]byte{0, 0}, unsigned)
		e.writeBytes([]byte{EttInteger}, data)

	case reflect.Int16:
		integer := int16(src.Int())
		data := binary.BigEndian.AppendUint16([]byte{0, 0}, uint16(integer))
		e.writeBytes([]byte{EttInteger}, data)

	case reflect.Uint32:
		unsigned := uint32(src.Uint())
		data := make([]byte, 4)
		binary.BigEndian.PutUint32(data, unsigned)
		e.writeBytes([]byte{EttInteger}, data)

	case reflect.Int32:
		integer := int32(src.Int())
		data := binary.BigEndian.AppendUint32(make([]byte, 0), uint32(integer))
		e.writeBytes([]byte{EttInteger}, data)

	case reflect.Int64:
		integer := src.Int()
		sign := byte(0)
		if integer < 0 {
			sign = 1
			integer *= -1
		}

		data := binary.LittleEndian.AppendUint64(make([]byte, 0), uint64(integer))
		if len(data) > 255 {
			return fmt.Errorf("parsing int64 error: invalid data size")
		}
		e.writeBytes([]byte{EttSmallBig, byte(len(data)), sign}, data)

	case reflect.Uint64:
		unsigned := src.Uint()
		data := binary.LittleEndian.AppendUint64(make([]byte, 0), unsigned)
		if len(data) > 255 {
			return fmt.Errorf("parsing int64 error: invalid data size")
		}
		e.writeBytes([]byte{EttSmallBig, byte(len(data)), 0}, data)

	case reflect.Float64, reflect.Float32:
		float := src.Float()
		data := binary.BigEndian.AppendUint64([]byte{}, math.Float64bits(float))
		e.writeBytes([]byte{EttNewFloat}, data)

	case reflect.String:
		str := src.String()
		data := []byte(str)
		isValidUTF8 := utf8.Valid(data)
		blen := make([]byte, 0)

		var tag ExternalTagType

		if len(data) <= 255 && isValidUTF8 && !strings.Contains(str, " ") {
			blen = binary.BigEndian.AppendUint16(blen, uint16(len(data)))[1:]
			tag = EttSmallAtomUTF8
		} else {
			blen = binary.BigEndian.AppendUint16(blen, uint16(len(data)))
			tag = EttString
		}

		e.writeByte(tag)
		e.writeBytes(blen, data)

	case reflect.Bool:
		b := src.Bool()
		e.writeByte(EttSmallAtomUTF8)
		if b {
			e.writeBytes([]byte{4, 116, 114, 117, 101})
		} else {
			e.writeBytes([]byte{5, 102, 97, 108, 115, 101})
		}

	case reflect.Chan, reflect.Func:
		e.writeNil()

	case reflect.Interface:
		v := e.assertInterfaceType(src.Interface())
		if v == nil {
			e.writeNil()
			return nil
		}

		if err := e.parseType(derefValueOf(v)); err != nil {
			return err
		}

	case reflect.Pointer:
		if src.IsNil() {
			e.writeNil()
		} else {
			e.parseType(derefValueOf(src))
		}

	case reflect.Slice:
		if src.IsNil() {
			e.writeNil()
			return nil
		}

		if src.Type() == typeOfBytes {
			e.parseBinary(src)
			break
		}

		tpLen, isLarge := src.Len(), false
		if tpLen <= 255 {
			e.writeByte(EttSmallTuple, byte(tpLen))
		} else {
			e.writeByte(EttLargeTuple)
			isLarge = true
		}

		if isLarge {
			arity := binary.BigEndian.AppendUint32(make([]byte, 0), uint32(tpLen))
			e.writeBytes(arity)
		}

		for i := 0; i < tpLen; i++ {
			elem := src.Index(i)
			if elem.IsValid() {
				if err := e.parseType(valueOf(elem)); err != nil {
					return err
				}

				if elem.Type().Kind() == reflect.Pointer && elem.IsNil() {
					e.writeNil()
				}
			}
		}

	case reflect.Array:
		e.writeByte(EttList)
		arrLen := src.Len()

		blen := make([]byte, 4)
		binary.BigEndian.PutUint32(blen, uint32(arrLen))
		e.writeBytes(blen)

		for i := 0; i < arrLen; i++ {
			elem := src.Index(i)
			if elem.IsValid() {
				if err := e.parseType(derefValueOf(elem)); err != nil {
					return err
				}

				if elem.Type().Kind() == reflect.Pointer && elem.IsNil() {
					e.writeNil()
				}
			}
		}

		e.writeByte(EttNil)

	case reflect.Map:
		e.writeByte(EttMap)
		mapLen := src.Len()

		keys := src.MapKeys()

		blen := make([]byte, 4)
		binary.BigEndian.PutUint32(blen, uint32(mapLen))
		e.writeBytes(blen)

		for _, key := range keys {
			if err := e.parseType(key); err != nil {
				return err
			}

			value := src.MapIndex(key)
			if value.IsValid() {
				if err := e.parseType(derefValueOf(value)); err != nil {
					return err
				}

				if value.Type().Kind() == reflect.Pointer && value.IsNil() {
					e.writeNil()
				}
			}
		}

	case reflect.Struct:
		if src.Type() == typeOfBigInt {
			return e.writeLargeBig(src)
		}

		e.writeByte(EttMap)
		length := src.NumField()

		blen := binary.BigEndian.AppendUint32([]byte{}, uint32(length))
		e.writeBytes(blen)

		for i := 0; i < length; i++ {
			name := derefValueOf(src).Type().Field(i).Name
			field := derefValueOf(src.Field(i))
			tag := src.Type().Field(i).Tag.Get("etf")

			var key string
			if tag != "" {
				key = tag
			} else {
				key = name
			}

			if err := e.parseType(valueOf(key)); err != nil {
				return err
			}

			if err := e.parseType(field); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Encoder) parseBinary(src reflect.Value) {
	l := src.Len()

	blen := binary.BigEndian.AppendUint32([]byte{}, uint32(l))
	if l%8 == 0 {
		e.writeByte(EttBinary)
		e.writeBytes(blen)
	} else {
		e.writeByte(EttBitBinary)
		e.writeBytes(blen)
		e.writeByte(8)
	}

	for i := 0; i < l; i++ {
		b := byte(src.Index(i).Uint())
		e.writeByte(b)
	}
}

func (e *Encoder) writeBytes(slices ...[]byte) (n int, err error) {
	c := 0
	for _, s := range slices {
		if n, err = e.stream.write(s); err != nil {
			return
		}
		c += n
	}
	n = c
	return
}

func (e *Encoder) writeByte(b ...byte) (n int, err error) {
	n, err = e.writeBytes(b)
	return
}

func (e *Encoder) writeNil() {
	e.writeBytes([]byte{119, 3, 110, 105, 108})
}

func (e *Encoder) writeLargeBig(src reflect.Value) error {
	num, ok := src.Interface().(big.Int)
	if !ok {
		return fmt.Errorf("encode error: invalid big int number")
	}
	bigNum := &num
	b := bigNum.Bytes()

	e.writeByte(EttLargeBig)
	e.writeBytes(binary.BigEndian.AppendUint32([]byte{}, uint32(len(b))))

	sign := bigNum.Sign()
	if sign < 0 {
		e.writeByte(1)
	} else {
		e.writeByte(0)
	}

	toLittleEndian(b)
	e.writeBytes(b)
	return nil
}

func (e *Encoder) assertNumericType(i int64) any {
	switch {
	case 0 < i && i < math.MaxUint8:
		return uint8(i)
	case math.MinInt16 < i && i < math.MaxInt16:
		return int16(i)
	case math.MinInt32 < i && i < math.MaxInt32:
		return int32(i)
	default:
		return i
	}
}

func (e *Encoder) assertUintType(i uint64) any {
	switch {
	case 0 < i && i < math.MaxUint8:
		return uint8(i)
	case 0 < i && i < math.MaxUint16:
		return uint16(i)
	case 0 < i && i < math.MaxUint32:
		return uint32(i)
	default:
		return i
	}
}

func (e *Encoder) assertInterfaceType(v any) any {
	switch v := v.(type) {
	case
		[]uint, []uint8, []uint16, []uint32, []uint64,
		[]int, []int16, []int32, []int64,
		[]float32, []float64,
		[]string,
		[]bool,
		uint, uint8, uint16, uint32, uint64,
		int, int16, int32, int64,
		float32, float64,
		string,
		bool:
		return v
	}

	return nil
}
