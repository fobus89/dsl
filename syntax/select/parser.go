package select_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.NudRegister(token.SELECT, parseSelect)
}

func parseSelect(p parser.Parser) (ast.Expr, error) {

	p.Next() // skip SELECT

	var fields []Ident

	for {
		tok := p.CurrentToken()
		{
			if tok.Type != token.IDENT {
				return nil, fmt.Errorf("expected field name")
			}
		}

		fields = append(fields, literal_parser.NewIdentExpr(tok.Literal))

		p.Next()

		if p.CurrentToken().Type != token.COMMA {
			break
		}

		p.Next() // skip ','
	}

	if p.CurrentToken().Type != token.FROM {
		return nil, fmt.Errorf("expected FROM")
	}

	p.Next() // skip FROM

	if p.CurrentToken().Type != token.IDENT {
		return nil, fmt.Errorf("expected source")
	}

	expr, err := p.ParseExpr(parser.Lowest)
	{
		if err != nil {
			return nil, err
		}
	}

	if !p.MatchNext(token.WHERE) {
		return NewSelectExpr(fields, expr, nil), nil
	}

	where, err := p.ParseStmt()
	{
		if err != nil {
			return nil, err
		}
	}

	return NewSelectExpr(fields, expr, where), nil
}
