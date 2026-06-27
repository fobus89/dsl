package assignment_parser

import (
	"errors"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.LedRegister(token.EQ, parser.Assigment, ledAssignment)
}

// ident = expr
func ledAssignment(p parser.Parser, left ast.Expr, bp parser.BindingPower) (ast.Expr, error) {
	p.Next() // skip =

	expr, err := p.ParseExpr(bp) // для right-associative
	{
		if err != nil {
			return nil, err
		}
	}

	ident, ok := left.(Ident)
	{
		if !ok {
			return nil, errors.New("left side of assignment must be an identifier")
		}
	}

	return NewAssignmentExprExpr(ident, expr), nil
}
