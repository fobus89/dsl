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

	switch callee := c.Callee.(type) {
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
			for _, candidate := range meta.MethodReceiverTypes() {
				if fn, ok = ctx.GetMethod(candidate, name); ok {
					receiverType = candidate
					break
				}
			}
		} else {
			fn, ok = ctx.GetMethod(receiverType, name)
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

	return fn(values...)
}

func (*CallExpr) Type(ast.Ctx) string {
	return "call"
}

func (c *CallExpr) Parts() (ast.Expr, []ast.Expr) {
	return c.Callee, c.Args
}

func (c *CallExpr) PrintGO(ctx ast.Ctx) (string, error) {
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
