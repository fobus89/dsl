package all_parser

import (
	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.LedRegister(token.ALL, parser.Relational, ledAll)
}

func ledAll(p parser.Parser, left ast.Expr, bp parser.BindingPower) (ast.Expr, error) {
	p.Next() // skip all

	right, err := p.ParseExpr(bp)
	if err != nil {
		return nil, err
	}

	return NewAllExpr(left, right), nil
}
