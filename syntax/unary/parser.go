package unary_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.NudRegister(token.BANG, nudUnary)
	p.NudRegister(token.MINUS, nudUnary)
	p.NudRegister(token.PLUS, nudUnary)
}

func nudUnary(p parser.Parser) (ast.Expr, error) {
	if !p.MatchAny(token.BANG, token.MINUS, token.PLUS) {
		return nil, fmt.Errorf("expected unary operator, got %v", p.CurrentToken())
	}

	if p.Match(token.BANG) {
		count := 0
		for p.MatchNext(token.BANG) {
			count++
		}

		expr, err := p.ParseExpr(parser.Unary)
		if err != nil {
			return nil, err
		}

		return NewUnaryExpr(token.BANG, expr, count), nil
	}

	opToken := p.Next()

	expr, err := p.ParseExpr(parser.Unary)
	if err != nil {
		return nil, err
	}

	return NewUnaryExpr(opToken.Type, expr, 1), nil
}
