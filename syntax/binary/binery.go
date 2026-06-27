package binary_parser

import (
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
		}
	}

	return value.NewTypeNil(), nil
}

func (b *BinaryExpr) Type(_ ast.Ctx) string {
	return "Binary"
}
