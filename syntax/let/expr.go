package let_parser

import (
	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type Ident = literal_parser.Ident

type LetExpr struct {
	Name  Ident
	Value ast.Expr
}

func NewLetExpr(name Ident, value ast.Expr) *LetExpr {
	return &LetExpr{Name: name, Value: value}
}

func (l *LetExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	evaluated, err := l.Value.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}
	ctx.SetValue(string(l.Name), evaluated)
	return value.NewTypeNil(), nil
}

func (*LetExpr) Type(ast.Ctx) string {
	return "let"
}

func (l *LetExpr) PrintGO(ctx ast.Ctx) (string, error) {
	var (
		printed string
		err     error
	)
	if printable, ok := l.Value.(interface {
		PrintGOAssign(ast.Ctx, string) (string, error)
	}); ok {
		printed, err = printable.PrintGOAssign(ctx, string(l.Name))
	} else {
		printed, err = l.Value.PrintGO(ctx)
		if err == nil {
			printed = string(l.Name) + " := " + printed
		}
	}
	if err != nil {
		return "", err
	}

	if typed, ok := l.Value.(interface {
		ValueType(ast.Ctx) ast.TypeRef
	}); ok {
		ref := typed.ValueType(ctx)
		if !ref.IsZero() {
			ctx.SetValue(
				string(l.Name),
				value.NewTypeWithExplicit(nil, ref.String()),
			)
		}
	}
	return printed, nil
}
