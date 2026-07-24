package member_parser

import (
	"errors"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.LedRegister(token.DOT, parser.Member, parseMember)
}

// ident.filed.N
func parseMember(p parser.Parser, left ast.Expr, bp parser.BindingPower) (ast.Expr, error) {
	expr := left

	for {
		p.Next() // skip '.'

		tok := p.CurrentToken()
		if tok.Type != token.IDENT &&
			tok.Type != token.INT_LITERAL {
			return nil, errors.New(
				"expected identifier or tuple index after '.'",
			)
		}

		property := literal_parser.NewIdentExpr(tok.Literal)

		expr = NewMemberExpr(expr, property)

		p.Next()

		if p.CurrentToken().Type != token.DOT {
			break
		}

	}

	return expr, nil
}
