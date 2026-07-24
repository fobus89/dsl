package tuple_parser

import (
	"fmt"
	"strings"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/value"
)

type TupleExpr struct {
	Elements []ast.Expr
}

func NewTupleExpr(elements []ast.Expr) *TupleExpr {
	return &TupleExpr{Elements: elements}
}

func (t *TupleExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	elements := make([]value.Type, 0, len(t.Elements))
	for _, expr := range t.Elements {
		evaluated, err := expr.Eval(ctx)
		if err != nil {
			return value.NewTypeNil(), err
		}
		elements = append(elements, evaluated)
	}
	ref := t.ValueType(ctx)
	return value.NewTypeWithExplicit(
		ast.TupleValue{Elements: elements},
		ref.String(),
	), nil
}

func (*TupleExpr) Type(ast.Ctx) string {
	return "tuple"
}

func (t *TupleExpr) ValueType(ctx ast.Ctx) ast.TypeRef {
	ref := ast.TypeRef{Kind: ast.TupleTypeRef}
	for _, expr := range t.Elements {
		elem := tupleExprType(ctx, expr)
		if elem.IsZero() {
			return ast.TypeRef{}
		}
		ref.Elems = append(ref.Elems, elem)
	}
	return ref
}

func (t *TupleExpr) PrintGO(ctx ast.Ctx) (string, error) {
	ref := t.ValueType(ctx)
	if ref.IsZero() {
		return "", fmt.Errorf("cannot infer tuple element types")
	}
	elements := make([]string, 0, len(t.Elements))
	for index, expr := range t.Elements {
		printed, err := expr.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		elements = append(
			elements,
			fmt.Sprintf("V%d: %s", index, printed),
		)
	}
	return ref.GoString(ctx) +
		"{" + strings.Join(elements, ", ") + "}", nil
}

func tupleExprType(ctx ast.Ctx, expr ast.Expr) ast.TypeRef {
	if typed, ok := expr.(interface {
		ValueType(ast.Ctx) ast.TypeRef
	}); ok {
		return typed.ValueType(ctx)
	}
	return ast.TypeRef{}
}
