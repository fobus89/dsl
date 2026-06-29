package all_parser

import (
	"fmt"
	"reflect"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/value"
)

type AllExpr struct {
	left  ast.Expr
	right ast.Expr
}

func NewAllExpr(left, right ast.Expr) *AllExpr {
	return &AllExpr{
		left:  left,
		right: right,
	}
}

func (a *AllExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	leftVal, err := a.left.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	rightVal, err := a.right.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	if leftVal.IsPrimitiveSlice() && rightVal.IsPrimitiveSlice() {
		if leftVal.Len() > rightVal.Len() || leftVal.Typeof() != rightVal.Typeof() {
			return value.NewTypeNil(), nil
		}

		switch lv := leftVal.Any().(type) {
		case []int:
			return foundValue(sliceAll(lv, rightVal.Any().([]int))), nil
		case []int8:
			return foundValue(sliceAll(lv, rightVal.Any().([]int8))), nil
		case []int16:
			return foundValue(sliceAll(lv, rightVal.Any().([]int16))), nil
		case []int32:
			return foundValue(sliceAll(lv, rightVal.Any().([]int32))), nil
		case []int64:
			return foundValue(sliceAll(lv, rightVal.Any().([]int64))), nil
		case []uint:
			return foundValue(sliceAll(lv, rightVal.Any().([]uint))), nil
		case []uint8:
			return foundValue(sliceAll(lv, rightVal.Any().([]uint8))), nil
		case []uint16:
			return foundValue(sliceAll(lv, rightVal.Any().([]uint16))), nil
		case []uint32:
			return foundValue(sliceAll(lv, rightVal.Any().([]uint32))), nil
		case []uint64:
			return foundValue(sliceAll(lv, rightVal.Any().([]uint64))), nil
		case []float32:
			return foundValue(sliceAll(lv, rightVal.Any().([]float32))), nil
		case []float64:
			return foundValue(sliceAll(lv, rightVal.Any().([]float64))), nil
		}

		return value.NewTypeNil(), nil
	}

	if leftVal.IsMapSlice() && rightVal.IsMapSlice() {
		left, _ := leftVal.MapSlice()
		right, _ := rightVal.MapSlice()
		return foundValue(mapSliceAllSlice(left, right)), nil
	}

	if leftVal.IsMapSlice() && rightVal.IsMap() {
		left, _ := leftVal.MapSlice()
		right, _ := rightVal.Map()
		return foundValue(mapSliceAll(left, right)), nil
	}

	if leftVal.IsMap() && rightVal.IsMapSlice() {
		left, _ := leftVal.Map()
		right, _ := rightVal.MapSlice()
		return foundValue(mapAllSlice(left, right)), nil
	}

	if leftVal.IsMap() && rightVal.IsMap() {
		left, _ := leftVal.Map()
		right, _ := rightVal.Map()
		return foundValue(mapAll(left, right)), nil
	}

	if leftVal.IsPrimitive() && rightVal.IsMap() {
		right, _ := rightVal.Map()
		return foundValue(valAllMap(leftVal.Any(), right)), nil
	}

	if leftVal.IsMap() && rightVal.IsPrimitive() {
		left, _ := leftVal.Map()
		return foundValue(valAllMap(rightVal.Any(), left)), nil
	}

	if leftVal.IsPrimitive() && rightVal.IsPrimitiveSlice() {
		return foundValue(valAllSlice(leftVal.Any(), rightVal.Any())), nil
	}

	if leftVal.IsPrimitiveSlice() && rightVal.IsPrimitive() {
		return foundValue(sliceAllVal(leftVal.Any(), rightVal.Any())), nil
	}

	if leftVal.IsPrimitive() && rightVal.IsPrimitive() {
		return foundValue(valAnyVal(leftVal.Any(), rightVal.Any())), nil
	}

	return value.NewTypeNil(), fmt.Errorf("operator all is not supported for %s and %s", leftVal.Typeof(), rightVal.Typeof())
}

func (*AllExpr) Type(_ ast.Ctx) string {
	return "all"
}

func foundValue(v any, ok bool) value.Type {
	if !ok {
		return value.NewTypeNil()
	}

	return value.NewType(v)
}

func valAnyVal(val1, val2 any) (any, bool) {
	if equalAny(val1, val2) {
		return val1, true
	}

	return nil, false
}

func valAllSlice(val any, slice any) (any, bool) {
	v := reflect.ValueOf(slice)
	for i := 0; i < v.Len(); i++ {
		if equalAny(val, v.Index(i).Interface()) {
			return val, true
		}
	}

	return nil, false
}

func sliceAllVal(slice any, val any) (any, bool) {
	v := reflect.ValueOf(slice)
	for i := 0; i < v.Len(); i++ {
		if !equalAny(v.Index(i).Interface(), val) {
			return nil, false
		}
	}

	return slice, true
}

func sliceAll[T comparable](left, right []T) (any, bool) {
	if len(left) == 0 {
		return left, true
	}

	set := make(map[T]struct{}, len(right))

	for _, v := range right {
		set[v] = struct{}{}
	}

	for _, v := range left {
		if _, ok := set[v]; !ok {
			return nil, false
		}
	}

	return left, true
}

func mapAll(left, right map[string]any) (any, bool) {
	for k, lv := range left {
		rv, ok := right[k]
		if !ok {
			return nil, false
		}

		if !equalAny(lv, rv) {
			return nil, false
		}
	}

	return right, true
}

func valAllMap(val any, m map[string]any) (any, bool) {
	for _, v := range m {
		if !equalAny(val, v) {
			return nil, false
		}
	}

	return val, true
}

func mapAllSlice(left map[string]any, right []map[string]any) (any, bool) {
	var out []map[string]any

	for _, m := range right {
		if v, ok := mapAll(left, m); ok {
			out = append(out, v.(map[string]any))
		}
	}

	if len(out) == 0 {
		return nil, false
	}

	return out, true
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

func mapSliceAll(left []map[string]any, right map[string]any) (any, bool) {
	for _, m := range left {
		if _, ok := mapAll(m, right); !ok {
			return nil, false
		}
	}

	return right, true
}

func mapSliceAllSlice(left, right []map[string]any) (any, bool) {
	var out []map[string]any

	for _, l := range left {
		found := false

		for _, r := range right {
			if v, ok := mapAll(l, r); ok {
				out = append(out, v.(map[string]any))
				found = true
				break
			}
		}

		if !found {
			return nil, false
		}
	}

	return out, true
}
