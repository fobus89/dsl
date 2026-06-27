package call_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.LedRegister(token.LPARENT, parser.Call, ledCall)
}

func ledCall(p parser.Parser, left ast.Expr, bp parser.BindingPower) (ast.Expr, error) {
	p.Next() // skip LPARENT

	var exprs []ast.Expr

	for !p.MatchNext(token.RPARENT) {
		expr, err := p.ParseExpr(parser.Lowest)
		{
			if err != nil {
				return nil, err
			}
		}

		p.MatchNext(token.COMMA) // skip ,

		exprs = append(exprs, expr)
	}

	ident, ok := left.(Ident)
	{
		if !ok {
			return nil, fmt.Errorf("ident %s not found", ident)
		}
	}

	return NewCallExpr(ident, exprs), nil
}
