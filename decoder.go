package goetf

import (
	"bufio"
	"bytes"
	"io"
	"reflect"

	"github.com/philpearl/intern"
)

// Decoder allows embedding ETF byte slices into valid Go code.
type Decoder struct {
	buf       []byte
	rd        *bufio.Reader
	atomCache *intern.Intern
	// todo: mu        sync.Mutex
}

func NewDecoder(b []byte) *Decoder {
	return &Decoder{
		buf:       b,
		rd:        bufio.NewReaderSize(bytes.NewReader(b), len(b)),
		atomCache: intern.New(2048),
	}
}

// Decode reads the next ETF-encoded value from its input and stores it in the value pointed to by v.
func (dec *Decoder) Decode(v any) error {
	if v == nil {
		return ErrNilDecodeValue
	}

	return dec.decode(v)
}

// DecodePacket reads the next ETF-encoded packet and stores it in the value pointed to by v.
func (dec *Decoder) DecodePacket(packet []byte, v any) error {
	if v == nil {
		return ErrNilDecodeValue
	}

	dec.rd.Reset(bytes.NewReader(packet))
	dec.buf = packet
	return dec.decode(v)
}

func (dec *Decoder) decode(v any) error {
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
		b, err = dec.rd.ReadByte()
		if err != nil && err != io.EOF {
			return err
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

func (dec *Decoder) decodeStatic(tag ExternalTagType, v any) error {
	switch tag {
	case EttAtom, EttAtomUTF8:
		_, b, err := dec.readAtomUTF8()
		if err != nil {
			return err
		}

		ptr, ok := (v).(*string)
		if !ok {
			return ErrDecodeType
		}

		s := dec.parseString(b)
		(*ptr) = dec.atomCache.Deduplicate(s)

	case EttSmallAtom, EttSmallAtomUTF8:
		_, b, err := dec.readSmallAtomUTF8()
		if err != nil {
			return err
		}

		ptr, ok := (v).(*string)
		if !ok {
			return ErrDecodeType
		}

		s := dec.parseString(b)
		(*ptr) = dec.atomCache.Deduplicate(s)

	case EttString:
		_, b, err := dec.readString()
		if err != nil {
			return err
		}

		switch v := v.(type) {
		case *string:
			(*v) = dec.parseString(b)

		default:
			return ErrDecodeType
		}

	case EttSmallInteger:
		_, b, err := dec.readSmallInteger()
		if err != nil {
			return err
		}

		switch v := v.(type) {
		case *uint8:
			(*v) = dec.parseSmallInteger(b)

		case *uint:
			(*v) = uint(dec.parseSmallInteger(b))

		case *int:
			(*v) = int(dec.parseSmallInteger(b))

		default:
			return ErrDecodeType
		}

	case EttInteger:
		_, b, err := dec.readInteger()
		if err != nil {
			return err
		}

		switch v := v.(type) {
		case *int32:
			(*v) = dec.parseInteger(b)

		case *int:
			(*v) = int(dec.parseInteger(b))

		default:
			return ErrDecodeType
		}

	case EttFloat:
		_, b, err := dec.readFloat()
		if err != nil {
			return err
		}

		switch v := v.(type) {
		case *float64:
			(*v) = dec.parseFloat(b)

		case *float32:
			(*v) = float32(dec.parseNewFloat(b))

		default:
			return ErrDecodeType
		}

	case EttNewFloat:
		_, b, err := dec.readNewFloat()
		if err != nil {
			return err
		}

		switch v := v.(type) {
		case *float64:
			(*v) = dec.parseNewFloat(b)

		case *float32:
			(*v) = float32(dec.parseNewFloat(b))

		default:
			return ErrDecodeType
		}

	case EttSmallBig:
		_, b, err := dec.readSmallBig()
		if err != nil {
			return err
		}

		switch v := v.(type) {
		case *int64:
			(*v) = dec.parseSmallBig(b)

		case *int:
			(*v) = int(dec.parseSmallBig(b))

		default:
			return ErrDecodeType
		}

	case EttLargeBig:
		_, b, err := dec.readLargeBig()
		if err != nil {
			return err
		}

		switch v := v.(type) {
		case *int64:
			(*v) = dec.parseLargeBig(b)

		case *int:
			(*v) = int(dec.parseLargeBig(b))

		default:
			return ErrDecodeType
		}

	case EttBinary:
		_, b, err := dec.readBinary()
		if err != nil {
			return err
		}

		switch v := v.(type) {
		case *[]byte:
			(*v) = b

		default:
			return ErrDecodeType
		}

	case EttBitBinary:
		_, b, err := dec.readBitBinary()
		if err != nil {
			return err
		}

		switch v := v.(type) {
		case *[]byte:
			(*v) = b

		default:
			return ErrDecodeType
		}

	case EttSmallTuple:
		sb, err := dec.rd.ReadByte()
		if err != nil {
			return err
		}

		arity := int(sb)

		vOf := reflect.ValueOf(v)
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

			item := dec.parseType(flag, b)
			vOf.Index(i).Set(reflect.ValueOf(item))
		}

	case EttLargeTuple:
		b := make([]byte, SizeLargeTupleArity)
		_, err := dec.rd.Read(b)
		if err != nil {
			return ErrMalformedLargeTuple
		}

		arity := int(dec.parseInteger(b))

		vOf := reflect.ValueOf(v)
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

			item := dec.parseType(flag, b)
			vOf.Index(i).Set(reflect.ValueOf(item))
		}

	case EttList:
		b := make([]byte, SizeListLength)
		_, err := dec.rd.Read(b)
		if err != nil {
			return ErrMalformedLargeTuple
		}

		length := int(dec.parseInteger(b))

		vOf := reflect.ValueOf(v)
		for i := 0; i < length; i++ {
			bflag, err := dec.rd.ReadByte()
			if err != nil {
				return ErrMalformedLargeTuple
			}
			flag := ExternalTagType(bflag)

			_, b, err := dec.readType(flag)
			if err != nil {
				return ErrMalformedLargeTuple
			}

			item := dec.parseType(flag, b)
			vOf.Index(i).Set(reflect.ValueOf(item))
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

		vOf := reflect.ValueOf(v)
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
			key := dec.parseType(keyFlag, b)

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
			value := dec.parseType(valueFlag, b)

			// Set key and value
			vOf.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		}
	}

	return nil
}
