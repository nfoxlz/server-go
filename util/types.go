// types
package util

import (
	"reflect"
)

func ConvertType(val reflect.Value, kind reflect.Kind) reflect.Value {
	switch kind {
	case reflect.Int:
		return reflect.ValueOf(int(val.Int()))
	case reflect.Int8:
		return reflect.ValueOf(int8(val.Int()))
	case reflect.Int16:
		return reflect.ValueOf(int16(val.Int()))
	case reflect.Int32:
		return reflect.ValueOf(int32(val.Int()))
	case reflect.Int64:
		return reflect.ValueOf(val.Int())
	case reflect.Uint:
		return reflect.ValueOf(uint(val.Uint()))
	case reflect.Uint8:
		return reflect.ValueOf(uint8(val.Uint()))
	case reflect.Uint16:
		return reflect.ValueOf(uint16(val.Uint()))
	case reflect.Uint32:
		return reflect.ValueOf(uint32(val.Uint()))
	case reflect.Uint64:
		return reflect.ValueOf(val.Uint())
	case reflect.Float32:
		return reflect.ValueOf(float32(val.Float()))
	case reflect.Float64:
		return reflect.ValueOf(val.Float())
	case reflect.String:
		return reflect.ValueOf(val.String())
	case reflect.Bool:
		return reflect.ValueOf(val.Bool())
	case reflect.Complex64:
		return reflect.ValueOf(complex64(val.Complex()))
	case reflect.Complex128:
		return reflect.ValueOf(val.Complex())
	default:
		return val
	}
}
