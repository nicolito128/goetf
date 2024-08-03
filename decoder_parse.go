package goetf

import (
	"encoding/binary"
	"math"
	"reflect"
)

func (d *Decoder) parseStaticType(kind reflect.Kind, tag ExternalTagType, data []byte) Term {
	switch tag {
	case EttNil:
		return nil

	case EttString:
		switch kind {
		default:
			return d.parseString(data)
		case reflect.Slice:
			return data
		}

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
		case reflect.Int:
			return int(d.parseLargeBig(data))
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
	sign := b[len(b)-1]
	rest := b[:len(b)-1]

	bits := binary.LittleEndian.Uint64(rest)
	smallBig := int64(bits)

	if sign == 1 {
		smallBig *= -1
	}

	return smallBig
}

func (d *Decoder) parseLargeBig(b []byte) int64 {
	sign := b[len(b)-1]
	rest := b[:len(b)-1]

	bits := binary.LittleEndian.Uint64(rest)
	largeBig := int64(bits)

	if sign == 1 {
		largeBig *= -1
	}

	return largeBig
}
