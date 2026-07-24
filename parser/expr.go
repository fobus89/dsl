package parser

import (
	"fmt"
	"slices"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/token"
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
	p.exprDepth++
	defer func() {
		p.exprDepth--
	}()

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
		if p.exprDepth == p.stopDepth &&
			slices.Contains(p.stopTokens, tokKind) {
			break
		}
		if tokKind == token.LBRACE &&
			!p.isStructLiteralTarget(left) {
			break
		}

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

func (p *parser) isStructLiteralTarget(expr ast.Expr) bool {
	if target, ok := expr.(interface {
		IsStructLiteralTarget(ast.Ctx) bool
	}); ok && target.IsStructLiteralTarget(p.ctx) {
		return true
	}

	target, ok := expr.(interface {
		StructTypeName() string
	})
	if !ok {
		return false
	}

	def, declared := p.ctx.GetType(target.StructTypeName())
	return declared && def.Kind == ast.StructType
}

func (p *parser) ParseExprUntil(
	bp BindingPower,
	stops ...token.TokenType,
) (ast.Expr, error) {
	oldStopDepth := p.stopDepth
	oldStopTokens := p.stopTokens

	p.stopDepth = p.exprDepth + 1
	p.stopTokens = stops

	defer func() {
		p.stopDepth = oldStopDepth
		p.stopTokens = oldStopTokens
	}()

	return p.ParseExpr(bp)
}
