package parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
)

func (p *parser) ParseStmt() (ast.Expr, error) {
	if stmtHandler, ok := p.StmtOrNone(p.CurrentTokenKind()); ok {
		return stmtHandler(p)
	}

	expr, err := p.ParseExpr(Lowest)
	{
		if err != nil {
			return nil, err
		}
	}

	return expr, nil
}

func (p *parser) ParseExpr(bp BindingPower) (ast.Expr, error) {
	tokKind := p.CurrentTokenKind()

	nudHandler, ok := p.NudOrNone(tokKind)
	{
		if !ok {
			return nil, fmt.Errorf("%s not found", tokKind)
		}
	}

	left, err := nudHandler(p)
	{
		if err != nil {
			return nil, err
		}
	}

	for {

		tokKind = p.CurrentTokenKind()

		curBp := p.Bp(tokKind)
		{
			if curBp <= bp {
				break
			}
		}

		ledHandler, ok := p.LedOrNone(tokKind)
		{
			if !ok {
				return nil, fmt.Errorf("%s not found", tokKind)
			}
		}

		left, err = ledHandler(p, left, curBp)
		{
			if err != nil {
				return nil, err
			}
		}
	}

	return left, nil
}
