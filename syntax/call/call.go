package call_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type Ident = literal_parser.Ident

type CallExpr struct {
	Callee Ident
	Args   []ast.Expr
}

func NewCallExpr(callee Ident, args []ast.Expr) *CallExpr {
	return &CallExpr{Callee: callee, Args: args}
}

func (c *CallExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	fn, ok := ctx.GetFunc(string(c.Callee))
	{
		if !ok {
			return value.NewTypeNil(), fmt.Errorf("func %s not found", c.Callee)
		}
	}

	var values []value.Type

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

func (_ *CallExpr) Type(_ ast.Ctx) string {
	return "Binary"
}
