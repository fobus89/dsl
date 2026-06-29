package comparison_parser

import (
	"fmt"
	"math"
	"reflect"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/token"
	"github.com/fobus89/dsl/value"
)

type ComparisonExpr struct {
	left  ast.Expr
	op    token.TokenType
	right ast.Expr
}

func NewComparisonExpr(op token.TokenType, left, right ast.Expr) *ComparisonExpr {
	return &ComparisonExpr{
		left:  left,
		op:    op,
		right: right,
	}
}

func (c *ComparisonExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	leftVal, err := c.left.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	rightVal, err := c.right.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	switch c.op {
	case token.EQ_EQ:
		return value.NewType(equalValues(leftVal.Any(), rightVal.Any())), nil
	case token.BANG_EQ:
		return value.NewType(!equalValues(leftVal.Any(), rightVal.Any())), nil
	}

	if leftVal.IsNumber() && rightVal.IsNumber() {
		left := leftVal.UnsafeCastFloat64()
		right := rightVal.UnsafeCastFloat64()

		switch c.op {
		case token.GT:
			return value.NewType(left > right), nil
		case token.LT:
			return value.NewType(left < right), nil
		case token.GT_EQ:
			return value.NewType(left >= right), nil
		case token.LT_EQ:
			return value.NewType(left <= right), nil
		}
	}

	return value.NewTypeNil(), fmt.Errorf("operator %q is not supported for %s and %s", c.op, leftVal.Typeof(), rightVal.Typeof())
}

func (*ComparisonExpr) Type(_ ast.Ctx) string {
	return "comparison"
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
