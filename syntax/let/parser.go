package let_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.StmtRegister(token.LET, parseLet)
}

func parseLet(p parser.Parser) (ast.Expr, error) {
	p.Next() // skip let

	if !p.Match(token.IDENT) {
		return nil, fmt.Errorf(
			"expected identifier after let, got %s",
			p.CurrentToken().Type,
		)
	}
	name := literal_parser.NewIdentExpr(p.Next().Literal)

	if !p.MatchNext(token.EQ) {
		return nil, fmt.Errorf(
			"expected = after let %s, got %s",
			name,
			p.CurrentToken().Type,
		)
	}

	var (
		expr ast.Expr
		err  error
	)
	if handler, ok := p.StmtOrNone(p.CurrentToken().Type); ok &&
		(p.Match(token.IF) ||
			p.Match(token.FOR) ||
			p.Match(token.MATCH)) {
		expr, err = handler(p)
	} else {
		expr, err = p.ParseExpr(parser.Lowest)
	}
	if err != nil {
		return nil, fmt.Errorf(
			"let %s value: %w",
			name,
			err,
		)
	}

	return NewLetExpr(name, expr), nil
}
