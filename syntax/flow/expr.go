package flow_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/value"
)

type BreakExpr struct{}
type ContinueExpr struct{}

type YieldExpr struct {
	Value ast.Expr
}

func (BreakExpr) Eval(ast.Ctx) (value.Type, error) {
	return value.NewTypeNil(), ast.FlowSignal{Kind: ast.BreakFlow}
}

func (BreakExpr) Type(ast.Ctx) string {
	return "break"
}

func (BreakExpr) PrintGO(ctx ast.Ctx) (string, error) {
	if loop, ok := ctx.(interface{ InLoop() bool }); !ok || !loop.InLoop() {
		return "", fmt.Errorf("break can only be used inside for")
	}
	return "break", nil
}

func (ContinueExpr) Eval(ast.Ctx) (value.Type, error) {
	return value.NewTypeNil(), ast.FlowSignal{Kind: ast.ContinueFlow}
}

func (ContinueExpr) Type(ast.Ctx) string {
	return "continue"
}

func (ContinueExpr) PrintGO(ctx ast.Ctx) (string, error) {
	if loop, ok := ctx.(interface{ InLoop() bool }); !ok || !loop.InLoop() {
		return "", fmt.Errorf("continue can only be used inside for")
	}
	return "continue", nil
}

func (y *YieldExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	result, err := y.Value.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}
	return value.NewTypeNil(), ast.FlowSignal{
		Kind:  ast.YieldFlow,
		Value: result,
	}
}

func (*YieldExpr) Type(ast.Ctx) string {
	return "yield"
}

func (y *YieldExpr) PrintGO(ctx ast.Ctx) (string, error) {
	target, ok := ctx.(interface{ YieldTarget() string })
	if !ok || target.YieldTarget() == "" {
		return "", fmt.Errorf("yield can only be generated inside a collected for expression")
	}

	printed, err := y.Value.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	name := target.YieldTarget()
	return name + " = append(" + name + ", " + printed + ")", nil
}
