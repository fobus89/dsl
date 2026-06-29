package any_parser

import (
	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.LedRegister(token.Any, parser.Relational, ledAny)
}

func ledAny(p parser.Parser, left ast.Expr, bp parser.BindingPower) (ast.Expr, error) {
	p.Next() // skip any

	right, err := p.ParseExpr(bp)
	if err != nil {
		return nil, err
	}

	return NewAnyExpr(left, right), nil
}
