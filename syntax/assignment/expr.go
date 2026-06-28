package assignment_parser

import (
	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type Ident = literal_parser.Ident

type AssignmentExpr struct {
	ident Ident
	expr  ast.Expr
}

func NewAssignmentExprExpr(name Ident, expr ast.Expr) *AssignmentExpr {
	return &AssignmentExpr{
		ident: name,
		expr:  expr,
	}
}

func (a *AssignmentExpr) Eval(ctx ast.Ctx) (value.Type, error) {

	v, err := a.expr.Eval(ctx)
	{
		if err != nil {
			return value.NewTypeNil(), err
		}
	}

	ctx.SetValue(string(a.ident), v)

	return value.NewTypeNil(), nil
}

func (*AssignmentExpr) Type(ctx ast.Ctx) string {
	return "assignment"
}
