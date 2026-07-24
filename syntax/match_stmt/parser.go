package matchstmt_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.StmtRegister(token.MATCH, parseMatch)
	p.NudRegister(token.MATCH, parseMatch)
}

func parseMatch(p parser.Parser) (ast.Expr, error) {
	p.Next() // skip match

	subject, err := p.ParseExprUntil(parser.Lowest, token.LBRACE)
	if err != nil {
		return nil, err
	}
	if !p.MatchNext(token.LBRACE) {
		return nil, fmt.Errorf(
			"expected { after match value, got %s",
			p.CurrentToken().Type,
		)
	}

	var arms []MatchArm
	for !p.MatchNext(token.RBRACE) {
		if !p.HasToken() {
			return nil, fmt.Errorf("unterminated match block")
		}

		arm, err := parseMatchArm(p)
		if err != nil {
			return nil, fmt.Errorf(
				"match arm %d: %w",
				len(arms),
				err,
			)
		}
		arms = append(arms, arm)

		if p.MatchNext(token.COMMA) ||
			p.MatchNext(token.SEMICOLON) {
			continue
		}
		if !p.Match(token.RBRACE) {
			return nil, fmt.Errorf(
				"expected , or } after match arm, got %s",
				p.CurrentToken().Type,
			)
		}
	}

	if len(arms) == 0 {
		return nil, fmt.Errorf("match requires at least one arm")
	}
	return NewMatchExpr(subject, arms), nil
}

func parseMatchArm(p parser.Parser) (MatchArm, error) {
	arm := MatchArm{}

	pattern, err := parseMatchPattern(p)
	if err != nil {
		return MatchArm{}, err
	}
	arm.Tree = &pattern

	if p.MatchNext(token.IF) {
		guard, err := p.ParseExprUntil(
			parser.Lowest,
			token.FAT_ARROW,
		)
		if err != nil {
			return MatchArm{}, err
		}
		arm.Guard = guard
	}

	if !p.MatchNext(token.FAT_ARROW) {
		return MatchArm{}, fmt.Errorf(
			"expected => in match arm, got %s",
			p.CurrentToken().Type,
		)
	}

	body, err := parseMatchArmBody(p)
	if err != nil {
		return MatchArm{}, err
	}
	arm.Body = body
	return arm, nil
}

func parseMatchPattern(p parser.Parser) (MatchPattern, error) {
	if p.Match(token.LPARENT) {
		return parseTuplePattern(p)
	}
	if p.Match(token.IDENT) {
		if p.Peek(1).Type == token.DOT {
			return parseEnumPattern(p)
		}
		name := p.Next().Literal
		if name == "_" {
			return MatchPattern{Kind: wildcardPattern}, nil
		}
		return MatchPattern{
			Kind: bindingPattern,
			Name: name,
		}, nil
	}

	expr, err := p.ParseExprUntil(
		parser.Lowest,
		token.IF,
		token.FAT_ARROW,
		token.COMMA,
		token.RPARENT,
		token.RBRACE,
	)
	if err != nil {
		return MatchPattern{}, err
	}
	return MatchPattern{Kind: literalPattern, Expr: expr}, nil
}

func parseTuplePattern(p parser.Parser) (MatchPattern, error) {
	p.Next() // skip (
	pattern := MatchPattern{Kind: tuplePattern}
	for !p.MatchNext(token.RPARENT) {
		child, err := parseMatchPattern(p)
		if err != nil {
			return MatchPattern{}, err
		}
		pattern.Children = append(pattern.Children, child)
		if p.MatchNext(token.RPARENT) {
			break
		}
		if !p.MatchNext(token.COMMA) {
			return MatchPattern{}, fmt.Errorf(
				"expected , in tuple pattern, got %s",
				p.CurrentToken().Type,
			)
		}
	}
	return pattern, nil
}

func parseEnumPattern(p parser.Parser) (MatchPattern, error) {
	enumType := p.Next().Literal
	p.Next() // skip dot
	if !p.Match(token.IDENT) {
		return MatchPattern{}, fmt.Errorf(
			"expected enum variant after %s., got %s",
			enumType,
			p.CurrentToken().Type,
		)
	}
	variantName := p.Next().Literal
	def, declared := p.Ctx().GetType(enumType)
	if !declared || def.Kind != ast.EnumType {
		return MatchPattern{}, fmt.Errorf(
			"type %s is not an enum",
			enumType,
		)
	}
	variant, exists := matchPatternVariant(def, variantName)
	if !exists {
		return MatchPattern{}, fmt.Errorf(
			"variant %s does not exist on enum %s",
			variantName,
			enumType,
		)
	}

	pattern := MatchPattern{
		Kind:        enumPattern,
		EnumType:    enumType,
		EnumVariant: variantName,
		FieldNames:  variant.FieldNames,
	}
	if variant.FieldNames != nil {
		if !p.MatchNext(token.LBRACE) {
			return MatchPattern{}, fmt.Errorf(
				"struct variant %s.%s expects { fields } pattern",
				enumType,
				variantName,
			)
		}
		byName := map[string]MatchPattern{}
		for !p.MatchNext(token.RBRACE) {
			if !p.Match(token.IDENT) {
				return MatchPattern{}, fmt.Errorf(
					"expected field name in %s.%s pattern",
					enumType,
					variantName,
				)
			}
			name := p.Next().Literal
			child := MatchPattern{
				Kind: bindingPattern,
				Name: name,
			}
			if p.MatchNext(token.COLON) {
				var err error
				child, err = parseMatchPattern(p)
				if err != nil {
					return MatchPattern{}, err
				}
			}
			byName[name] = child
			if p.MatchNext(token.RBRACE) {
				break
			}
			if !p.MatchNext(token.COMMA) {
				return MatchPattern{}, fmt.Errorf(
					"expected , in struct variant pattern",
				)
			}
		}
		for _, name := range variant.FieldNames {
			child, ok := byName[name]
			if !ok {
				return MatchPattern{}, fmt.Errorf(
					"missing field %s in %s.%s pattern",
					name,
					enumType,
					variantName,
				)
			}
			pattern.Children = append(pattern.Children, child)
			delete(byName, name)
		}
		for name := range byName {
			return MatchPattern{}, fmt.Errorf(
				"unknown field %s in %s.%s pattern",
				name,
				enumType,
				variantName,
			)
		}
		return pattern, nil
	}

	if len(variant.Fields) != 0 {
		if !p.MatchNext(token.LPARENT) {
			return MatchPattern{}, fmt.Errorf(
				"enum variant %s.%s expects %d payload patterns",
				enumType,
				variantName,
				len(variant.Fields),
			)
		}
		for !p.MatchNext(token.RPARENT) {
			child, err := parseMatchPattern(p)
			if err != nil {
				return MatchPattern{}, err
			}
			pattern.Children = append(pattern.Children, child)
			if p.MatchNext(token.RPARENT) {
				break
			}
			if !p.MatchNext(token.COMMA) {
				return MatchPattern{}, fmt.Errorf(
					"expected , in enum payload pattern",
				)
			}
		}
	}
	return pattern, nil
}

func matchPatternVariant(
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

func parseMatchArmBody(p parser.Parser) ([]ast.Expr, error) {
	if !p.MatchNext(token.LBRACE) {
		stmt, err := p.ParseStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Expr{stmt}, nil
	}

	var body []ast.Expr
	for !p.MatchNext(token.RBRACE) {
		if !p.HasToken() {
			return nil, fmt.Errorf("unterminated match arm block")
		}
		stmt, err := p.ParseStmt()
		if err != nil {
			return nil, err
		}
		body = append(body, stmt)
		p.MatchNext(token.SEMICOLON)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("match arm cannot be empty")
	}
	return body, nil
}

func (m *MatchExpr) MapBranchValues(
	mapper func(ast.Expr) ast.Expr,
) {
	for index := range m.Arms {
		m.Arms[index].Body = mapLastMatchValue(
			m.Arms[index].Body,
			mapper,
		)
	}
}

func mapLastMatchValue(
	body []ast.Expr,
	mapper func(ast.Expr) ast.Expr,
) []ast.Expr {
	if len(body) == 0 {
		return body
	}
	last := len(body) - 1
	if nested, ok := body[last].(interface {
		MapBranchValues(func(ast.Expr) ast.Expr)
	}); ok {
		nested.MapBranchValues(mapper)
		return body
	}
	body[last] = mapper(body[last])
	return body
}
