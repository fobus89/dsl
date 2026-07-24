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

	if p.Match(token.IDENT) &&
		p.Peek(1).Type == token.DOT {
		enumType := p.Next().Literal
		p.Next() // skip dot
		if !p.Match(token.IDENT) {
			return MatchArm{}, fmt.Errorf(
				"expected enum variant after %s., got %s",
				enumType,
				p.CurrentToken().Type,
			)
		}
		arm.EnumType = enumType
		arm.EnumVariant = p.Next().Literal

		def, declared := p.Ctx().GetType(enumType)
		if !declared || def.Kind != ast.EnumType {
			return MatchArm{}, fmt.Errorf(
				"type %s is not an enum",
				enumType,
			)
		}

		if p.MatchNext(token.LPARENT) {
			for !p.MatchNext(token.RPARENT) {
				if !p.Match(token.IDENT) {
					return MatchArm{}, fmt.Errorf(
						"enum payload pattern expects a binding, got %s",
						p.CurrentToken().Type,
					)
				}
				arm.EnumBindings = append(
					arm.EnumBindings,
					p.Next().Literal,
				)
				if p.MatchNext(token.RPARENT) {
					break
				}
				if !p.MatchNext(token.COMMA) {
					return MatchArm{}, fmt.Errorf(
						"expected , in enum payload pattern, got %s",
						p.CurrentToken().Type,
					)
				}
			}
		}
	} else if p.Match(token.IDENT) {
		name := p.CurrentToken().Literal
		if name == "_" {
			arm.Wildcard = true
		} else {
			arm.Binding = name
		}
		p.Next()
	} else {
		pattern, err := p.ParseExprUntil(
			parser.Lowest,
			token.IF,
			token.FAT_ARROW,
		)
		if err != nil {
			return MatchArm{}, err
		}
		arm.Pattern = pattern
	}

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
