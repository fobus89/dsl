package binary_parser

import (
	"fmt"
	"math"
	"reflect"

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

func (b *BinaryExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	leftVal, err := b.Left.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	rightVal, err := b.Right.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	switch b.Op {
	case token.AMP_AMP, token.AND:
		return value.NewType(leftVal.UnsafeCastBool() && rightVal.UnsafeCastBool()), nil
	case token.PIPE_PIPE, token.OR:
		return value.NewType(leftVal.UnsafeCastBool() || rightVal.UnsafeCastBool()), nil
	}

	if (leftVal.IsNumber() && rightVal.IsNumber()) ||
		(leftVal.IsBool() || rightVal.IsBool()) {

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
				return value.NewType(equalValues(leftVal.Any(), rightVal.Any())), nil
			case token.BANG_EQ:
				return value.NewType(!equalValues(leftVal.Any(), rightVal.Any())), nil
			}
		}
	}

	switch b.Op {
	case token.EQ_EQ:
		return value.NewType(equalValues(leftVal.Any(), rightVal.Any())), nil
	case token.BANG_EQ:
		return value.NewType(!equalValues(leftVal.Any(), rightVal.Any())), nil
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

func equalValues(left, right any) bool {
	leftNumber, leftNumberOK := castNumber(left)
	rightNumber, rightNumberOK := castNumber(right)
	if leftNumberOK && rightNumberOK {
		if math.IsNaN(leftNumber) && math.IsNaN(rightNumber) {
			return true
		}

		return leftNumber == rightNumber
	}

	leftFloat, leftIsFloat := left.(float64)
	rightFloat, rightIsFloat := right.(float64)
	if leftIsFloat && rightIsFloat && math.IsNaN(leftFloat) && math.IsNaN(rightFloat) {
		return true
	}

	return reflect.DeepEqual(left, right)
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
