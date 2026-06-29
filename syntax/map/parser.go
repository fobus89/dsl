package map_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.NudRegister(token.LBRACE, nudMap)
}

func nudMap(p parser.Parser) (ast.Expr, error) {
	p.Next() // skip {

	var entries []Entry

	for !p.MatchNext(token.RBRACE) {
		keyTok := p.CurrentToken()
		if keyTok.Type != token.IDENT && keyTok.Type != token.STRING_LITERAL {
			return nil, fmt.Errorf("expected map key, got %v", keyTok)
		}

		p.Next()

		if !p.MatchNext(token.COLON) {
			return nil, fmt.Errorf("expected ':' after map key")
		}

		expr, err := p.ParseExpr(parser.Lowest)
		if err != nil {
			return nil, err
		}

		entries = append(entries, Entry{
			key:   keyTok.Literal,
			value: expr,
		})

		if p.MatchNext(token.COMMA) {
			continue
		}

		if !p.MatchNext(token.RBRACE) {
			return nil, fmt.Errorf("expected ',' or '}' after map value")
		}

		break
	}

	return NewMapExpr(entries), nil
}
