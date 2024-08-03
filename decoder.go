package goetf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"

	"github.com/philpearl/intern"
)

func Unmarshal(data []byte, v any) error {
	dec := NewDecoder(bytes.NewReader(data))
	return dec.Decode(v)
}

type Decoder struct {
	r io.Reader

	scan *scanner

	cache *intern.Intern

	err error

	dirty bool
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode(v any) error {
	if d.err != nil {
		return d.err
	}

	d.init()
	return d.decode(v)
}

func (d *Decoder) init() {
	if d.cache == nil {
		d.cache = intern.New(2048)
	}

	if d.scan == nil {
		d.scan = newScanner(d.r)
	}
}

func (d *Decoder) decode(v any) error {
	// if the buffer is not dirty check for the version number
	if !d.dirty {
		ver, err := d.scan.readByte()
		if err != nil {
			return err
		}

		if ver != Version {
			return ErrMalformed
		}

		d.dirty = true
	}

	for !d.scan.eof() {
		elem, err := d.readNext()
		if err != nil {
			return err
		}

		parsed := d.decodeValue(elem, v)
		parsedOf := valueOf(parsed)
		vOf := derefValueOf(v)
		if vOf.Type().Kind() == reflect.Map && parsedOf.Type().Kind() == reflect.Map {
			keys := parsedOf.MapKeys()
			for _, key := range keys {
				value := parsedOf.MapIndex(key)
				vOf.SetMapIndex(key, value)
			}
		} else {
			if vOf.Type().Kind() != reflect.Struct {
				vOf.Set(valueOf(parsed))
			}
		}
	}

	return nil
}

func (d *Decoder) readNext() (*binaryElement, error) {
	typeTag, err := d.scan.readByte()
	if err != nil {
		return nil, ErrMalformed
	}

	dst := newBinaryElement(typeTag, nil)
	switch typeTag {
	default:
		_, data, err := d.readStaticType(typeTag)
		if err != nil {
			return nil, err
		}
		dst.put(typeTag, data)

	case EttSmallTuple:
		bArity, err := d.scan.readByte()
		if err != nil {
			return nil, ErrMalformedSmallTuple
		}

		arity := int(bArity)
		if arity == 0 {
			return nil, ErrMalformedSmallTuple
		}

		for range arity {
			elem, err := d.readNext()
			if err != nil {
				return nil, err
			}

			dst.append(typeTag, elem)
		}

	case EttLargeTuple:
		_, bArity, err := d.scan.readN(SizeLargeTupleArity)
		if err != nil {
			return nil, ErrMalformedLargeTuple
		}

		arity := int(binary.BigEndian.Uint32(bArity))
		if arity == 0 {
			return nil, ErrMalformedLargeTuple
		}

		for range arity {
			elem, err := d.readNext()
			if err != nil {
				return nil, err
			}

			dst.append(typeTag, elem)
		}

	case EttList:
		_, bLen, err := d.scan.readN(SizeListLength)
		if err != nil {
			return nil, ErrMalformedList
		}

		length := int(binary.BigEndian.Uint32(bLen))
		if length == 0 {
			return nil, ErrMalformedList
		}

		for range length + 1 {
			elem, err := d.readNext()
			if err != nil {
				return nil, err
			}

			dst.append(typeTag, elem)
		}

	case EttMap:
		_, bArity, err := d.scan.readN(SizeMapArity)
		if err != nil {
			return nil, ErrMalformedMap
		}

		arity := int(binary.BigEndian.Uint32(bArity))
		if arity == 0 {
			return nil, ErrMalformedMap
		}

		for range arity {
			keyElem, err := d.readNext()
			if err != nil {
				return nil, ErrMalformedMap
			}

			valElem, err := d.readNext()
			if err != nil {
				return nil, ErrMalformedMap
			}

			dst.append(typeTag, keyElem)
			dst.append(typeTag, valElem)
		}
	}

	return dst, nil
}

func (d *Decoder) decodeValue(elem *binaryElement, v any) any {
	vOf := derefValueOf(v)
	kind := vOf.Type().Kind()

	switch elem.tag {
	default:
		parsed := d.parseStaticType(vOf.Type().Kind(), elem.tag, elem.body)
		vOf.Set(valueOf(parsed))
		return parsed

	case EttSmallTuple, EttLargeTuple:
		if len(elem.items) > 0 && (kind == reflect.Slice) {
			fmt.Println(elem)
			tuple := reflect.MakeSlice(vOf.Type(), len(elem.items), len(elem.items))
			for i, item := range elem.items {
				d.decodeValue(item, tuple.Index(i))
			}

			fmt.Println(tuple)
			return tuple
		}

	case EttList:
		if len(elem.items) > 0 && (kind == reflect.Array) {
			arrType := reflect.ArrayOf(len(elem.items)-1, derefValueOf(vOf).Type().Elem())
			arr := derefValueOf(reflect.New(arrType))

			for i, item := range elem.items {
				if item.tag != EttNil {
					d.decodeValue(item, derefValueOf(arr).Index(i))
				}
			}

			return arr
		}

	case EttMap:
		if len(elem.dict) > 0 && (kind == reflect.Map || kind == reflect.Struct) {
			if kind == reflect.Map {
				mapType := reflect.MapOf(vOf.Type().Key(), vOf.Type().Elem())
				m := reflect.MakeMap(mapType)

				for i := 0; i < len(elem.dict)-1; i += 2 {
					keyElem := elem.dict[i]
					valElem := elem.dict[i+1]

					keyOf := derefValueOf(reflect.New(vOf.Type().Key()))
					d.decodeValue(keyElem, keyOf)

					valOf := derefValueOf(reflect.New(vOf.Type().Elem()))
					d.decodeValue(valElem, valOf)

					m.SetMapIndex(keyOf, valOf)
				}

				return m
			}

			if kind == reflect.Struct {
				fields := map[string]reflect.Value{}
				for i := 0; i < vOf.NumField(); i++ {
					fieldTag := vOf.Type().Field(i).Tag.Get("etf")
					if fieldTag != "" {
						fields[fieldTag] = vOf.Field(i)
					}
				}

				str := ""
				for i := 0; i < len(elem.dict)-1; i += 2 {
					keyElem := elem.dict[i]

					keyOf := derefValueOf(valueOf(&str))
					d.decodeValue(keyElem, keyOf)

					etfTag := keyOf.String()
					if field, ok := fields[etfTag]; ok {
						valElem := elem.dict[i+1]

						valOf := derefValueOf(field)
						valParsed := d.decodeValue(valElem, valOf)

						field.Set(valueOf(valParsed))
					}
				}
			}
		}
	}

	return nil
}
