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

func (b *BinaryExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	leftVal, err := b.Left.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	rightVal, err := b.Right.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	if leftVal.IsString() || rightVal.IsString() {
		switch b.Op {
		case token.PLUS:
			return value.NewType(leftVal.UnsafeCastString() + rightVal.UnsafeCastString()), nil
		}
	}

	if leftVal.IsNumber() && rightVal.IsNumber() {
		left := leftVal.UnsafeCastFloat64()
		right := rightVal.UnsafeCastFloat64()

		switch b.Op {
		case token.PLUS:
			return value.NewType(left + right), nil
		case token.MINUS:
			return value.NewType(left - right), nil
		case token.STAR:
			return value.NewType(left * right), nil
		case token.SLASH:
			return value.NewType(left / right), nil
		case token.PERCENT:
			return value.NewType(math.Mod(left, right)), nil
		}
	}

	return value.NewTypeNil(), fmt.Errorf("operator %q is not supported for %s and %s", b.Op, leftVal.Typeof(), rightVal.Typeof())
}

func (b *BinaryExpr) Type(_ ast.Ctx) string {
	return "binary"
}
