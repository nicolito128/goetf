package goetf

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"

	"github.com/philpearl/intern"
)

// Unmarshaler is the interface implemented by types that can unmarshal a ETF description of themselves.
// The input can be assumed to be a valid encoding of a ETF value.
// UnmarshalETF must copy the ETF data if it wishes to retain the data after returning.
type Unmarshaler interface {
	UnmarshalETF(data []byte, dst any) (err error)
}

// Unmarshal parses the ETF-encoded data and stores the result in the value pointed to by v.
func Unmarshal(data []byte, v any) error {
	dec := NewDecoder(bytes.NewReader(data))
	return dec.Decode(v)
}

// A Decoder reads and decodes ETF values from an input stream buffer.
type Decoder struct {
	// buffer reader
	r io.Reader
	// scanner for buffer
	scan *scanner
	// cache for atoms
	cache *intern.Intern
	// number of atoms that can be stored, by default it's 4096
	cacheSize int
	// if the buffer was touched, starting with the version flag
	dirty bool
}

// NewDecoder returns a new *Decoder that reads from r.
//
// The decoder uses its own buffering.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r, cacheSize: 4096}
}

// Decode reads the next ETF-encoded data from its buffer and stores it in the value pointed to by v.
func (d *Decoder) Decode(v any) error {
	d.init()
	return d.decode(v)
}

// SetCache sets a new atom cache with the input capacity.
//
// The default cache capacity is 4096.
func (d *Decoder) SetCache(cap int) {
	d.cacheSize = cap
	d.cache = intern.New(cap)
}

func (d *Decoder) init() {
	if d.cache == nil {
		d.cache = intern.New(d.cacheSize)
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
			return errMalformed
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
		return nil, errMalformed
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
			return nil, errMalformedSmallTuple
		}

		arity := int(bArity)
		if arity == 0 {
			return nil, errMalformedSmallTuple
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
			return nil, errMalformedLargeTuple
		}

		arity := int(binary.BigEndian.Uint32(bArity))
		if arity == 0 {
			return nil, errMalformedLargeTuple
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
			return nil, errMalformedList
		}

		length := int(binary.BigEndian.Uint32(bLen))
		if length == 0 {
			return nil, errMalformedList
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
			return nil, errMalformedMap
		}

		arity := int(binary.BigEndian.Uint32(bArity))
		if arity == 0 {
			return nil, errMalformedMap
		}

		for range arity {
			keyElem, err := d.readNext()
			if err != nil {
				return nil, errMalformedMap
			}

			valElem, err := d.readNext()
			if err != nil {
				return nil, errMalformedMap
			}

			dst.append(typeTag, keyElem)
			dst.append(typeTag, valElem)
		}
	}

	return dst, nil
}

func (d *Decoder) decodeValue(elem *binaryElement, v any) any {
	vOf := derefValueOf(v)

	var kind reflect.Kind
	if v != nil {
		kind = vOf.Type().Kind()
	}

	switch elem.tag {
	default:
		parsed := d.parseStaticType(vOf.Type().Kind(), elem.tag, elem.body)

		if v != nil {
			parsedOf := valueOf(parsed)
			if parsedOf.IsValid() {
				vOf.Set(parsedOf)
			}

			return parsed
		}

	case EttSmallTuple, EttLargeTuple:
		if len(elem.items) > 0 {
			if kind == reflect.Slice {
				tuple := reflect.MakeSlice(vOf.Type(), len(elem.items), len(elem.items))
				for i, item := range elem.items {
					d.decodeValue(item, tuple.Index(i))
				}

				return tuple
			}

			if kind == reflect.Interface {
				tuple := make([]any, len(elem.items))
				tupleOf := valueOf(tuple)
				for i, item := range elem.items {
					d.decodeValue(item, tupleOf.Index(i))
				}

				return tuple
			}
		}

	case EttList:
		if len(elem.items) > 0 {
			if kind == reflect.Interface {
				arrType := reflect.ArrayOf(len(elem.items)-1, derefValueOf(vOf).Type())
				arr := derefValueOf(reflect.New(arrType))
				for i, item := range elem.items {
					if item.tag != EttNil {
						d.decodeValue(item, (arr).Index(i))
					}
				}

				return arr
			}

			if kind == reflect.Array {
				arrType := reflect.ArrayOf(len(elem.items)-1, derefValueOf(vOf).Type().Elem())
				arr := derefValueOf(reflect.New(arrType))

				for i, item := range elem.items {
					if item.tag != EttNil {
						d.decodeValue(item, derefValueOf(arr).Index(i))
					}
				}

				return arr
			}
		}

	case EttMap:
		if len(elem.dict) > 0 {
			if kind == reflect.Map {
				mapType := reflect.MapOf(vOf.Type().Key(), vOf.Type().Elem())
				m := reflect.MakeMap(mapType)

				for i := 0; i < len(elem.dict)-1; i += 2 {
					keyElem := elem.dict[i]
					valElem := elem.dict[i+1]

					keyOf := derefValueOf(reflect.New(vOf.Type().Key()))
					key := d.decodeValue(keyElem, keyOf)

					valOf := derefValueOf(reflect.New(vOf.Type().Elem()))
					val := d.decodeValue(valElem, valOf)

					if val != nil && key != nil {
						m.SetMapIndex(valueOf(key), derefValueOf(val))
					} else {
						m.SetMapIndex(derefValueOf(keyOf), derefValueOf(valOf))
					}
				}

				return m
			}

			if kind == reflect.Interface {
				m := map[any]any{}
				mapOf := valueOf(m)

				for i := 0; i < len(elem.dict)-1; i += 2 {
					keyElem := elem.dict[i]
					valElem := elem.dict[i+1]

					keyOf := reflect.New(mapOf.Type().Elem())
					key := d.decodeValue(keyElem, keyOf)

					valOf := reflect.New(mapOf.Type().Elem())
					val := d.decodeValue(valElem, valOf)

					if val != nil && key != nil {
						mapOf.SetMapIndex(valueOf(key), derefValueOf(val))
					} else {
						mapOf.SetMapIndex(derefValueOf(keyOf), derefValueOf(valOf))
					}
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

						if valParsed != nil {
							field.Set(valueOf(valParsed))
						} else {
							field.Set(derefValueOf(valOf))
						}
					}
				}
			}
		}
	}

	return nil
}
