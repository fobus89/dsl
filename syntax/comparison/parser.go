package comparison_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.LedRegister(token.GT, parser.Relational, ledComparison)
	p.LedRegister(token.LT, parser.Relational, ledComparison)
	p.LedRegister(token.GT_EQ, parser.Relational, ledComparison)
	p.LedRegister(token.LT_EQ, parser.Relational, ledComparison)
	p.LedRegister(token.EQ_EQ, parser.Relational, ledComparison)
	p.LedRegister(token.BANG_EQ, parser.Relational, ledComparison)
}

func ledComparison(p parser.Parser, left ast.Expr, bp parser.BindingPower) (ast.Expr, error) {
	if !p.MatchAny(token.GT, token.LT, token.GT_EQ, token.LT_EQ, token.EQ_EQ, token.BANG_EQ) {
		return nil, fmt.Errorf("expected comparison operator, got %v", p.CurrentToken())
	}

	opToken := p.Next()

	right, err := p.ParseExpr(bp)
	if err != nil {
		return nil, err
	}

	return NewComparisonExpr(opToken.Type, left, right), nil
}
