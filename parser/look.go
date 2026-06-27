package parser

import (
	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/token"
)

type BindingPower int

const (
	Lowest BindingPower = iota
	Comma
	Assigment
	Logical
	Relational
	Additive
	Muptiplicative
	Unary
	Call
	Member
	Primary
	Highest
)

// TODO: func(ctx CtxScope, p Parser)
type (
	StmtHandlerType = func(p Parser) (ast.Expr, error)
	NudHandlerType  = func(p Parser) (ast.Expr, error)
	LedHandlerType  = func(p Parser, left ast.Expr, bp BindingPower) (ast.Expr, error)
)

type Handler[T comparable, E any] map[T]E

func (h Handler[T, E]) Get(key T) E {
	handler := h[key]
	return handler
}

func (h Handler[T, E]) GetOrNone(key T) (E, bool) {
	handler, ok := h[key]
	return handler, ok
}

func (h Handler[T, E]) Has(key T) bool {
	_, ok := h[key]
	return ok
}

func (h Handler[T, E]) Add(key T, val E) {
	h[key] = val
}

func (p *parser) NudRegister(kind token.TokenType, nudHander NudHandlerType) {
	p.nudLookup.Add(kind, nudHander)
}

func (p *parser) LedRegister(kind token.TokenType, bp BindingPower, ledHander LedHandlerType) {
	p.bpLookup.Add(kind, bp)
	p.ledLookup.Add(kind, ledHander)
}

func (p *parser) StmtRegister(kind token.TokenType, stmtHander StmtHandlerType) {
	p.bpLookup.Add(kind, Lowest)
	p.stmtLookup.Add(kind, stmtHander)
}

func (p *parser) StmtOrNone(kind token.TokenType) (StmtHandlerType, bool) {
	return p.stmtLookup.GetOrNone(kind)
}

func (p *parser) BpOrNone(kind token.TokenType) (BindingPower, bool) {
	return p.bpLookup.GetOrNone(kind)
}

func (p *parser) NudOrNone(kind token.TokenType) (NudHandlerType, bool) {
	return p.nudLookup.GetOrNone(kind)
}

func (p *parser) LedOrNone(kind token.TokenType) (LedHandlerType, bool) {
	return p.ledLookup.GetOrNone(kind)
}

func (p *parser) Stmt(kind token.TokenType) StmtHandlerType {
	return p.stmtLookup.Get(kind)
}

func (p *parser) Bp(kind token.TokenType) BindingPower {
	return p.bpLookup.Get(kind)
}

func (p *parser) Nud(kind token.TokenType) NudHandlerType {
	return p.nudLookup.Get(kind)
}

func (p *parser) Led(kind token.TokenType) LedHandlerType {
	return p.ledLookup.Get(kind)
}
