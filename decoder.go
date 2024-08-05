package goetf

import (
	"bytes"
	"encoding/binary"
	"io"
	"maps"
	"math/big"
	"reflect"

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

	for !d.scan.eof() {
		elem, err := d.readNext()
		if err != nil {
			return err
		}

		parsed := d.decodeValue(elem, v)
		if parsed != nil {
			parsedOf := derefValueOf(parsed)
			if parsedOf.IsValid() {
				vOf := derefValueOf(v)

				// Set map keys
				if vOf.Type().Kind() == reflect.Map && parsedOf.Type().Kind() == reflect.Map {
					keys := parsedOf.MapKeys()
					for _, key := range keys {
						mValOf := parsedOf.MapIndex(key)
						if key.IsValid() && mValOf.IsValid() {
							vOf.SetMapIndex(key, mValOf)
						}
					}

					return nil
				}

				vOf.Set(parsedOf)
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
	if !vOf.IsValid() || !derefValueOf(v).IsValid() {
		return nil
	}
	kind = derefValueOf(vOf).Type().Kind()

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
			if kind == reflect.Struct {
				switch derefValueOf(vOf).Type() {
				default:
					return d.decodeStruct(elem, vOf)
				case typeOfBigInt:
					return vOf.Interface().(big.Int)
				}
			}

			if kind == reflect.Interface {
				return d.decodeAnyMap(elem, vOf)
			}

			return d.decodeMap(elem, vOf)
		}
	}

	return nil
}

func (d *Decoder) decodeTuple(elem *binaryElement, src reflect.Value) any {
	if src.Type().Kind() == reflect.Pointer {
		src = derefValueOf(src)
	}
	if src.Type().Kind() != reflect.Slice {
		panic("error trying to decode a no-slice type")
	}

	length := src.Len()
	for i, item := range elem.items {
		if length > 0 {
			srcElem := src.Index(i)

			parsed := d.decodeValue(item, srcElem)
			if parsed != nil {
				srcElem.Set(valueOf(parsed))
			}

		} else {
			newElem := reflect.New(derefTypeOf(src.Type().Elem())).Elem()
			if newElem.IsValid() {
				parsed := d.decodeValue(item, newElem)
				parsedOf := valueOf(parsed)
				if src.Type().Elem().Kind() == reflect.Pointer {
					ptr := reflect.New(parsedOf.Type())
					ptr.Elem().Set(parsedOf)
					src = reflect.Append(src, ptr)
				} else {
					src = reflect.Append(src, valueOf(parsed))
				}
			}
		}
	}

	return src.Interface()
}

func (d *Decoder) decodeAnyTuple(elem *binaryElement, src reflect.Value) any {
	if src.Type().Kind() == reflect.Pointer {
		src = derefValueOf(src)
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
		src = derefValueOf(src)
	}
	if src.Type().Kind() != reflect.Array {
		panic("error trying to decode a no-array type")
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
		arrElem := derefValueOf(arr.Index(i))
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

func (d *Decoder) decodeAnyList(elem *binaryElement, src reflect.Value) any {
	if src.Type().Kind() == reflect.Pointer {
		src = derefValueOf(src)
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
		src = derefValueOf(src)
	}
	if src.Type().Kind() != reflect.Map {
		panic("error trying to decode a no-map type")
	}

	mapType := reflect.MapOf(src.Type().Key(), src.Type().Elem())
	m := reflect.MakeMap(mapType)

	for i := 0; i < len(elem.dict)-1; i += 2 {
		keyElem := elem.dict[i]
		valElem := elem.dict[i+1]

		keyOf := reflect.New(src.Type().Key()).Elem()
		key := d.decodeValue(keyElem, keyOf)

		valOf := reflect.New(src.Type().Elem()).Elem()
		value := d.decodeValue(valElem, valOf)

		if keyOf.IsValid() && valOf.IsValid() {
			if key != nil && value != nil {
				m.SetMapIndex(valueOf(key), valueOf(value))
			} else {
				m.SetMapIndex((keyOf), (valOf))
			}
		}
	}

	return m.Interface()
}

func (d *Decoder) decodeAnyMap(elem *binaryElement, src reflect.Value) any {
	if src.Type().Kind() == reflect.Pointer {
		src = derefValueOf(src)
	}
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
				m.SetMapIndex(valueOf(key), valueOf(val))
			} else {
				m.SetMapIndex((keyOf), (valOf))
			}
		}
	}

	return m.Interface()
}

func (d *Decoder) decodeStruct(elem *binaryElement, src reflect.Value) any {
	if src.Type().Kind() == reflect.Pointer {
		src = derefValueOf(src)
	}
	if src.Type().Kind() != reflect.Struct {
		panic("error trying to decode a no-struct type")
	}
	fields := d.parseFieldsFrom(src)

	for i := 0; i < len(elem.dict)-1; i += 2 {
		keyElem := elem.dict[i]
		valElem := elem.dict[i+1]

		keyOf := reflect.New(reflect.TypeOf("")).Elem()
		key := d.decodeValue(keyElem, keyOf).(string)

		if field, ok := fields[key]; ok {
			valOf := reflect.New(derefTypeOf(field.Type())).Elem()
			val := d.decodeValue(valElem, valOf)

			if keyOf.IsValid() && valOf.IsValid() {
				if val != nil {
					if field.Type().Kind() == reflect.Pointer {
						// making the value addresable
						parsed := reflect.New(reflect.TypeOf(val))
						parsed.Elem().Set(valueOf(val))
						field.Set(parsed)
					} else {
						field.Set(valueOf(val))
					}
				} else {
					if field.Type().Kind() == reflect.Pointer {
						// making the value addresable
						parsed := reflect.New(valOf.Type())
						parsed.Elem().Set(valOf)
						field.Set(parsed)
					} else {
						field.Set(valOf)
					}
				}
			}
		}
	}

	return src.Interface()
}

func (d *Decoder) parseFieldsFrom(src reflect.Value) map[string]reflect.Value {
	if src.Type().Kind() != reflect.Struct {
		panic("error trying to decode a no-struct type")
	}

	result := map[string]reflect.Value{}
	for i := 0; i < src.NumField(); i++ {
		fval := src.Field(i)
		ftyp := src.Type().Field(i)
		tag := ftyp.Tag.Get("etf")

		if ftyp.Anonymous {
			m := d.parseFieldsFrom(fval)
			maps.Copy(result, m)
		} else {
			if tag != "" {
				result[tag] = fval
			} else {
				result[ftyp.Name] = fval
			}
		}
	}

	return result
}
