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
		return value.NewType(collected), nil
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
			return value.NewType(collected), nil
		case ast.ContinueFlow:
			continue
		case ast.YieldFlow:
			collected = append(collected, signal.Value.Any())
		}
	}

	return value.NewType(collected), nil
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

	return name + " := make([]any, 0)\n" + loop, nil
}

func (f *ForExpr) printGO(ctx ast.Ctx) (string, error) {
	var header string
	switch {
	case f.Iterable != nil:
		iterable, err := f.Iterable.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		header = "for _, " + string(*f.Ident) + " := range " + iterable
	case f.Condition != nil:
		condition, err := f.Condition.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		header = "for " + condition
	default:
		header = "for"
	}

	lines := make([]string, 0, len(f.Body))
	for _, expr := range f.Body {
		printed, err := expr.PrintGO(ctx)
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
