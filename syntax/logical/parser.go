package logical_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.LedRegister(token.AMP_AMP, parser.Logical, ledLogical)
	p.LedRegister(token.PIPE_PIPE, parser.Logical, ledLogical)
	p.LedRegister(token.AND, parser.Logical, ledLogical)
	p.LedRegister(token.OR, parser.Logical, ledLogical)
}

func ledLogical(p parser.Parser, left ast.Expr, bp parser.BindingPower) (ast.Expr, error) {
	if !p.MatchAny(token.AMP_AMP, token.PIPE_PIPE, token.AND, token.OR) {
		return nil, fmt.Errorf("expected logical operator, got %v", p.CurrentToken())
	}

	opToken := p.Next()

	right, err := p.ParseExpr(bp)
	if err != nil {
		return nil, err
	}

	return NewLogicalExpr(opToken.Type, left, right), nil
}
