package any_parser

import (
	"fmt"
	"reflect"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/value"
)

type AnyExpr struct {
	left  ast.Expr
	right ast.Expr
}

func NewAnyExpr(left, right ast.Expr) *AnyExpr {
	return &AnyExpr{
		left:  left,
		right: right,
	}
}

func (a *AnyExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	leftVal, err := a.left.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	rightVal, err := a.right.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	if leftVal.IsMapSlice() && rightVal.IsMapSlice() {
		left, _ := leftVal.MapSlice()
		right, _ := rightVal.MapSlice()
		return foundValue(mapSliceAnySlice(left, right)), nil
	}

	if leftVal.IsMapSlice() && rightVal.IsMap() {
		left, _ := leftVal.MapSlice()
		right, _ := rightVal.Map()
		return foundValue(mapSliceAny(left, right)), nil
	}

	if leftVal.IsMap() && rightVal.IsMapSlice() {
		left, _ := leftVal.Map()
		right, _ := rightVal.MapSlice()
		return foundValue(mapAnySlice(left, right)), nil
	}

	if leftVal.IsMap() && rightVal.IsMap() {
		left, _ := leftVal.Map()
		right, _ := rightVal.Map()
		return foundValue(mapAny(left, right)), nil
	}

	if leftVal.IsPrimitive() && rightVal.IsMap() {
		right, _ := rightVal.Map()
		return foundValue(valAnyMap(leftVal.Any(), right)), nil
	}

	if leftVal.IsMap() && rightVal.IsPrimitive() {
		left, _ := leftVal.Map()
		return foundValue(valAnyMap(rightVal.Any(), left)), nil
	}

	if leftVal.IsPrimitive() && rightVal.IsPrimitiveSlice() {
		return foundValue(valAnySlice(leftVal.Any(), rightVal.Any())), nil
	}

	if leftVal.IsPrimitiveSlice() && rightVal.IsPrimitive() {
		return foundValue(sliceAnyVal(leftVal.Any(), rightVal.Any())), nil
	}

	if leftVal.IsPrimitive() && rightVal.IsPrimitive() {
		return foundValue(valAnyVal(leftVal.Any(), rightVal.Any())), nil
	}

	if leftVal.IsPrimitiveSlice() && rightVal.IsPrimitiveSlice() {
		if leftVal.Typeof() != rightVal.Typeof() {
			return value.NewTypeNil(), nil
		}

		switch lv := leftVal.Any().(type) {
		case []int:
			return foundValue(sliceAny(lv, rightVal.Any().([]int))), nil
		case []int8:
			return foundValue(sliceAny(lv, rightVal.Any().([]int8))), nil
		case []int16:
			return foundValue(sliceAny(lv, rightVal.Any().([]int16))), nil
		case []int32:
			return foundValue(sliceAny(lv, rightVal.Any().([]int32))), nil
		case []int64:
			return foundValue(sliceAny(lv, rightVal.Any().([]int64))), nil
		case []uint:
			return foundValue(sliceAny(lv, rightVal.Any().([]uint))), nil
		case []uint8:
			return foundValue(sliceAny(lv, rightVal.Any().([]uint8))), nil
		case []uint16:
			return foundValue(sliceAny(lv, rightVal.Any().([]uint16))), nil
		case []uint32:
			return foundValue(sliceAny(lv, rightVal.Any().([]uint32))), nil
		case []uint64:
			return foundValue(sliceAny(lv, rightVal.Any().([]uint64))), nil
		case []float32:
			return foundValue(sliceAny(lv, rightVal.Any().([]float32))), nil
		case []float64:
			return foundValue(sliceAny(lv, rightVal.Any().([]float64))), nil
		}

		return value.NewTypeNil(), nil
	}

	return value.NewTypeNil(), fmt.Errorf("operator any is not supported for %s and %s", leftVal.Typeof(), rightVal.Typeof())
}

func (*AnyExpr) Type(_ ast.Ctx) string {
	return "any"
}

func foundValue(v any, ok bool) value.Type {
	if !ok {
		return value.NewTypeNil()
	}

	return value.NewType(v)
}

func mapAny(left, right map[string]any) (any, bool) {
	for k, lv := range left {
		rv, ok := right[k]
		if !ok {
			continue
		}

		if equalAny(lv, rv) {
			return right, true
		}
	}

	return nil, false
}

func valAnyMap(val any, m map[string]any) (any, bool) {
	for _, v := range m {
		if equalAny(val, v) {
			return v, true
		}
	}

	return nil, false
}

func valAnyVal(val1, val2 any) (any, bool) {
	if equalAny(val1, val2) {
		return val1, true
	}

	return nil, false
}

func valAnySlice(val any, slice any) (any, bool) {
	v := reflect.ValueOf(slice)
	for i := 0; i < v.Len(); i++ {
		if equalAny(val, v.Index(i).Interface()) {
			return val, true
		}
	}

	return nil, false
}

func equalAny(left, right any) bool {
	if reflect.DeepEqual(left, right) {
		return true
	}

	leftNumber, leftOK := castNumber(left)
	rightNumber, rightOK := castNumber(right)

	return leftOK && rightOK && leftNumber == rightNumber
}

func castNumber(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	}

	return 0, false
}

func sliceAnyVal(slice any, val any) (any, bool) {
	return valAnySlice(val, slice)
}

func mapAnySlice(left map[string]any, right []map[string]any) (any, bool) {
	for _, m := range right {
		if v, ok := mapAny(left, m); ok {
			return v, true
		}
	}

	return nil, false
}

func mapSliceAny(left []map[string]any, right map[string]any) (any, bool) {
	for _, m := range left {
		if v, ok := mapAny(m, right); ok {
			return v, true
		}
	}

	return nil, false
}

func mapSliceAnySlice(left, right []map[string]any) (any, bool) {
	for _, l := range left {
		for _, r := range right {
			if v, ok := mapAny(l, r); ok {
				return v, true
			}
		}
	}

	return nil, false
}

func sliceAny[T comparable](left, right []T) (any, bool) {
	if len(left) == 0 || len(right) == 0 {
		return nil, false
	}

	set := make(map[T]struct{}, len(right))

	for _, v := range right {
		set[v] = struct{}{}
	}

	for _, v := range left {
		if _, ok := set[v]; ok {
			return v, true
		}
	}

	return nil, false
}
