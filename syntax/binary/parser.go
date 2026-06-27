package binary_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {

	p.NudRegister(token.LPARENT, nudGrouping)

	p.LedRegister(token.PLUS, parser.Additive, ledBinary)
	p.LedRegister(token.MINUS, parser.Additive, ledBinary)

	p.LedRegister(token.STAR, parser.Muptiplicative, ledBinary)
	p.LedRegister(token.SLASH, parser.Muptiplicative, ledBinary)
	p.LedRegister(token.PERCENT, parser.Muptiplicative, ledBinary)
}

func nudGrouping(p parser.Parser) (ast.Expr, error) {
	if !p.MatchNext(token.LPARENT) {
		return nil, fmt.Errorf("expected LPARENT, got %v", p.CurrentToken())
	}

	expr, err := p.ParseExpr(parser.Lowest)
	{
		if err != nil {
			return nil, err
		}
	}

	if !p.MatchNext(token.RPARENT) {
		return nil, fmt.Errorf("expected RPARENT, got %v", p.CurrentToken())
	}

	return expr, nil
}

func ledBinary(p parser.Parser, left ast.Expr, bp parser.BindingPower) (ast.Expr, error) {
	if !p.MatchAny(token.PLUS, token.MINUS, token.STAR, token.SLASH, token.PERCENT) {
		return nil, fmt.Errorf("expected PLUS, MINUS, STAR, SLASH, PERCEN got %v", p.CurrentToken())
	}

	opToken := p.Next()

	right, err := p.ParseExpr(bp)
	{
		if err != nil {
			return nil, err
		}
	}

	return NewBinaryExpr(opToken.Type, left, right), nil
}
