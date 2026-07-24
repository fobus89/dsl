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
	p.StmtRegister(token.ENUM, parseEnumDecl)
	p.LedRegister(token.LBRACE, parser.Call, parseStructLiteral)
}

func parseEnumDecl(p parser.Parser) (ast.Expr, error) {
	p.Next() // skip enum

	name, err := parseIdent(p, "enum name")
	if err != nil {
		return nil, err
	}
	if !p.MatchNext(token.LBRACE) {
		return nil, expected(p, token.LBRACE)
	}

	def := ast.TypeDef{
		Name:       string(name),
		Kind:       ast.EnumType,
		Underlying: ast.TypeRef{Name: token.Enum.String()},
		Fields:     map[string]ast.FieldDef{},
	}
	seen := map[string]struct{}{}

	for !p.MatchNext(token.RBRACE) {
		variant, err := parseIdent(p, "enum variant")
		if err != nil {
			return nil, err
		}
		variantName := string(variant)
		if _, exists := seen[variantName]; exists {
			return nil, fmt.Errorf(
				"enum variant %s declared more than once",
				variantName,
			)
		}
		seen[variantName] = struct{}{}

		var fields []ast.TypeRef
		var fieldNames []string
		if p.MatchNext(token.LPARENT) {
			for !p.MatchNext(token.RPARENT) {
				fieldType, err := parseTypeName(p)
				if err != nil {
					return nil, err
				}
				fields = append(fields, fieldType)
				if p.MatchNext(token.RPARENT) {
					break
				}
				if !p.MatchNext(token.COMMA) {
					return nil, expected(p, token.COMMA)
				}
			}
		} else if p.MatchNext(token.LBRACE) {
			seenFields := map[string]struct{}{}
			for !p.MatchNext(token.RBRACE) {
				fieldName, err := parseIdent(
					p,
					"enum variant field",
				)
				if err != nil {
					return nil, err
				}
				if _, exists := seenFields[string(fieldName)]; exists {
					return nil, fmt.Errorf(
						"enum variant field %s declared more than once",
						fieldName,
					)
				}
				seenFields[string(fieldName)] = struct{}{}
				if !p.MatchNext(token.COLON) {
					return nil, expected(p, token.COLON)
				}
				fieldType, err := parseTypeName(p)
				if err != nil {
					return nil, err
				}
				fieldNames = append(fieldNames, string(fieldName))
				fields = append(fields, fieldType)
				if p.MatchNext(token.RBRACE) {
					break
				}
				if !p.MatchNext(token.COMMA) {
					return nil, expected(p, token.COMMA)
				}
			}
		}
		def.Variants = append(def.Variants, ast.EnumVariantDef{
			Name:       variantName,
			Fields:     fields,
			FieldNames: fieldNames,
		})

		if p.MatchNext(token.RBRACE) {
			break
		}
		if !p.MatchNext(token.COMMA) &&
			!p.MatchNext(token.SEMICOLON) {
			return nil, fmt.Errorf(
				"expected , or } after enum variant, got %s",
				p.CurrentToken().Type,
			)
		}
	}

	if len(def.Variants) == 0 {
		return nil, fmt.Errorf("enum %s requires at least one variant", name)
	}
	return NewTypeDecl(p.Ctx(), def), nil
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
		Fields: map[string]ast.FieldDef{},
	}

	if p.MatchNext(token.Struct) {
		def.Kind = ast.StructType
		def.Underlying = ast.TypeRef{Name: token.Struct.String()}

		fields, err := parseStructFields(p)
		if err != nil {
			return nil, err
		}
		def.Fields = fields
	} else {
		p.MatchNext(token.EQ)
		underlying, err := parseTypeName(p)
		if err != nil {
			return nil, err
		}
		def.Underlying = underlying
	}

	return NewTypeDecl(p.Ctx(), def), nil
}

func parseStructFields(p parser.Parser) (map[string]ast.FieldDef, error) {
	if !p.MatchNext(token.LBRACE) {
		return nil, expected(p, token.LBRACE)
	}

	fields := map[string]ast.FieldDef{}
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

		var defaultValue ast.Expr
		if p.MatchNext(token.EQ) {
			defaultValue, err = p.ParseExpr(parser.Lowest)
			if err != nil {
				return nil, err
			}
		}
		fields[string(name)] = ast.FieldDef{
			Type:    fieldType,
			Default: defaultValue,
		}

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
	if path, ok := left.(interface {
		Path() ([]string, bool)
	}); ok {
		parts, valid := path.Path()
		if valid && len(parts) == 2 {
			return parseEnumStructLiteral(p, parts[0], parts[1])
		}
	}

	typeName, ok := left.(Ident)
	if !ok {
		return nil, fmt.Errorf(
			"struct literal requires a type name, got %T",
			left,
		)
	}

	p.Next() // skip {

	fields, err := parseFieldValues(p)
	if err != nil {
		return nil, err
	}
	return NewStructLiteral(typeName, fields), nil
}

func parseEnumStructLiteral(
	p parser.Parser,
	enumName string,
	variantName string,
) (ast.Expr, error) {
	def, declared := p.Ctx().GetType(enumName)
	if !declared || def.Kind != ast.EnumType {
		return nil, fmt.Errorf("type %s is not an enum", enumName)
	}
	variant, exists := enumVariantByName(def, variantName)
	if !exists || variant.FieldNames == nil {
		return nil, fmt.Errorf(
			"%s.%s is not a struct enum variant",
			enumName,
			variantName,
		)
	}
	p.Next() // skip {
	fields, err := parseFieldValues(p)
	if err != nil {
		return nil, err
	}
	return NewEnumStructLiteral(
		enumName,
		variantName,
		fields,
	), nil
}

func parseFieldValues(
	p parser.Parser,
) ([]FieldValue, error) {
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

	return fields, nil
}

func enumVariantByName(
	def ast.TypeDef,
	name string,
) (ast.EnumVariantDef, bool) {
	for _, variant := range def.Variants {
		if variant.Name == name {
			return variant, true
		}
	}
	return ast.EnumVariantDef{}, false
}

func parseTypeName(p parser.Parser) (ast.TypeRef, error) {
	return parser.ParseTypeRef(p, "type name")
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
