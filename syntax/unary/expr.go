package unary_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/token"
	"github.com/fobus89/dsl/value"
)

type UnaryExpr struct {
	op    token.TokenType
	count int
	expr  ast.Expr
}

func NewUnaryExpr(op token.TokenType, expr ast.Expr, count int) *UnaryExpr {
	return &UnaryExpr{
		op:    op,
		count: count,
		expr:  expr,
	}
}

func (u *UnaryExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	val, err := u.expr.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	switch u.op {
	case token.BANG:
		if u.count%2 == 0 {
			return value.NewType(val.UnsafeCastBool()), nil
		}

		return value.NewType(!val.UnsafeCastBool()), nil
	case token.MINUS:
		if !val.IsNumber() {
			return value.NewTypeNil(), fmt.Errorf("operator %q is not supported for %s", u.op, val.Typeof())
		}

		return value.NewType(-val.UnsafeCastFloat64()), nil
	case token.PLUS:
		if !val.IsNumber() {
			return value.NewTypeNil(), fmt.Errorf("operator %q is not supported for %s", u.op, val.Typeof())
		}

		return value.NewType(val.UnsafeCastFloat64()), nil
	}

	return value.NewTypeNil(), fmt.Errorf("operator %q is not supported", u.op)
}

func (*UnaryExpr) Type(_ ast.Ctx) string {
	return "unary"
}
