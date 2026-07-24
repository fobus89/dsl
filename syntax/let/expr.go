package let_parser

import (
	"fmt"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type Ident = literal_parser.Ident

type LetExpr struct {
	Name  Ident
	Names []Ident
	Value ast.Expr
}

func NewLetExpr(name Ident, value ast.Expr) *LetExpr {
	return &LetExpr{Name: name, Value: value}
}

func NewTupleLetExpr(names []Ident, value ast.Expr) *LetExpr {
	return &LetExpr{Names: names, Value: value}
}

func (l *LetExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	evaluated, err := l.Value.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}
	if len(l.Names) != 0 {
		tuple, ok := evaluated.Any().(ast.TupleValue)
		if !ok {
			return value.NewTypeNil(), fmt.Errorf(
				"cannot destructure %s as tuple",
				evaluated.TypeName(),
			)
		}
		if len(tuple.Elements) != len(l.Names) {
			return value.NewTypeNil(), fmt.Errorf(
				"tuple has %d elements but destructuring expects %d",
				len(tuple.Elements),
				len(l.Names),
			)
		}
		for index, name := range l.Names {
			if name != "_" {
				ctx.SetValue(string(name), tuple.Elements[index])
			}
		}
		return value.NewTypeNil(), nil
	}
	ctx.SetValue(string(l.Name), evaluated)
	return value.NewTypeNil(), nil
}

func (*LetExpr) Type(ast.Ctx) string {
	return "let"
}

func (l *LetExpr) PrintGO(ctx ast.Ctx) (string, error) {
	if len(l.Names) != 0 {
		return l.printGOTupleDestructure(ctx)
	}
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

func (l *LetExpr) printGOTupleDestructure(
	ctx ast.Ctx,
) (string, error) {
	typed, ok := l.Value.(interface {
		ValueType(ast.Ctx) ast.TypeRef
	})
	if !ok {
		return "", fmt.Errorf(
			"cannot infer tuple destructuring type",
		)
	}
	ref := typed.ValueType(ctx)
	if ref.Kind != ast.TupleTypeRef {
		return "", fmt.Errorf(
			"cannot destructure %s as tuple",
			ref.String(),
		)
	}
	if len(ref.Elems) != len(l.Names) {
		return "", fmt.Errorf(
			"tuple has %d elements but destructuring expects %d",
			len(ref.Elems),
			len(l.Names),
		)
	}
	printed, err := l.Value.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	names := make([]string, 0, len(l.Names))
	resultTypes := make([]string, 0, len(l.Names))
	results := make([]string, 0, len(l.Names))
	for index, name := range l.Names {
		names = append(names, string(name))
		resultTypes = append(
			resultTypes,
			ref.Elems[index].GoString(ctx),
		)
		results = append(
			results,
			fmt.Sprintf("__tuple.V%d", index),
		)
		if name != "_" {
			ctx.SetValue(
				string(name),
				value.NewTypeWithExplicit(
					nil,
					ref.Elems[index].String(),
				),
			)
		}
	}

	return strings.Join(names, ", ") +
		" := func() (" +
		strings.Join(resultTypes, ", ") +
		") {\n\t__tuple := " + printed +
		"\n\treturn " + strings.Join(results, ", ") +
		"\n}()", nil
}
