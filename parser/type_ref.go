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

	if p.MatchNext(token.LPARENT) {
		ref := ast.TypeRef{
			IsPtr: isPtr,
			Kind:  ast.TupleTypeRef,
		}
		for !p.MatchNext(token.RPARENT) {
			elem, err := ParseTypeRef(p, label)
			if err != nil {
				return ast.TypeRef{}, err
			}
			ref.Elems = append(ref.Elems, elem)
			if p.MatchNext(token.RPARENT) {
				break
			}
			if !p.MatchNext(token.COMMA) {
				return ast.TypeRef{}, fmt.Errorf(
					"expected , in tuple type, got %s at %d:%d",
					p.CurrentToken().Type,
					p.CurrentToken().Line,
					p.CurrentToken().Col,
				)
			}
		}
		return ref, nil
	}

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

	if !p.Match(token.IDENT) &&
		!p.Match(token.TYPE) &&
		!p.CurrentToken().Type.IsType() {
		return ast.TypeRef{}, typeRefError(p, label)
	}

	tok := p.Next()
	ref := ast.TypeRef{Name: tok.Literal, IsPtr: isPtr}
	if p.MatchNext(token.LBRACKET) {
		for !p.MatchNext(token.RBRACKET) {
			arg, err := ParseTypeRef(p, "generic type argument")
			if err != nil {
				return ast.TypeRef{}, err
			}
			ref.Args = append(ref.Args, arg)
			if p.MatchNext(token.RBRACKET) {
				break
			}
			if !p.MatchNext(token.COMMA) {
				return ast.TypeRef{}, fmt.Errorf(
					"expected , in generic type arguments, got %s",
					p.CurrentToken().Type,
				)
			}
		}
	}
	return ref, nil
}

func ParseTypeParams(
	p Parser,
	label string,
) ([]ast.TypeParam, error) {
	if !p.MatchNext(token.LBRACKET) {
		return nil, nil
	}
	var params []ast.TypeParam
	seen := map[string]struct{}{}
	for !p.MatchNext(token.RBRACKET) {
		if !p.Match(token.IDENT) {
			return nil, typeRefError(p, label+" parameter")
		}
		name := p.Next().Literal
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf(
				"generic parameter %s declared more than once",
				name,
			)
		}
		seen[name] = struct{}{}
		constraint, err := ParseTypeRef(p, label+" constraint")
		if err != nil {
			return nil, err
		}
		params = append(params, ast.TypeParam{
			Name:       name,
			Constraint: constraint,
		})
		if p.MatchNext(token.RBRACKET) {
			break
		}
		if !p.MatchNext(token.COMMA) {
			return nil, fmt.Errorf(
				"expected , in generic parameters, got %s",
				p.CurrentToken().Type,
			)
		}
	}
	return params, nil
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
