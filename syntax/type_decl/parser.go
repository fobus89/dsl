package typedecl_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.StmtRegister(token.TYPE, parseTypeDecl)
	p.LedRegister(token.LBRACE, parser.Call, parseStructLiteral)
}

func parseTypeDecl(p parser.Parser) (ast.Expr, error) {
	p.Next() // skip type

	name, err := parseIdent(p, "type name")
	if err != nil {
		return nil, err
	}

	def := ast.TypeDef{
		Name:   string(name),
		Kind:   ast.AliasType,
		Fields: map[string]string{},
	}

	if p.MatchNext(token.Struct) {
		def.Kind = ast.StructType
		def.Underlying = token.Struct.String()

		fields, err := parseStructFields(p)
		if err != nil {
			return nil, err
		}
		def.Fields = fields
	} else {
		underlying, err := parseTypeName(p)
		if err != nil {
			return nil, err
		}
		def.Underlying = underlying
	}

	return NewTypeDecl(p.Ctx(), def), nil
}

func parseStructFields(p parser.Parser) (map[string]string, error) {
	if !p.MatchNext(token.LBRACE) {
		return nil, expected(p, token.LBRACE)
	}

	fields := map[string]string{}
	for !p.MatchNext(token.RBRACE) {
		name, err := parseIdent(p, "field name")
		if err != nil {
			return nil, err
		}
		if _, exists := fields[string(name)]; exists {
			return nil, fmt.Errorf("field %s declared more than once", name)
		}

		if !p.MatchNext(token.COLON) {
			return nil, expected(p, token.COLON)
		}

		fieldType, err := parseTypeName(p)
		if err != nil {
			return nil, err
		}
		fields[string(name)] = fieldType

		if p.MatchNext(token.RBRACE) {
			break
		}
		p.MatchNext(token.COMMA)
		p.MatchNext(token.SEMICOLON)
	}

	return fields, nil
}

func parseStructLiteral(
	p parser.Parser,
	left ast.Expr,
	_ parser.BindingPower,
) (ast.Expr, error) {
	typeName, ok := left.(Ident)
	if !ok {
		return nil, fmt.Errorf("struct literal requires a type name, got %T", left)
	}

	p.Next() // skip {

	var fields []FieldValue
	seen := map[string]struct{}{}
	for !p.MatchNext(token.RBRACE) {
		name, err := parseIdent(p, "field name")
		if err != nil {
			return nil, err
		}
		if _, exists := seen[string(name)]; exists {
			return nil, fmt.Errorf("field %s initialized more than once", name)
		}
		seen[string(name)] = struct{}{}

		if !p.MatchNext(token.COLON) {
			return nil, expected(p, token.COLON)
		}

		expr, err := p.ParseExpr(parser.Lowest)
		if err != nil {
			return nil, err
		}
		fields = append(fields, FieldValue{Name: string(name), Value: expr})

		if p.MatchNext(token.RBRACE) {
			break
		}
		if !p.MatchNext(token.COMMA) {
			return nil, expected(p, token.COMMA)
		}
	}

	return NewStructLiteral(typeName, fields), nil
}

func parseTypeName(p parser.Parser) (string, error) {
	tok := p.CurrentToken()
	if tok.Type != token.IDENT && !tok.Type.IsType() {
		return "", fmt.Errorf(
			"expected type name, got %s at %d:%d",
			tok.Type,
			tok.Line,
			tok.Col,
		)
	}

	p.Next()
	return tok.Literal, nil
}

func parseIdent(p parser.Parser, label string) (Ident, error) {
	tok := p.CurrentToken()
	if tok.Type != token.IDENT {
		return "", fmt.Errorf(
			"expected %s, got %s at %d:%d",
			label,
			tok.Type,
			tok.Line,
			tok.Col,
		)
	}

	p.Next()
	return literal_parser.NewIdentExpr(tok.Literal), nil
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
