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
	if p.Match(token.RBRACE) ||
		p.Match(token.SEMICOLON) ||
		p.Match(token.ELSE) ||
		!p.HasToken() {
		return &BreakExpr{}, nil
	}

	if p.Match(token.IF) {
		handler, ok := p.StmtOrNone(token.IF)
		if !ok {
			return nil, fmt.Errorf("if parser is not registered")
		}
		expr, err := handler(p)
		if err != nil {
			return nil, err
		}
		conditional, ok := expr.(interface {
			MapBranchValues(func(ast.Expr) ast.Expr)
		})
		if !ok {
			return nil, fmt.Errorf(
				"break if expects a conditional expression",
			)
		}
		conditional.MapBranchValues(func(value ast.Expr) ast.Expr {
			return &BreakExpr{Value: value}
		})
		return expr, nil
	}

	expr, err := p.ParseExpr(parser.Lowest)
	if err != nil {
		return nil, err
	}
	return &BreakExpr{Value: expr}, nil
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

	if p.Match(token.IF) {
		handler, ok := p.StmtOrNone(token.IF)
		if !ok {
			return nil, fmt.Errorf("if parser is not registered")
		}
		expr, err := handler(p)
		if err != nil {
			return nil, err
		}
		conditional, ok := expr.(interface {
			MapBranchValues(func(ast.Expr) ast.Expr)
		})
		if !ok {
			return nil, fmt.Errorf(
				"yield if expects a conditional expression",
			)
		}
		conditional.MapBranchValues(func(value ast.Expr) ast.Expr {
			return &YieldExpr{Value: value}
		})
		return expr, nil
	}

	expr, err := p.ParseExpr(parser.Lowest)
	if err != nil {
		return nil, err
	}
	return &YieldExpr{Value: expr}, nil
}
