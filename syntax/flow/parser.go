package flow_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.StmtRegister(token.BREAK, parseBreak)
	p.StmtRegister(token.CONTINUE, parseContinue)
	p.StmtRegister(token.YIELD, parseYield)
}

func parseBreak(p parser.Parser) (ast.Expr, error) {
	p.Next()
	return BreakExpr{}, nil
}

func parseContinue(p parser.Parser) (ast.Expr, error) {
	p.Next()
	return ContinueExpr{}, nil
}

func parseYield(p parser.Parser) (ast.Expr, error) {
	p.Next()
	if p.Match(token.RBRACE) || p.Match(token.SEMICOLON) {
		return nil, fmt.Errorf("yield expects a value")
	}

	expr, err := p.ParseExpr(parser.Lowest)
	if err != nil {
		return nil, err
	}
	return &YieldExpr{Value: expr}, nil
}
