// Package value provides runtime value representation and type operations
// for the foo_lang interpreter.
package value

import (
	"reflect"
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

func Is[T any](v any) bool {
	_, ok := To[T](v)
	return ok
}

func To[T any](v any) (T, bool) {
	switch t := v.(type) {
	case T:
		return t, true
	}

	var none T

	return none, false
}

func Cast[T Number](v any) (T, bool) {
	switch t := v.(type) {
	case uint8:
		return T(t), true
	case uint16:
		return T(t), true
	case uint32:
		return T(t), true
	case uint64:
		return T(t), true
	case int8:
		return T(t), true
	case int16:
		return T(t), true
	case int32:
		return T(t), true
	case int64:
		return T(t), true
	case int:
		return T(t), true
	case uint:
		return T(t), true
	case float32:
		return T(t), true
	case float64:
		return T(t), true
	}

	return 0, false
}

type Type struct {
	value any
}

func NewTypeNil() Type {
	return Type{}
}

func NewType(v any) Type {
	return Type{
		value: v,
	}
}

func NewTypeWithExplicit(v any, explicitType string) Type {
	return Type{
		value: v,
	}
}

func (t Type) Any() any {
	return t.value
}

func (t Type) Map() (map[string]any, bool) {
	m, ok := t.value.(map[string]any)

	return m, ok
}

func (t Type) MapSlice() ([]map[string]any, bool) {
	m, ok := t.value.([]map[string]any)

	return m, ok
}

func (t Type) IsNil() bool {
	return t.value == nil
}

func (t Type) IsMap() bool {
	_, ok := t.value.(map[string]any)

	return ok
}

func (t Type) IsMapSlice() bool {
	ty := reflect.TypeOf(t.value)
	if ty == nil || ty.Kind() != reflect.Slice {
		return false
	}

	elem := ty.Elem()
	if elem.Kind() != reflect.Map {
		return false
	}

	return elem.Key().Kind() == reflect.String &&
		elem.Elem().Kind() == reflect.Interface
}

func (t Type) Len() int {
	v := reflect.ValueOf(t.value)

	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return v.Len()
	default:
		return 0
	}
}

func (t Type) SliceElement() int {

	v := reflect.ValueOf(t.value)

	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return v.Len()
	default:
		return 0
	}
}

func (t Type) IsPrimitive() bool {
	ty := reflect.TypeOf(t.value)

	switch ty.Kind() {
	case reflect.Bool,
		reflect.String,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64:
		return true
	}

	return false
}

func (t Type) IsPrimitiveSlice() bool {

	ty := reflect.TypeOf(t.value)

	if ty == nil || ty.Kind() != reflect.Slice {
		return false
	}

	switch ty.Elem().Kind() {
	case reflect.Bool,
		reflect.String,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64:
		return true
	}

	return false
}

func (t Type) IsSlice() bool {
	return reflect.TypeOf(t.value).Kind() == reflect.Slice
}

func (t Type) Typeof() string {
	switch t.value.(type) {
	case uint8:
		return "uint8"
	case uint16:
		return "uint16"
	case uint32:
		return "uint32"
	case uint64:
		return "uint64"
	case int8:
		return "int8"
	case int16:
		return "int16"
	case int32:
		return "int32"
	case int64:
		return "int64"
	case int:
		return "int"
	case uint:
		return "uint"
	case float32:
		return "float32"
	case float64:
		return "float64"
	case string:
		return "string"
	case bool:
		return "bool"
	case Undefined:
		return "undefined"
	case []int:
		return "[]int"
	case []int8:
		return "[]int8"
	case []int16:
		return "[]int16"
	case []int32:
		return "[]int32"
	case []int64:
		return "[]int64"
	case []uint:
		return "[]uint"
	case []uint8:
		return "[]uint8"
	case []uint16:
		return "[]uint16"
	case []uint32:
		return "[]uint32"
	case []uint64:
		return "[]uint64"
	case []float32:
		return "[]float32"
	case []float64:
		return "[]float64"
	case []string:
		return "[]string"
	case []bool:
		return "[]bool"
	case []any:
		return "[]any"
	case map[string]any:
		return "map[string]any"
	case []map[string]any:
		return "[]map[string]any"
	}

	return "unknow"
}
