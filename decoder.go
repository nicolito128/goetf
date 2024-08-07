package goetf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"slices"

	"github.com/philpearl/intern"
)

const defaultCacheSize = 1024 * 1024

var defaultCache = intern.New(defaultCacheSize)

// Unmarshaler is the interface implemented by types that can unmarshal a ETF description of themselves.
// The input can be assumed to be a valid encoding of a ETF value.
// UnmarshalETF must copy the ETF data if it wishes to retain the data after returning.
type Unmarshaler interface {
	UnmarshalETF(data []byte, dst any) (err error)
}

// Unmarshal parses the ETF-encoded data and stores the result in the value pointed to by v.
func Unmarshal(data []byte, v any) error {
	dec := NewDecoder(bytes.NewReader(data))
	dec.cache = defaultCache
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
	// error to check
	err error
}

// NewDecoder returns a new *Decoder that reads from r.
//
// The decoder uses its own buffering.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r, cacheSize: 1024 * 1024}
}

// Decode reads the next ETF-encoded data from its buffer and stores it in the value pointed to by v.
func (d *Decoder) Decode(v any) error {
	d.init()
	return d.decode(v)
}

// SetCache sets a new atom cache with the input capacity.
//
// The default cache capacity is 1048576.
func (d *Decoder) SetCache(cap int) {
	d.cacheSize = cap
	d.cache = intern.New(cap)
}

func (d *Decoder) init() {
	if d.cache == nil {
		d.cache = intern.New(defaultCacheSize)
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

	vOf := valueOf(v)
	if vOf.Type().Kind() == reflect.Pointer {
		vOf = derefValueOf(vOf.Elem())
	}

	switch vOf.Type().Kind() {
	case reflect.Map, reflect.Slice:
		if vOf.IsNil() {
			ptr := reflect.New(vOf.Type())
			vOf.Set(ptr.Elem())
		}
	}

	for !d.scan.eof() {
		elem, err := d.readNext()
		if err != nil {
			return err
		}

		parsed := d.decodeValue(elem, v)
		if d.err != nil {
			return d.err
		}

		if parsed != nil {
			parsedOf := derefValueOf(parsed)
			if parsedOf.IsValid() {
				if vOf.Type().Kind() == reflect.Map || parsedOf.Type().Kind() == reflect.Map {
					return nil
				}

				if vOf.Type() == parsedOf.Type() {
					vOf.Set(parsedOf)
				}
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
	var vOf reflect.Value
	var kind reflect.Kind

	vOf = valueOf(v)
	if !vOf.IsValid() {
		d.err = fmt.Errorf("invalid value to decode")
		return nil
	}
	kind = derefTypeOf(vOf.Type()).Kind()

	switch elem.tag {
	default:
		if vOf.IsValid() {
			parsed := d.parseStaticType(kind, elem.tag, elem.body)
			if parsed != nil {
				return parsed
			}
		}

	case EttSmallTuple, EttLargeTuple:
		if len(elem.items) > 0 {
			if kind == reflect.Interface {
				return d.decodeAnyTuple(elem, vOf)
			} else {
				return d.decodeTuple(elem, vOf)
			}
		}

	case EttList:
		if len(elem.items) > 0 {
			if kind == reflect.Interface {
				return d.decodeAnyList(elem, vOf)
			} else {
				return d.decodeList(elem, vOf)
			}
		}

	case EttMap:
		if len(elem.dict) > 0 {
			switch kind {
			case reflect.Struct:
				if vOf.Type() == typeOfBigInt {
					return vOf.Interface().(big.Int)
				} else {
					if vOf.Type().Kind() == reflect.Pointer {
						if vOf.IsNil() && vOf.CanSet() {
							ptr := reflect.New(vOf.Type().Elem())
							ptr.Elem().SetZero()
							vOf.Set(ptr)
						}
					}

					return d.decodeStruct(elem, vOf)
				}

			case reflect.Interface:
				return d.decodeAnyMap(elem)

			case reflect.Map:
				parsedOf := valueOf(d.decodeMap(elem, vOf))
				if vOf.Type().Kind() == reflect.Map && parsedOf.Type().Kind() == reflect.Map {
					keys := parsedOf.MapKeys()
					for _, key := range keys {
						mValOf := parsedOf.MapIndex(key)
						if !key.IsValid() || !mValOf.IsValid() {
							return nil
						}

						setValueNotPtr(vOf.Type().Elem(), mValOf, func(out reflect.Value) {
							vOf.SetMapIndex(key.Convert(vOf.Type().Key()), out.Convert(vOf.Type().Elem()))
						})
					}

					return vOf
				}
			}
		}
	}

	return nil
}

func (d *Decoder) decodeTuple(elem *binaryElement, src reflect.Value) any {
	if src.Type().Kind() == reflect.Pointer {
		src = derefValueOf(src.Elem())
	}
	if src.Type().Kind() != reflect.Slice {
		d.err = fmt.Errorf("error trying to decode a no-slice type")
		return nil
	}

	length := src.Len()
	for i, item := range elem.items {
		if length > 0 {
			srcElem := src.Index(i)

			parsed := d.decodeValue(item, srcElem)
			if parsed == nil {
				return nil
			}

			parsedOf := valueOf(parsed)
			if !parsedOf.IsValid() {
				return nil
			}

			setValueNotPtr(srcElem.Type(), parsedOf, func(out reflect.Value) {
				srcElem.Set(out.Convert(srcElem.Type()))
			})
		} else {
			newElem := reflect.New(src.Type().Elem()).Elem()
			if !newElem.IsValid() {
				return nil
			}

			parsed := d.decodeValue(item, newElem)
			if parsed == nil {
				return nil
			}

			parsedOf := valueOf(parsed)
			if !parsedOf.IsValid() {
				return nil
			}

			setValueNotPtr(newElem.Type(), parsedOf, func(out reflect.Value) {
				src = reflect.Append(src, out.Convert(src.Type().Elem()))
			})
		}
	}

	return src.Interface()
}

func (d *Decoder) decodeAnyTuple(elem *binaryElement, src reflect.Value) any {
	if src.Type().Kind() == reflect.Pointer {
		src = derefValueOf(src.Elem())
	}

	tuple := reflect.MakeSlice(src.Type(), len(elem.items), len(elem.items))
	for i, item := range elem.items {
		tpElem := derefValueOf(tuple.Index(i))
		if tpElem.IsValid() {
			parsed := d.decodeValue(item, tpElem)
			if parsed != nil {
				tpElem.Set(valueOf(parsed))
			}
		}
	}

	return tuple.Interface()
}

func (d *Decoder) decodeList(elem *binaryElement, src reflect.Value) any {
	if src.Type().Kind() == reflect.Pointer {
		src = derefValueOf(src.Elem())
	}
	if src.Type().Kind() != reflect.Array {
		d.err = fmt.Errorf("error trying to decode a no-array type")
		return nil
	}

	arrLength := 0
	// Improper list check ([a | b])
	if elem.items[len(elem.items)-1].tag != EttNil {
		arrLength = len(elem.items)
	} else {
		arrLength = len(elem.items) - 1
	}

	arrType := reflect.ArrayOf(arrLength, src.Type().Elem())
	arr := reflect.New(arrType).Elem()

	for i := 0; i < arrLength; i++ {
		arrElem := arr.Index(i)
		if arrElem.IsValid() {
			item := elem.items[i]

			parsed := d.decodeValue(item, arrElem)
			if parsed != nil {
				parsedOf := valueOf(parsed)
				setValueNotPtr(arrElem.Type(), parsedOf, func(out reflect.Value) {
					arrElem.Set(out.Convert(arr.Type().Elem()))
				})
			}
		}
	}

	return arr.Interface()
}

func (d *Decoder) decodeAnyList(elem *binaryElement, src reflect.Value) any {
	if src.Type().Kind() == reflect.Pointer {
		src = derefValueOf(src.Elem())
	}

	arrLength := 0
	// Improper list check ([a | b])
	if elem.items[len(elem.items)-1].tag != EttNil {
		arrLength = len(elem.items)
	} else {
		arrLength = len(elem.items) - 1
	}

	arr := reflect.MakeSlice(reflect.SliceOf(src.Type()), arrLength, arrLength)
	for i := 0; i < arrLength; i++ {
		arrElem := (arr).Index(i)
		if arrElem.IsValid() {
			item := elem.items[i]

			parsed := d.decodeValue(item, arrElem)
			if parsed != nil {
				arrElem.Set(valueOf(parsed))
			}
		}
	}
	return arr.Interface()
}

func (d *Decoder) decodeMap(elem *binaryElement, src reflect.Value) any {
	if src.Type().Kind() == reflect.Pointer {
		src = derefValueOf(src.Elem())
	}
	if src.Type().Kind() != reflect.Map {
		d.err = fmt.Errorf("error trying to decode a no-map type")
		return nil
	}

	var mapType reflect.Type
	if src.Type().Elem().Kind() == reflect.Pointer {
		mapType = reflect.MapOf(src.Type().Key(), src.Type().Elem().Elem())
	} else {
		mapType = reflect.MapOf(src.Type().Key(), src.Type().Elem())
	}

	m := reflect.MakeMap(mapType)

	for i := 0; i < len(elem.dict)-1; i += 2 {
		keyElem := elem.dict[i]
		valElem := elem.dict[i+1]

		keyOf := reflect.New(m.Type().Key()).Elem()
		key := d.decodeValue(keyElem, keyOf)

		valOf := reflect.New(m.Type().Elem()).Elem()
		value := d.decodeValue(valElem, valOf)

		if key != nil && value != nil {
			keyOf = valueOf(key)
			valOf = valueOf(value)
		}

		if !valueOf(key).IsZero() && keyOf.IsValid() && valOf.IsValid() && !keyOf.IsZero() {
			m.SetMapIndex(keyOf, valOf)
		}
	}

	return m.Interface()
}

func (d *Decoder) decodeAnyMap(elem *binaryElement) any {
	m := reflect.MakeMap(reflect.TypeOf(map[string]any{}))

	for i := 0; i < len(elem.dict)-1; i += 2 {
		keyElem := elem.dict[i]
		valElem := elem.dict[i+1]

		keyOf := reflect.New(m.Type().Key())
		key := d.decodeValue(keyElem, keyOf)

		valOf := reflect.New(m.Type().Elem())
		val := d.decodeValue(valElem, valOf)

		if keyOf.IsValid() && valOf.IsValid() {
			if !keyOf.Comparable() && !valueOf(key).Comparable() {
				panic("key of any map must be a comparable")
			}

			if key != nil && val != nil {
				keyOf = valueOf(key)
				valOf = valueOf(val)
			}

			if !keyOf.IsZero() && valOf.IsValid() {
				m.SetMapIndex(keyOf, valOf)
			}
		}
	}

	return m.Interface()
}

func (d *Decoder) decodeStruct(elem *binaryElement, src reflect.Value) any {
	if src.Type().Kind() == reflect.Pointer {
		src = derefValueOf(src.Elem())
	}
	if src.Type().Kind() != reflect.Struct {
		panic("error trying to decode a no-struct type")
	}
	fields := deepFieldsFrom(src)

	str := ""
	for i := 0; i < len(elem.dict); i += 2 {
		keyElem := elem.dict[i]
		valElem := elem.dict[i+1]

		keyOf := reflect.New(reflect.TypeOf(str)).Elem()
		key := d.decodeValue(keyElem, keyOf)

		if field, ok := fields[valueOf(key).String()]; ok {
			// Decode nil pointer
			if field.Type().Kind() == reflect.Pointer {
				if slices.Equal(valElem.body, []byte{110, 105, 108}) {
					field.SetZero()
					continue
				}
			}

			valOf := reflect.New(derefTypeOf(field.Type())).Elem()
			val := d.decodeValue(valElem, valOf)
			if val != nil {
				valOf = valueOf(val)
			}

			if keyOf.IsValid() && valOf.IsValid() {
				setValueNotPtr(field.Type(), valOf, func(out reflect.Value) {
					field.Set(out.Convert(field.Type()))
				})
			}
		}
	}

	return src.Interface()
}
