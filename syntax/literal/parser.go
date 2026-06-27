package literal_parser

import (
	"strconv"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	"github.com/fobus89/dsl/token"
)

func RegisterParser(p parser.Parser) {
	p.NudRegister(token.INT_LITERAL, nudIntLiteral)
	p.NudRegister(token.FLOAT_LITERAL, nudFloat64Literal)
	p.NudRegister(token.FALSE, nudBoolLiteral)
	p.NudRegister(token.TRUE, nudBoolLiteral)
	p.NudRegister(token.STRING_LITERAL, nudStringLiteral)
	p.NudRegister(token.IDENT, nudIdentLiteral)
}

func nudIntLiteral(p parser.Parser) (ast.Expr, error) {
	literal := p.Next()

	numb, err := strconv.Atoi(literal.Literal)
	{
		if err != nil {
			return nil, err
		}
	}

	return NewIntExpr(numb), nil
}

func nudFloat64Literal(p parser.Parser) (ast.Expr, error) {
	literal := p.Next()

	numb, err := strconv.ParseFloat(literal.Literal, 64)
	{
		if err != nil {
			return nil, err
		}
	}

	return NewFloat64Expr(numb), nil
}

func nudBoolLiteral(p parser.Parser) (ast.Expr, error) {
	literal := p.Next()

	b, err := strconv.ParseBool(literal.Literal)
	{
		if err != nil {
			return nil, err
		}
	}

	return NewBoolExpr(b), nil
}

func nudStringLiteral(p parser.Parser) (ast.Expr, error) {
	literal := p.Next()
	return NewStringExpr(literal.Literal), nil
}

func nudIdentLiteral(p parser.Parser) (ast.Expr, error) {
	literal := p.Next()
	return NewIdentExpr(literal.Literal), nil
}
