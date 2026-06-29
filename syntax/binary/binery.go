package binary_parser

import (
	"fmt"
	"math"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/token"
	"github.com/fobus89/dsl/value"
)

type BinaryExpr struct {
	Left  ast.Expr
	Op    token.TokenType
	Right ast.Expr
}

func NewBinaryExpr(op token.TokenType, left, right ast.Expr) *BinaryExpr {
	return &BinaryExpr{
		Left:  left,
		Op:    op,
		Right: right,
	}
}

func foundValue(v any, ok bool) value.Type {
	if !ok {
		return value.NewTypeNil()
	}

	return value.NewType(v)
}

func (b *BinaryExpr) Eval(ctx ast.Ctx) (value.Type, error) {

	leftExpr := b.Left
	rightExpr := b.Right

	leftVal, err := leftExpr.Eval(ctx)
	{
		if err != nil {
			return value.NewTypeNil(), err
		}
	}

	rightVal, err := rightExpr.Eval(ctx)
	{
		if err != nil {
			return value.NewTypeNil(), err
		}
	}

	switch b.Op {
	case token.ALL:

		if leftVal.IsPrimitiveSlice() && rightVal.IsPrimitiveSlice() {
			if leftVal.IsPrimitiveSlice() && rightVal.IsPrimitiveSlice() {

				if leftVal.Len() > rightVal.Len() {
					return value.NewTypeNil(), nil
				}

				if leftVal.Typeof() == rightVal.Typeof() {

					switch lv := leftVal.Any().(type) {
					case []int:
						rv := rightVal.Any().([]int)
						return foundValue(sliceAll(lv, rv)), nil

					case []int8:
						rv := rightVal.Any().([]int8)
						return foundValue(sliceAll(lv, rv)), nil
					case []int16:
						rv := rightVal.Any().([]int16)
						return foundValue(sliceAll(lv, rv)), nil
					case []int32:
						rv := rightVal.Any().([]int32)
						return foundValue(sliceAll(lv, rv)), nil
					case []int64:
						rv := rightVal.Any().([]int64)
						return foundValue(sliceAll(lv, rv)), nil
					case []uint:
						rv := rightVal.Any().([]uint)
						return foundValue(sliceAll(lv, rv)), nil
					case []uint8:
						rv := rightVal.Any().([]uint8)
						return foundValue(sliceAll(lv, rv)), nil
					case []uint16:
						rv := rightVal.Any().([]uint16)
						return foundValue(sliceAll(lv, rv)), nil
					case []uint32:
						rv := rightVal.Any().([]uint32)
						return foundValue(sliceAll(lv, rv)), nil
					case []uint64:
						rv := rightVal.Any().([]uint64)
						return foundValue(sliceAll(lv, rv)), nil
					case []float32:
						rv := rightVal.Any().([]float32)
						return foundValue(sliceAll(lv, rv)), nil
					case []float64:
						rv := rightVal.Any().([]float64)
						return foundValue(sliceAll(lv, rv)), nil
					}

				}

				return value.NewTypeNil(), nil
			}
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

	case token.Any:
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

			if leftVal.Typeof() == rightVal.Typeof() {

				switch lv := leftVal.Any().(type) {
				case []int:
					rv := rightVal.Any().([]int)
					return foundValue(sliceAny(lv, rv)), nil

				case []int8:
					rv := rightVal.Any().([]int8)
					return foundValue(sliceAny(lv, rv)), nil
				case []int16:
					rv := rightVal.Any().([]int16)
					return foundValue(sliceAny(lv, rv)), nil
				case []int32:
					rv := rightVal.Any().([]int32)
					return foundValue(sliceAny(lv, rv)), nil
				case []int64:
					rv := rightVal.Any().([]int64)
					return foundValue(sliceAny(lv, rv)), nil
				case []uint:
					rv := rightVal.Any().([]uint)
					return foundValue(sliceAny(lv, rv)), nil
				case []uint8:
					rv := rightVal.Any().([]uint8)
					return foundValue(sliceAny(lv, rv)), nil
				case []uint16:
					rv := rightVal.Any().([]uint16)
					return foundValue(sliceAny(lv, rv)), nil
				case []uint32:
					rv := rightVal.Any().([]uint32)
					return foundValue(sliceAny(lv, rv)), nil
				case []uint64:
					rv := rightVal.Any().([]uint64)
					return foundValue(sliceAny(lv, rv)), nil
				case []float32:
					rv := rightVal.Any().([]float32)
					return foundValue(sliceAny(lv, rv)), nil
				case []float64:
					rv := rightVal.Any().([]float64)
					return foundValue(sliceAny(lv, rv)), nil
				}

			}

			return value.NewTypeNil(), nil
		}

	}

	if (leftVal.IsNumber() && rightVal.IsNumber()) ||
		(leftVal.IsBool() || rightVal.IsBool()) {

		// math op
		{
			l := leftVal.UnsafeCastFloat64()
			r := rightVal.UnsafeCastFloat64()

			switch b.Op {
			case token.PLUS:
				return value.NewType(l + r), nil

			case token.MINUS:
				return value.NewType(l - r), nil

			case token.STAR:
				return value.NewType(l * r), nil

			case token.SLASH:
				return value.NewType(l / r), nil

			case token.PERCENT:
				return value.NewType(math.Mod(l, r)), nil
			}
		}

		// logical op
		{
			l := leftVal.UnsafeCastBool()
			r := rightVal.UnsafeCastBool()
			switch b.Op {

			case token.AMP_AMP, token.AND:
				return value.NewType(l && r), nil

			case token.PIPE_PIPE, token.OR:
				return value.NewType(l || r), nil
			}
		}

		// comparison op
		{
			l := leftVal.UnsafeCastFloat64()
			r := rightVal.UnsafeCastFloat64()

			switch b.Op {
			case token.GT:
				return value.NewType(l > r), nil
			case token.LT:
				return value.NewType(l < r), nil
			case token.GT_EQ:
				return value.NewType(l >= r), nil
			case token.LT_EQ:
				return value.NewType(l <= r), nil
			case token.EQ_EQ:
				return value.NewType(l == r), nil
			case token.BANG_EQ:
				return value.NewType(l != r), nil
			}

		}
	}

	if leftVal.IsString() || rightVal.IsString() {
		switch b.Op {
		case token.PLUS:
			return value.NewType(leftVal.UnsafeCastString() + rightVal.UnsafeCastString()), nil
		case token.EQ_EQ:
			return value.NewType(leftVal.UnsafeCastString() == rightVal.UnsafeCastString()), nil
		case token.BANG_EQ:
			return value.NewType(leftVal.UnsafeCastString() != rightVal.UnsafeCastString()), nil
		}
	}

	return value.NewTypeNil(), fmt.Errorf("operator %q is not supported for %s and %s", b.Op, leftVal.Typeof(), rightVal.Typeof())
}

func (b *BinaryExpr) Type(_ ast.Ctx) string {
	return "Binary"
}
