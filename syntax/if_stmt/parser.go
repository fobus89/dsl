package ifstmt_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.StmtRegister(token.IF, parseIf)
}

func parseIf(p parser.Parser) (ast.Expr, error) {
	p.Next() // skip if

	condition, err := p.ParseExprUntil(
		parser.Lowest,
		token.LBRACE,
		token.COLON,
	)
	if err != nil {
		return nil, err
	}

	thenBranch, err := parseBranch(p, "if")
	if err != nil {
		return nil, err
	}

	var elseBranch []ast.Expr
	if p.MatchNext(token.ELSE) {
		if p.Match(token.IF) {
			nested, err := parseIf(p)
			if err != nil {
				return nil, err
			}
			elseBranch = []ast.Expr{nested}
		} else {
			elseBranch, err = parseBranch(p, "else")
			if err != nil {
				return nil, err
			}
		}
	}

	return NewIfExpr(condition, thenBranch, elseBranch), nil
}

func parseBranch(p parser.Parser, owner string) ([]ast.Expr, error) {
	if p.MatchNext(token.COLON) {
		stmt, err := p.ParseStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Expr{stmt}, nil
	}

	if !p.MatchNext(token.LBRACE) {
		return nil, fmt.Errorf(
			"expected { after %s condition, got %s",
			owner,
			p.CurrentToken().Type,
		)
	}

	var body []ast.Expr
	for !p.MatchNext(token.RBRACE) {
		if !p.HasToken() {
			return nil, fmt.Errorf("unterminated %s block", owner)
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
