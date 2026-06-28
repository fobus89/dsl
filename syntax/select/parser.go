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

	var fields [][2]Ident

	for {
		tok := p.CurrentToken()
		{
			if tok.Type != token.IDENT {
				return nil, fmt.Errorf("expected field name")
			}
		}

		var asIndet [2]Ident

		p.Next()

		if p.MatchAllNext(token.AS, token.IDENT) {
			asIndet[1] = literal_parser.NewIdentExpr(p.Peek(-1).Literal)
		} else {
			asIndet[1] = literal_parser.NewIdentExpr(tok.Literal)
		}

		asIndet[0] = literal_parser.NewIdentExpr(tok.Literal)

		fields = append(fields, asIndet)

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

	var (
		where ast.Expr
		limit ast.Expr
	)

	if p.MatchNext(token.WHERE) {
		_where, err := p.ParseStmt()
		{
			if err != nil {
				return nil, err
			}
		}
		where = _where
	}

	if p.MatchNext(token.LIMIT) {
		_limit, err := p.ParseStmt()
		{
			if err != nil {
				return nil, err
			}
		}
		limit = _limit
	}

	return NewSelectExpr(fields, expr, where, limit), nil
}
