package funcdecl_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.StmtRegister(token.FN, parseFuncDecl)
	p.StmtRegister(token.RETURN, parseReturnStmt)
}

func parseFuncDecl(p parser.Parser) (ast.Expr, error) {
	p.Next() // skip fn

	recv, err := parseReceiver(p)
	if err != nil {
		return nil, err
	}

	name, err := parseIdent(p, "function name")
	if err != nil {
		return nil, err
	}

	params, err := parseParams(p)
	if err != nil {
		return nil, err
	}

	var returnType *ast.TypeRef
	if p.Match(token.STAR) ||
		p.Match(token.IDENT) ||
		p.CurrentToken().Type.IsType() {
		typeRef, err := parseTypeRef(p, "return type")
		if err != nil {
			return nil, err
		}
		returnType = typeRef
	}

	if !p.MatchNext(token.LBRACE) {
		return nil, expected(p, token.LBRACE)
	}

	var body []ast.Expr
	for !p.MatchNext(token.RBRACE) {
		if !p.HasToken() {
			return nil, fmt.Errorf("unterminated body of func %s", name)
		}

		stmt, err := p.ParseStmt()
		if err != nil {
			return nil, err
		}
		body = append(body, stmt)
		p.MatchNext(token.SEMICOLON)
	}

	return NewFuncDecl(p.Ctx(), recv, name, params, returnType, body), nil
}

func parseReceiver(p parser.Parser) (*Param, error) {
	if !p.MatchNext(token.LPARENT) {
		return nil, nil
	}

	name, err := parseIdent(p, "receiver name")
	if err != nil {
		return nil, err
	}

	if !p.MatchNext(token.COLON) {
		return nil, expected(p, token.COLON)
	}

	receiverType, err := parseTypeRef(p, "receiver type")
	if err != nil {
		return nil, err
	}

	if !p.MatchNext(token.RPARENT) {
		return nil, expected(p, token.RPARENT)
	}

	return &Param{Name: name, Type: receiverType}, nil
}

func parseParams(p parser.Parser) ([]Param, error) {
	if !p.MatchNext(token.LPARENT) {
		return nil, expected(p, token.LPARENT)
	}

	var params []Param
	for !p.MatchNext(token.RPARENT) {
		name, err := parseIdent(p, "parameter name")
		if err != nil {
			return nil, err
		}

		param := Param{Name: name}
		if p.MatchNext(token.COLON) {
			paramType, err := parseTypeRef(p, "parameter type")
			if err != nil {
				return nil, err
			}
			param.Type = paramType
		}
		params = append(params, param)

		if p.MatchNext(token.RPARENT) {
			break
		}
		if !p.MatchNext(token.COMMA) {
			return nil, expected(p, token.COMMA)
		}
	}

	return params, nil
}

func parseReturnStmt(p parser.Parser) (ast.Expr, error) {
	p.Next() // skip return

	if p.Match(token.RBRACE) || p.Match(token.SEMICOLON) {
		return NewReturnStmt(nil), nil
	}

	expr, err := p.ParseExpr(parser.Lowest)
	if err != nil {
		return nil, err
	}

	return NewReturnStmt(expr), nil
}

func parseTypeRef(p parser.Parser, label string) (*ast.TypeRef, error) {
	isPtr := p.MatchNext(token.STAR)

	if !p.Match(token.IDENT) && !p.CurrentToken().Type.IsType() {
		return nil, fmt.Errorf(
			"expected %s, got %s at %d:%d",
			label,
			p.CurrentToken().Type,
			p.CurrentToken().Line,
			p.CurrentToken().Col,
		)
	}

	tok := p.Next()
	return &ast.TypeRef{Name: tok.Literal, IsPtr: isPtr}, nil
}

func parseIdent(p parser.Parser, label string) (Ident, error) {
	if !p.Match(token.IDENT) {
		return "", fmt.Errorf(
			"expected %s, got %s at %d:%d",
			label,
			p.CurrentToken().Type,
			p.CurrentToken().Line,
			p.CurrentToken().Col,
		)
	}

	tok := p.Next()
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
