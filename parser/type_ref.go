package parser

import (
	"fmt"
	"strconv"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/token"
)

// ParseTypeRef parses a recursive type such as *[2][]*User.
func ParseTypeRef(p Parser, label string) (ast.TypeRef, error) {
	isPtr := p.MatchNext(token.STAR)

	if p.MatchNext(token.LBRACKET) {
		ref := ast.TypeRef{IsPtr: isPtr}
		if p.MatchNext(token.RBRACKET) {
			ref.Kind = ast.SliceTypeRef
		} else {
			lengthToken := p.CurrentToken()
			if !p.MatchNext(token.INT_LITERAL) {
				return ast.TypeRef{}, typeRefError(p, label)
			}

			length, err := strconv.Atoi(lengthToken.Literal)
			if err != nil || length < 0 {
				return ast.TypeRef{}, fmt.Errorf(
					"invalid array length %q at %d:%d",
					lengthToken.Literal,
					lengthToken.Line,
					lengthToken.Col,
				)
			}
			if !p.MatchNext(token.RBRACKET) {
				return ast.TypeRef{}, fmt.Errorf(
					"expected ], got %s at %d:%d",
					p.CurrentToken().Type,
					p.CurrentToken().Line,
					p.CurrentToken().Col,
				)
			}
			ref.Kind = ast.ArrayTypeRef
			ref.Len = length
		}

		elem, err := ParseTypeRef(p, label)
		if err != nil {
			return ast.TypeRef{}, err
		}
		ref.Elem = &elem
		return ref, nil
	}

	if !p.Match(token.IDENT) && !p.CurrentToken().Type.IsType() {
		return ast.TypeRef{}, typeRefError(p, label)
	}

	tok := p.Next()
	return ast.TypeRef{Name: tok.Literal, IsPtr: isPtr}, nil
}

func typeRefError(p Parser, label string) error {
	tok := p.CurrentToken()
	return fmt.Errorf(
		"expected %s, got %s at %d:%d",
		label,
		tok.Type,
		tok.Line,
		tok.Col,
	)
}
