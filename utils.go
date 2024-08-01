package goetf

import (
	"reflect"
)

func valueOf(value any) reflect.Value {
	switch v := value.(type) {
	case reflect.Value:
		return v
	default:
		return reflect.ValueOf(value)
	}
}

func deepTypeOf(value reflect.Type) reflect.Type {
	var kind reflect.Type

	switch value.Kind() {
	case reflect.Pointer:
		kind = deepTypeOf(value.Elem())

	case reflect.Array:
		kind = deepTypeOf(value.Elem())

	case reflect.Slice:
		kind = deepTypeOf(value.Elem())

	default:
		kind = value
	}

	return kind
}
