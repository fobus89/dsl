package assignment_parser

import (
	"fmt"

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
	current, ok := ctx.GetValue(string(a.ident))
	if !ok {
		return value.NewTypeNil(), fmt.Errorf(
			"cannot assign to undeclared value %s",
			a.ident,
		)
	}
	if err := ctx.AssignValue(string(a.ident), current); err != nil {
		return value.NewTypeNil(), err
	}

	v, err := a.expr.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	if err := ctx.AssignValue(string(a.ident), v); err != nil {
		return value.NewTypeNil(), err
	}

	return value.NewTypeNil(), nil
}

func (*AssignmentExpr) Type(ctx ast.Ctx) string {
	return "assignment"
}

func (a *AssignmentExpr) PrintGO(ctx ast.Ctx) (string, error) {
	current, ok := ctx.GetValue(string(a.ident))
	if !ok {
		return "", fmt.Errorf(
			"cannot assign to undeclared value %s",
			a.ident,
		)
	}
	if err := ctx.AssignValue(string(a.ident), current); err != nil {
		return "", err
	}

	expr, err := a.expr.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	return string(a.ident) + " = " + expr, nil
}
