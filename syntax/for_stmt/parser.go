package forstmt_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.StmtRegister(token.FOR, parseFor)
}

func parseFor(p parser.Parser) (ast.Expr, error) {
	p.Next() // skip for

	var (
		ident     *Ident
		iterable  ast.Expr
		condition ast.Expr
		err       error
	)

	if p.Match(token.IDENT) && p.Peek(1).Type == token.IN {
		name := literal_parser.NewIdentExpr(p.Next().Literal)
		ident = &name
		p.Next() // skip in

		iterable, err = p.ParseExprUntil(
			parser.Lowest,
			token.LBRACE,
			token.COLON,
		)
		if err != nil {
			return nil, err
		}
	} else if !p.Match(token.LBRACE) {
		condition, err = p.ParseExprUntil(
			parser.Lowest,
			token.LBRACE,
			token.COLON,
		)
		if err != nil {
			return nil, err
		}
	}

	body, err := parseForBlock(p)
	if err != nil {
		return nil, err
	}

	return NewForExpr(ident, iterable, condition, body), nil
}

func parseForBlock(p parser.Parser) ([]ast.Expr, error) {
	if p.MatchNext(token.COLON) {
		stmt, err := p.ParseStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Expr{stmt}, nil
	}

	if !p.MatchNext(token.LBRACE) {
		return nil, fmt.Errorf(
			"expected { after for clause, got %s",
			p.CurrentToken().Type,
		)
	}

	var body []ast.Expr
	for !p.MatchNext(token.RBRACE) {
		if !p.HasToken() {
			return nil, fmt.Errorf("unterminated for block")
		}

		stmt, err := p.ParseStmt()
		if err != nil {
			return nil, err
		}
		body = append(body, stmt)
		p.MatchNext(token.SEMICOLON)
	}

	return body, nil
}
