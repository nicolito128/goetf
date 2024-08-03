package goetf

import (
	"reflect"
)

func valueOf(v any) reflect.Value {
	switch v := v.(type) {
	case reflect.Value:
		return v
	default:
		return reflect.ValueOf(v)
	}
}

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

func TagString(ett ExternalTagType) string {
	if tag, ok := tagNames[ett]; ok {
		return tag
	}

	return ""
}

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
