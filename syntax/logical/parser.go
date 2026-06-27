package logical_parser

import (
	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	"github.com/fobus89/dsl/token"
)

func ledBinary(p parser.Parser, left ast.Expr, bp parser.BindingPower) (ast.Expr, error) {
	opToken := p.Next()

	right, err := p.ParseExpr(bp)
	{
		if err != nil {
			return nil, err
		}
	}

	return binary_parser.NewBinaryExpr(opToken.Type, left, right), nil
}

func RegisterParser(p parser.Parser) {
	//&& ||
	p.LedRegister(token.AMP_AMP, parser.Logical, ledBinary)
	p.LedRegister(token.PIPE_PIPE, parser.Logical, ledBinary)

	p.LedRegister(token.AND, parser.Logical, ledBinary)
	p.LedRegister(token.OR, parser.Logical, ledBinary)

	// > < != == <= >=
	p.LedRegister(token.GT, parser.Relational, ledBinary)
	p.LedRegister(token.LT, parser.Relational, ledBinary)
	p.LedRegister(token.GT_EQ, parser.Relational, ledBinary)
	p.LedRegister(token.LT_EQ, parser.Relational, ledBinary)
	p.LedRegister(token.EQ_EQ, parser.Relational, ledBinary)
	p.LedRegister(token.BANG_EQ, parser.Relational, ledBinary)
}
