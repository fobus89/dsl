package let_parser

import (
	"fmt"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type Ident = literal_parser.Ident

type ValueDecl struct {
	Name     Ident
	Names    []Ident
	Value    ast.Expr
	Constant bool
}

// LetExpr is kept as an alias for callers using the old syntax API.
type LetExpr = ValueDecl

func NewValueDecl(
	name Ident,
	value ast.Expr,
	constant bool,
) *ValueDecl {
	return &ValueDecl{
		Name:     name,
		Value:    value,
		Constant: constant,
	}
}

func NewLetExpr(name Ident, value ast.Expr) *ValueDecl {
	return NewValueDecl(name, value, false)
}

func NewConstExpr(name Ident, value ast.Expr) *ValueDecl {
	return NewValueDecl(name, value, true)
}

func NewTupleLetExpr(names []Ident, value ast.Expr) *ValueDecl {
	return &ValueDecl{Names: names, Value: value}
}

func (l *ValueDecl) Eval(ctx ast.Ctx) (value.Type, error) {
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
				ctx.DeclareValue(
					string(name),
					tuple.Elements[index],
					false,
				)
			}
		}
		return value.NewTypeNil(), nil
	}
	ctx.DeclareValue(string(l.Name), evaluated, l.Constant)
	return value.NewTypeNil(), nil
}

func (l *ValueDecl) Type(ast.Ctx) string {
	if l.Constant {
		return "const"
	}
	return "let"
}

func (l *ValueDecl) PrintGO(ctx ast.Ctx) (string, error) {
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
		if l.Constant {
			return "", fmt.Errorf(
				"const %s requires a constant initializer",
				l.Name,
			)
		}
		printed, err = printable.PrintGOAssign(ctx, string(l.Name))
	} else {
		printed, err = l.Value.PrintGO(ctx)
		if err == nil {
			operator := " := "
			if l.Constant {
				operator = " = "
				printed = "const " + string(l.Name) +
					operator + printed
			} else {
				printed = string(l.Name) + operator + printed
			}
		}
	}
	if err != nil {
		return "", err
	}

	declared := value.NewTypeNil()
	if typed, ok := l.Value.(interface {
		ValueType(ast.Ctx) ast.TypeRef
	}); ok {
		ref := typed.ValueType(ctx)
		if !ref.IsZero() {
			declared = value.NewTypeWithExplicit(nil, ref.String())
		}
	}
	ctx.DeclareValue(string(l.Name), declared, l.Constant)
	return printed, nil
}

func (l *ValueDecl) printGOTupleDestructure(
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
			ctx.DeclareValue(
				string(name),
				value.NewTypeWithExplicit(
					nil,
					ref.Elems[index].String(),
				),
				false,
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
