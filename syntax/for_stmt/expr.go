package forstmt_parser

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type Ident = literal_parser.Ident

type ForExpr struct {
	Ident     *Ident
	Iterable  ast.Expr
	Condition ast.Expr
	Body      []ast.Expr
}

func NewForExpr(
	ident *Ident,
	iterable ast.Expr,
	condition ast.Expr,
	body []ast.Expr,
) *ForExpr {
	return &ForExpr{
		Ident:     ident,
		Iterable:  iterable,
		Condition: condition,
		Body:      body,
	}
}

func (f *ForExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	var collected []any

	if f.Iterable != nil {
		iterable, err := f.Iterable.Eval(ctx)
		if err != nil {
			return value.NewTypeNil(), err
		}
		collected, err = f.evalIterable(ctx, iterable.Any())
		if err != nil {
			return value.NewTypeNil(), err
		}
		return f.collectedValue(ctx, collected), nil
	}

	for {
		if f.Condition != nil {
			condition, err := f.Condition.Eval(ctx)
			if err != nil {
				return value.NewTypeNil(), err
			}
			if !condition.UnsafeCastBool() {
				break
			}
		}

		err := evalBody(ctx, f.Body)
		if err == nil {
			continue
		}

		signal, ok := errors.AsType[ast.FlowSignal](err)
		if !ok {
			return value.NewTypeNil(), err
		}
		switch signal.Kind {
		case ast.BreakFlow:
			return f.collectedValue(ctx, collected), nil
		case ast.ContinueFlow:
			continue
		case ast.YieldFlow:
			collected = append(collected, signal.Value.Any())
		}
	}

	return f.collectedValue(ctx, collected), nil
}

func (f *ForExpr) collectedValue(
	ctx ast.Ctx,
	collected []any,
) value.Type {
	if ref := f.ValueType(ctx); !ref.IsZero() {
		return value.NewTypeWithExplicit(collected, ref.String())
	}
	return value.NewType(collected)
}

func (f *ForExpr) evalIterable(
	ctx ast.Ctx,
	iterable any,
) ([]any, error) {
	var collected []any

	if text, ok := iterable.(string); ok {
		for _, item := range text {
			ctx.SetValue(string(*f.Ident), value.NewType(item))
			stop, err := collectFlow(ctx, f.Body, &collected)
			if err != nil {
				return nil, err
			}
			if stop {
				break
			}
		}
		return collected, nil
	}

	reflected := reflect.ValueOf(iterable)
	if !reflected.IsValid() {
		return collected, nil
	}

	switch reflected.Kind() {
	case reflect.Slice, reflect.Array:
		for index := 0; index < reflected.Len(); index++ {
			ctx.SetValue(
				string(*f.Ident),
				value.NewType(reflected.Index(index).Interface()),
			)
			stop, err := collectFlow(ctx, f.Body, &collected)
			if err != nil {
				return nil, err
			}
			if stop {
				break
			}
		}
	case reflect.Map:
		iterator := reflected.MapRange()
		for iterator.Next() {
			ctx.SetValue(
				string(*f.Ident),
				value.NewType(iterator.Value().Interface()),
			)
			stop, err := collectFlow(ctx, f.Body, &collected)
			if err != nil {
				return nil, err
			}
			if stop {
				break
			}
		}
	default:
		return nil, fmt.Errorf("cannot iterate over %T", iterable)
	}

	return collected, nil
}

func collectFlow(
	ctx ast.Ctx,
	body []ast.Expr,
	collected *[]any,
) (bool, error) {
	err := evalBody(ctx, body)
	if err == nil {
		return false, nil
	}

	signal, ok := errors.AsType[ast.FlowSignal](err)
	if !ok {
		return false, err
	}
	switch signal.Kind {
	case ast.BreakFlow:
		return true, nil
	case ast.ContinueFlow:
		return false, nil
	case ast.YieldFlow:
		*collected = append(*collected, signal.Value.Any())
		return false, nil
	default:
		return false, err
	}
}

func evalBody(ctx ast.Ctx, body []ast.Expr) error {
	for _, expr := range body {
		if _, err := expr.Eval(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (*ForExpr) Type(ast.Ctx) string {
	return "for"
}

func (f *ForExpr) ValueType(ctx ast.Ctx) ast.TypeRef {
	ref, known, err := f.yieldCollectionType(ctx)
	if err != nil || !known {
		return ast.TypeRef{}
	}
	return ref
}

func (f *ForExpr) ChildExprs() []ast.Expr {
	return f.Body
}

func (f *ForExpr) PrintGO(ctx ast.Ctx) (string, error) {
	return f.printGO(flowContext{Ctx: ctx})
}

func (f *ForExpr) PrintGOAssign(
	ctx ast.Ctx,
	name string,
) (string, error) {
	yieldCtx := flowContext{Ctx: ctx, target: name}
	loop, err := f.printGO(yieldCtx)
	if err != nil {
		return "", err
	}

	collectionType := "[]any"
	if ref, known, err := f.yieldCollectionType(ctx); err != nil {
		return "", err
	} else if known {
		collectionType = ref.GoString(ctx)
	}

	return name + " := make(" + collectionType + ", 0)\n" + loop, nil
}

func (f *ForExpr) printGO(ctx ast.Ctx) (string, error) {
	printCtx := f.loopTypeContext(ctx)

	var header string
	switch {
	case f.Iterable != nil:
		iterable, err := f.Iterable.PrintGO(printCtx)
		if err != nil {
			return "", err
		}
		header = "for _, " + string(*f.Ident) + " := range " + iterable
	case f.Condition != nil:
		condition, err := f.Condition.PrintGO(printCtx)
		if err != nil {
			return "", err
		}
		header = "for " + condition
	default:
		header = "for"
	}

	lines := make([]string, 0, len(f.Body))
	for _, expr := range f.Body {
		printed, err := expr.PrintGO(printCtx)
		if err != nil {
			return "", err
		}
		lines = append(
			lines,
			"\t"+strings.ReplaceAll(printed, "\n", "\n\t"),
		)
	}

	return header + " {\n" + strings.Join(lines, "\n") + "\n}", nil
}

func (f *ForExpr) loopTypeContext(ctx ast.Ctx) ast.Ctx {
	if f.Iterable == nil || f.Ident == nil {
		return ctx
	}

	element, known := iterableElementType(ctx, f.Iterable)
	if !known {
		return ctx
	}

	local := ctx.GetLocalCtx()
	local.SetValue(
		string(*f.Ident),
		value.NewTypeWithExplicit(nil, element.String()),
	)

	target := ""
	if yield, ok := ctx.(interface{ YieldTarget() string }); ok {
		target = yield.YieldTarget()
	}
	return flowContext{Ctx: local, target: target}
}

func (f *ForExpr) yieldCollectionType(
	ctx ast.Ctx,
) (ast.TypeRef, bool, error) {
	yieldCtx := f.loopTypeContext(ctx)
	var element ast.TypeRef
	known := true
	found := false

	var visit func([]ast.Expr) error
	visit = func(exprs []ast.Expr) error {
		for _, expr := range exprs {
			if _, nested := expr.(*ForExpr); nested {
				continue
			}
			if yielded, ok := expr.(interface {
				YieldedExpr() ast.Expr
			}); ok {
				ref, ok := inferExprType(yieldCtx, yielded.YieldedExpr())
				if !ok {
					known = false
					found = true
					continue
				}
				if !found {
					element = ref
					found = true
					continue
				}
				if known &&
					ref.GoString(yieldCtx) != element.GoString(yieldCtx) {
					return fmt.Errorf(
						"for yields incompatible types %s and %s",
						element.String(),
						ref.String(),
					)
				}
				continue
			}
			if container, ok := expr.(interface {
				ChildExprs() []ast.Expr
			}); ok {
				if err := visit(container.ChildExprs()); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := visit(f.Body); err != nil {
		return ast.TypeRef{}, false, err
	}
	if !found || !known {
		return ast.TypeRef{}, false, nil
	}

	return ast.TypeRef{
		Kind: ast.SliceTypeRef,
		Elem: &element,
	}, true, nil
}

func iterableElementType(
	ctx ast.Ctx,
	iterable ast.Expr,
) (ast.TypeRef, bool) {
	ref, known := inferExprType(ctx, iterable)
	if !known {
		return ast.TypeRef{}, false
	}

	for ref.Kind == ast.NamedTypeRef {
		def, declared := ctx.GetType(ref.Name)
		if !declared || def.Kind != ast.AliasType {
			break
		}
		ref = def.Underlying
	}

	switch ref.Kind {
	case ast.ArrayTypeRef, ast.SliceTypeRef:
		if ref.Elem != nil {
			return *ref.Elem, true
		}
	default:
		if strings.EqualFold(ref.Name, "string") {
			return ast.TypeRef{Name: "rune"}, true
		}
	}
	return ast.TypeRef{}, false
}

func inferExprType(
	ctx ast.Ctx,
	expr ast.Expr,
) (ast.TypeRef, bool) {
	if typed, ok := expr.(interface {
		ValueType(ast.Ctx) ast.TypeRef
	}); ok {
		ref := typed.ValueType(ctx)
		if !ref.IsZero() {
			return ref, true
		}
	}

	if ident, ok := expr.(Ident); ok {
		if val, found := ctx.GetValue(string(ident)); found {
			return ast.ParseTypeRef(val.TypeName()), true
		}
		return ast.TypeRef{}, false
	}

	switch name := expr.Type(ctx); name {
	case "bool", "comparison":
		return ast.TypeRef{Name: "bool"}, true
	case "int", "float64", "string":
		return ast.TypeRef{Name: name}, true
	}
	return ast.TypeRef{}, false
}

type flowContext struct {
	ast.Ctx
	target string
}

func (y flowContext) YieldTarget() string {
	return y.target
}

func (flowContext) InLoop() bool {
	return true
}
