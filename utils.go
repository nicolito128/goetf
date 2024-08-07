package goetf

import (
	"maps"
	"reflect"
)

// valueOf ensures that reflect.ValueOf(v) is not used on another reflect.Value.
func valueOf(v any) reflect.Value {
	switch v := v.(type) {
	case reflect.Value:
		return v
	default:
		return reflect.ValueOf(v)
	}
}

// typeOf ensures that reflect.TypeOf(v) is not used on another reflect.Type.
func typeOf(v any) reflect.Type {
	switch v := v.(type) {
	case reflect.Type:
		return v
	default:
		return reflect.TypeOf(v)
	}
}

// derefValueOf recursively dereferences all pointers of a value by calling its method v.Elem().
func derefValueOf(v any) reflect.Value {
	vOf := valueOf(v)
	if vOf.IsValid() {
		switch vOf.Type().Kind() {
		case reflect.Pointer:
			return derefValueOf(vOf.Elem())
		default:
			return vOf
		}
	}

	return vOf
}

// derefTypeOf recursively dereferences all pointers of a type by calling its method v.Elem().
func derefTypeOf(v any) reflect.Type {
	vOf := typeOf(v)
	switch vOf.Kind() {
	case reflect.Pointer:
		return derefTypeOf(vOf.Elem())
	default:
		return vOf
	}
}

// toLittleEndian flips a BigEndian slice.
func toLittleEndian(b []byte) {
	for i := 0; i < len(b)/2; i++ {
		b[i], b[len(b)-i-1] = b[len(b)-i-1], b[i]
	}
}

func setValueNotPtr(dist reflect.Type, element reflect.Value, handleSet func(reflect.Value)) {
	if dist.Kind() == reflect.Pointer {
		ptr := reflect.New(element.Type())
		ptr.Elem().Set(element)
		handleSet(ptr)
	} else {
		handleSet(element)
	}
}

// deepFieldsFrom filters all the public fields from the src struct.
//
// This function enters those embedded parameters that do not have the "etf" tag.
func deepFieldsFrom(src reflect.Value) map[string]reflect.Value {
	if src.Type().Kind() != reflect.Struct {
		panic("error trying to decode a no-struct type")
	}

	result := map[string]reflect.Value{}
	for i := 0; i < src.NumField(); i++ {
		fval := src.Field(i)
		ftyp := src.Type().Field(i)
		tag := ftyp.Tag.Get("etf")

		if ftyp.Anonymous {
			m := deepFieldsFrom(fval)
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

// TagString returns the string representation for an external format tag.
func TagString(ett ExternalTagType) string {
	if tag, ok := tagNames[ett]; ok {
		return tag
	}

	return ""
}

// IsValidEtt validates whether the byte argument is an external format flag.
func IsValidEtt(b byte) bool {
	return TagString(b) != ""
}

var tagNames = map[ExternalTagType]string{
	EttAtom:          "ATOM_EXT",
	EttAtomUTF8:      "ATOM_UTF8_EXT",
	EttBinary:        "BINARY_EXT",
	EttBitBinary:     "BIT_BINARY_EXT",
	EttAtomCacheRef:  "ATOM_CACHE_REF",
	EttExport:        "EXPORT_EXT",
	EttFloat:         "FLOAT_EXT",
	EttFun:           "FUN_EXT",
	EttInteger:       "INTEGER_EXT",
	EttLargeBig:      "LARGE_BIG_EXT",
	EttLargeTuple:    "LARGE_TUPLE_EXT",
	EttList:          "LIST_EXT",
	EttNewFloat:      "NEW_FLOAT_EXT",
	EttNewFun:        "NEW_FUN_EXT",
	EttNewReference:  "NEW_REFERENCE_EXT",
	EttNil:           "NIL_EXT",
	EttPid:           "PID_EXT",
	EttPort:          "PORT_EXT",
	EttRef:           "REFERENCE_EXT",
	EttSmallAtom:     "SMALL_ATOM_EXT",
	EttSmallAtomUTF8: "SMALL_ATOM_UTF8_EXT",
	EttSmallBig:      "SMALL_BIG_EXT",
	EttSmallInteger:  "SMALL_INTEGER_EXT",
	EttSmallTuple:    "SMALL_TUPLE_EXT",
	EttMap:           "MAP_EXT",
	EttString:        "STRING_EXT",
	EttV4Port:        "V4_PORT_EXT",
	EttLocal:         "LOCAL_EXT",
}
