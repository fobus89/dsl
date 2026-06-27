// Package value provides runtime value representation and type operations
// for the foo_lang interpreter.
package value

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
	Value any
}

func NewTypeNil() Type {
	return Type{}
}

func NewType(v any) Type {
	return Type{
		Value: v,
	}
}

func NewTypeWithExplicit(v any, explicitType string) Type {
	return Type{
		Value: v,
	}
}

func (t Type) Any() any {
	return t.Value
}

func (t Type) Typeof() string {
	switch t.Value.(type) {
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
	}

	return ""
}

// type Type interface{}
//
// type IntType struct{}
//
// type ArrayType struct {
//     Size int
//     Elem Type
// }
//
// type SliceType struct {
//     Elem Type
// }
