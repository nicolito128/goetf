package goetf

import "reflect"

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
	switch vOf.Type().Kind() {
	case reflect.Pointer:
		return derefValueOf(vOf.Elem())
	default:
		return vOf
	}
}
