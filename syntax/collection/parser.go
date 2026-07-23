package collection_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.NudRegister(token.LBRACKET, nudCollection)
}

func nudCollection(p parser.Parser) (ast.Expr, error) {
	typeRef, err := parser.ParseTypeRef(p, "array or slice type")
	if err != nil {
		return nil, err
	}
	if typeRef.Kind != ast.ArrayTypeRef &&
		typeRef.Kind != ast.SliceTypeRef {
		return nil, fmt.Errorf("collection literal requires an array or slice type")
	}
	if !p.MatchNext(token.LBRACE) {
		return nil, expected(p, token.LBRACE)
	}

	return parseCollectionBody(p, typeRef)
}

func parseCollectionBody(
	p parser.Parser,
	typeRef ast.TypeRef,
) (ast.Expr, error) {
	var elements []ast.Expr

	for !p.MatchNext(token.RBRACE) {
		if !p.HasToken() {
			return nil, fmt.Errorf(
				"unterminated collection literal %s",
				typeRef.String(),
			)
		}

		var element ast.Expr
		var err error
		if typeRef.Elem != nil &&
			(typeRef.Elem.Kind == ast.ArrayTypeRef ||
				typeRef.Elem.Kind == ast.SliceTypeRef) &&
			p.MatchNext(token.LBRACE) {
			element, err = parseCollectionBody(p, *typeRef.Elem)
		} else {
			element, err = p.ParseExpr(parser.Lowest)
		}
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)

		if p.MatchNext(token.RBRACE) {
			break
		}
		if !p.MatchNext(token.COMMA) {
			return nil, expected(p, token.COMMA)
		}
	}

	if typeRef.Kind == ast.ArrayTypeRef &&
		len(elements) > typeRef.Len {
		return nil, fmt.Errorf(
			"array %s has length %d but %d values were provided",
			typeRef.String(),
			typeRef.Len,
			len(elements),
		)
	}

	return NewCollectionExpr(typeRef, elements), nil
}

func expected(p parser.Parser, kind token.TokenType) error {
	tok := p.CurrentToken()
	return fmt.Errorf(
		"expected %s, got %s at %d:%d",
		kind,
		tok.Type,
		tok.Line,
		tok.Col,
	)
}
