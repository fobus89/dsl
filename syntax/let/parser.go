package let_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/token"
	"github.com/fobus89/dsl/value"
)

func RegisterParser(p parser.Parser) {
	p.StmtRegister(token.LET, parseValueDecl)
	p.StmtRegister(token.CONST, parseValueDecl)
}

func parseValueDecl(p parser.Parser) (ast.Expr, error) {
	kind := p.Next()
	constant := kind.Type == token.CONST

	var (
		names        []Ident
		tuplePattern bool
	)
	if p.MatchNext(token.LPARENT) {
		if constant {
			return nil, fmt.Errorf(
				"const tuple destructuring is not supported",
			)
		}
		tuplePattern = true
		for !p.MatchNext(token.RPARENT) {
			if !p.Match(token.IDENT) {
				return nil, fmt.Errorf(
					"tuple destructuring expects an identifier, got %s",
					p.CurrentToken().Type,
				)
			}
			names = append(
				names,
				literal_parser.NewIdentExpr(p.Next().Literal),
			)
			if p.MatchNext(token.RPARENT) {
				break
			}
			if !p.MatchNext(token.COMMA) {
				return nil, fmt.Errorf(
					"expected comma in tuple destructuring, got %s",
					p.CurrentToken().Type,
				)
			}
		}
		if len(names) == 0 {
			return nil, fmt.Errorf(
				"tuple destructuring cannot be empty",
			)
		}
	} else if !p.Match(token.IDENT) {
		return nil, fmt.Errorf(
			"expected identifier after %s, got %s",
			kind.Literal,
			p.CurrentToken().Type,
		)
	} else {
		names = []Ident{
			literal_parser.NewIdentExpr(p.Next().Literal),
		}
	}
	name := names[0]

	if !p.MatchNext(token.EQ) {
		return nil, fmt.Errorf(
			"expected = after %s binding, got %s",
			kind.Literal,
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
			"%s %s value: %w",
			kind.Literal,
			name,
			err,
		)
	}

	if !tuplePattern {
		if typed, ok := expr.(interface {
			ValueType(ast.Ctx) ast.TypeRef
		}); ok {
			ref := typed.ValueType(p.Ctx())
			if !ref.IsZero() {
				p.Ctx().SetValueTypeHint(
					string(name),
					value.NewTypeWithExplicit(nil, ref.String()),
				)
			}
		}
		return NewValueDecl(name, expr, constant), nil
	}
	return NewTupleLetExpr(names, expr), nil
}
