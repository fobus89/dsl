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

type forResultMode uint8

const (
	noForResult forResultMode = iota
	yieldForResult
	breakForResult
)

type forResultInfo struct {
	mode  forResultMode
	ref   ast.TypeRef
	known bool
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
	info, err := f.resultInfo(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	if f.Iterable != nil {
		iterable, err := f.Iterable.Eval(ctx)
		if err != nil {
			return value.NewTypeNil(), err
		}
		var scalar *value.Type
		collected, scalar, err = f.evalIterable(ctx, iterable.Any())
		if err != nil {
			return value.NewTypeNil(), err
		}
		if scalar != nil {
			return scalarResult(info, *scalar), nil
		}
		return emptyForResult(ctx, info, collected), nil
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
			if signal.HasValue {
				return scalarResult(info, signal.Value), nil
			}
			return emptyForResult(ctx, info, collected), nil
		case ast.ContinueFlow:
			continue
		case ast.YieldFlow:
			collected = append(collected, signal.Value.Any())
		}
	}

	return emptyForResult(ctx, info, collected), nil
}

func emptyForResult(
	ctx ast.Ctx,
	info forResultInfo,
	collected []any,
) value.Type {
	if info.mode == breakForResult {
		if info.known {
			return value.NewTypeWithExplicit(
				zeroForType(ctx, info.ref),
				info.ref.String(),
			)
		}
		return value.NewTypeNil()
	}
	if info.mode == yieldForResult && info.known {
		return value.NewTypeWithExplicit(collected, info.ref.String())
	}
	return value.NewType(collected)
}

func zeroForType(ctx ast.Ctx, ref ast.TypeRef) any {
	if ref.IsPtr || ref.Kind == ast.SliceTypeRef {
		return nil
	}
	if ref.Kind == ast.ArrayTypeRef {
		result := make([]any, ref.Len)
		if ref.Elem != nil {
			for i := range result {
				result[i] = zeroForType(ctx, *ref.Elem)
			}
		}
		return result
	}
	if def, declared := ctx.GetType(ref.Name); declared {
		if def.Kind == ast.AliasType {
			return zeroForType(ctx, def.Underlying)
		}
		fields := make(map[string]any, len(def.Fields))
		for name, field := range def.Fields {
			fields[name] = zeroForType(ctx, field.Type)
		}
		return fields
	}
	switch strings.ToLower(ref.Name) {
	case "string":
		return ""
	case "bool":
		return false
	case "float32", "float64":
		return float64(0)
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"rune", "byte":
		return int64(0)
	}
	return nil
}

func scalarResult(
	info forResultInfo,
	result value.Type,
) value.Type {
	if info.known {
		return value.NewTypeWithExplicit(
			result.Any(),
			info.ref.String(),
		)
	}
	return result
}

func (f *ForExpr) evalIterable(
	ctx ast.Ctx,
	iterable any,
) ([]any, *value.Type, error) {
	var collected []any

	if text, ok := iterable.(string); ok {
		for _, item := range text {
			ctx.SetValue(string(*f.Ident), value.NewType(item))
			stop, err := collectFlow(ctx, f.Body, &collected)
			if err != nil {
				return nil, nil, err
			}
			if stop.scalar != nil {
				return collected, stop.scalar, nil
			}
			if stop.stop {
				break
			}
		}
		return collected, nil, nil
	}

	reflected := reflect.ValueOf(iterable)
	if !reflected.IsValid() {
		return collected, nil, nil
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
				return nil, nil, err
			}
			if stop.scalar != nil {
				return collected, stop.scalar, nil
			}
			if stop.stop {
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
				return nil, nil, err
			}
			if stop.scalar != nil {
				return collected, stop.scalar, nil
			}
			if stop.stop {
				break
			}
		}
	default:
		return nil, nil, fmt.Errorf("cannot iterate over %T", iterable)
	}

	return collected, nil, nil
}

type flowStop struct {
	stop   bool
	scalar *value.Type
}

func collectFlow(
	ctx ast.Ctx,
	body []ast.Expr,
	collected *[]any,
) (flowStop, error) {
	err := evalBody(ctx, body)
	if err == nil {
		return flowStop{}, nil
	}

	signal, ok := errors.AsType[ast.FlowSignal](err)
	if !ok {
		return flowStop{}, err
	}
	switch signal.Kind {
	case ast.BreakFlow:
		if signal.HasValue {
			scalar := signal.Value
			return flowStop{stop: true, scalar: &scalar}, nil
		}
		return flowStop{stop: true}, nil
	case ast.ContinueFlow:
		return flowStop{}, nil
	case ast.YieldFlow:
		*collected = append(*collected, signal.Value.Any())
		return flowStop{}, nil
	default:
		return flowStop{}, err
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
	info, err := f.resultInfo(ctx)
	if err != nil || !info.known {
		return ast.TypeRef{}
	}
	return info.ref
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
	info, err := f.resultInfo(ctx)
	if err != nil {
		return "", err
	}

	printCtx := flowContext{Ctx: ctx}
	resultType := "[]any"
	switch info.mode {
	case yieldForResult:
		printCtx.yieldTarget = name
		if info.known {
			resultType = info.ref.GoString(ctx)
		}
	case breakForResult:
		printCtx.breakReturns = true
		scalarType := "any"
		if info.known {
			scalarType = info.ref.GoString(ctx)
		}
		loop, err := f.printGO(printCtx)
		if err != nil {
			return "", err
		}
		indentedLoop := "\t" +
			strings.ReplaceAll(loop, "\n", "\n\t")
		return name + " := func() " + scalarType + " {\n" +
			indentedLoop + "\n" +
			"\tvar zero " + scalarType + "\n" +
			"\treturn zero\n" +
			"}()", nil
	}

	loop, err := f.printGO(printCtx)
	if err != nil {
		return "", err
	}
	indentedLoop := "\t" +
		strings.ReplaceAll(loop, "\n", "\n\t")
	return name + " := func() " + resultType + " {\n" +
		"\t" + name + " := make(" + resultType + ", 0)\n" +
		indentedLoop + "\n" +
		"\treturn " + name + "\n" +
		"}()", nil
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

	yieldTarget := ""
	if yield, ok := ctx.(interface{ YieldTarget() string }); ok {
		yieldTarget = yield.YieldTarget()
	}
	breakReturns := false
	if breaker, ok := ctx.(interface{ BreakReturnsValue() bool }); ok {
		breakReturns = breaker.BreakReturnsValue()
	}
	return flowContext{
		Ctx:          local,
		yieldTarget:  yieldTarget,
		breakReturns: breakReturns,
	}
}

func (f *ForExpr) resultInfo(
	ctx ast.Ctx,
) (forResultInfo, error) {
	yieldCtx := f.loopTypeContext(ctx)
	var element ast.TypeRef
	known := true
	found := false
	mode := noForResult

	var visit func([]ast.Expr) error
	visit = func(exprs []ast.Expr) error {
		for _, expr := range exprs {
			if _, nested := expr.(*ForExpr); nested {
				continue
			}
			if yielded, ok := expr.(interface {
				YieldedExpr() ast.Expr
			}); ok {
				if mode == breakForResult {
					return fmt.Errorf(
						"for cannot mix yield with break value",
					)
				}
				mode = yieldForResult
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
			if breaker, ok := expr.(interface {
				BreakValue() (ast.Expr, bool)
			}); ok {
				valueExpr, hasValue := breaker.BreakValue()
				if !hasValue {
					continue
				}
				if mode == yieldForResult {
					return fmt.Errorf(
						"for cannot mix yield with break value",
					)
				}
				mode = breakForResult
				ref, ok := inferExprType(yieldCtx, valueExpr)
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
						"for breaks with incompatible types %s and %s",
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
		return forResultInfo{}, err
	}
	if !found || !known {
		return forResultInfo{mode: mode}, nil
	}

	if mode == yieldForResult {
		return forResultInfo{
			mode:  mode,
			known: true,
			ref: ast.TypeRef{
				Kind: ast.SliceTypeRef,
				Elem: &element,
			},
		}, nil
	}
	return forResultInfo{
		mode:  mode,
		ref:   element,
		known: true,
	}, nil
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
	yieldTarget  string
	breakReturns bool
}

func (y flowContext) YieldTarget() string {
	return y.yieldTarget
}

func (y flowContext) BreakReturnsValue() bool {
	return y.breakReturns
}

func (flowContext) InLoop() bool {
	return true
}
