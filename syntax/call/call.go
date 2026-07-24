package call_parser

import (
	"fmt"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type Ident = literal_parser.Ident

type methodCallee interface {
	Receiver() ast.Expr
	MethodName() string
}

type enumCallPrinter interface {
	PrintGOEnumCall(ast.Ctx, []ast.Expr) (string, bool, error)
}

type genericCallee interface {
	BaseExpr() ast.Expr
	TypeArgs() []ast.TypeRef
}

type genericValidator interface {
	ValidateGeneric(ast.Ctx) error
}

type genericMethodPrinter interface {
	PrintGOGenericMethodCall(
		ast.Ctx,
		[]ast.TypeRef,
		[]ast.Expr,
	) (string, bool, error)
}

type CallExpr struct {
	Callee ast.Expr
	Args   []ast.Expr
}

func NewCallExpr(callee ast.Expr, args []ast.Expr) *CallExpr {
	return &CallExpr{Callee: callee, Args: args}
}

func (c *CallExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	var values []value.Type
	var name string
	var fn ast.Func

	calleeExpr := c.Callee
	if generic, ok := calleeExpr.(genericCallee); ok {
		if validator, ok := calleeExpr.(genericValidator); ok {
			if err := validator.ValidateGeneric(ctx); err != nil {
				return value.NewTypeNil(), err
			}
		}
		calleeExpr = generic.BaseExpr()
	}

	switch callee := calleeExpr.(type) {
	case Ident:
		name = string(callee)
		var ok bool
		fn, ok = ctx.GetFunc(name)
		if !ok {
			return value.NewTypeNil(), fmt.Errorf("func %s not found", name)
		}
	case methodCallee:
		name = callee.MethodName()

		receiver, err := callee.Receiver().Eval(ctx)
		if err != nil {
			return value.NewTypeNil(), err
		}

		receiverType := receiver.TypeName()
		var ok bool
		if meta, isMeta := receiver.Any().(ast.MetaTypeValue); isMeta {
			if meta.Def != nil &&
				meta.Def.Kind == ast.EnumType {
				fn, ok = ctx.GetMethod(meta.Def.Name, name)
				receiverType = meta.Def.Name
			}
			for _, candidate := range meta.MethodReceiverTypes() {
				if ok {
					break
				}
				if fn, ok = ctx.GetMethod(candidate, name); ok {
					receiverType = candidate
					break
				}
			}
		} else {
			fn, ok = ctx.GetMethod(receiverType, name)
			if !ok {
				ref := ast.ParseTypeRef(receiverType)
				if len(ref.Args) != 0 {
					fn, ok = ctx.GetMethod(ref.Name, name)
					receiverType = ref.Name
				}
			}
		}
		if !ok {
			return value.NewTypeNil(), fmt.Errorf(
				"method %s not found for type %s",
				name,
				receiverType,
			)
		}
		values = append(values, receiver)
	default:
		return value.NewTypeNil(), fmt.Errorf("expression %T is not callable", c.Callee)
	}

	for _, v := range c.Args {
		val, err := v.Eval(ctx)
		{
			if err != nil {
				return value.NewTypeNil(), err
			}
		}
		values = append(values, val)
	}

	result, err := fn(values...)
	if err != nil {
		return value.NewTypeNil(), err
	}
	if generic, ok := c.Callee.(genericCallee); ok {
		if ident, ok := generic.BaseExpr().(Ident); ok {
			if def, declared := ctx.GetType(
				string(ident),
			); declared {
				ref := ast.TypeRef{
					Name: def.Name,
					Args: generic.TypeArgs(),
				}
				return value.NewTypeWithExplicit(
					result.Any(),
					ref.String(),
				), nil
			}
		}
	}
	return result, nil
}

func (*CallExpr) Type(ast.Ctx) string {
	return "call"
}

func (c *CallExpr) Parts() (ast.Expr, []ast.Expr) {
	return c.Callee, c.Args
}

func (c *CallExpr) PrintGO(ctx ast.Ctx) (string, error) {
	if generic, ok := c.Callee.(genericCallee); ok {
		if validator, ok := c.Callee.(genericValidator); ok {
			if err := validator.ValidateGeneric(ctx); err != nil {
				return "", err
			}
		}
		if printer, ok := generic.BaseExpr().(genericMethodPrinter); ok {
			printed, handled, err :=
				printer.PrintGOGenericMethodCall(
					ctx,
					generic.TypeArgs(),
					c.Args,
				)
			if err != nil {
				return "", err
			}
			if handled {
				return printed, nil
			}
		}
	}
	if printer, ok := c.Callee.(enumCallPrinter); ok {
		printed, handled, err := printer.PrintGOEnumCall(
			ctx,
			c.Args,
		)
		if err != nil {
			return "", err
		}
		if handled {
			return printed, nil
		}
	}
	callee, err := c.Callee.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	args := make([]string, 0, len(c.Args))
	for _, arg := range c.Args {
		printed, err := arg.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		args = append(args, printed)
	}

	return callee + "(" + strings.Join(args, ", ") + ")", nil
}

func (c *CallExpr) ValueType(ctx ast.Ctx) ast.TypeRef {
	if typed, ok := c.Callee.(interface {
		CallValueType(ast.Ctx) ast.TypeRef
	}); ok {
		return typed.CallValueType(ctx)
	}
	return ast.TypeRef{}
}
