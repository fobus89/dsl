package generic_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.LedRegister(token.LBRACKET, parser.Call, parseGeneric)
}

func parseGeneric(
	p parser.Parser,
	left ast.Expr,
	_ parser.BindingPower,
) (ast.Expr, error) {
	p.Next() // skip [
	var args []ast.TypeRef
	for !p.MatchNext(token.RBRACKET) {
		arg, err := parser.ParseTypeRef(
			p,
			"generic type argument",
		)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		if p.MatchNext(token.RBRACKET) {
			break
		}
		if !p.MatchNext(token.COMMA) {
			return nil, fmt.Errorf(
				"expected , in generic arguments, got %s",
				p.CurrentToken().Type,
			)
		}
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("generic arguments cannot be empty")
	}
	expr := NewGenericExpr(left, args)
	expr.RegisterInstantiation(p.Ctx())
	return expr, nil
}
