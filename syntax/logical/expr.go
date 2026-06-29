package logical_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/token"
	"github.com/fobus89/dsl/value"
)

type LogicalExpr struct {
	left  ast.Expr
	op    token.TokenType
	right ast.Expr
}

func NewLogicalExpr(op token.TokenType, left, right ast.Expr) *LogicalExpr {
	return &LogicalExpr{
		left:  left,
		op:    op,
		right: right,
	}
}

func (l *LogicalExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	leftVal, err := l.left.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	rightVal, err := l.right.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	switch l.op {
	case token.AMP_AMP, token.AND:
		return value.NewType(leftVal.UnsafeCastBool() && rightVal.UnsafeCastBool()), nil
	case token.PIPE_PIPE, token.OR:
		return value.NewType(leftVal.UnsafeCastBool() || rightVal.UnsafeCastBool()), nil
	}

	return value.NewTypeNil(), fmt.Errorf("operator %q is not supported", l.op)
}

func (*LogicalExpr) Type(_ ast.Ctx) string {
	return "logical"
}
