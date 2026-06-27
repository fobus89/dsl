package parser

import (
	"slices"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/lexer"
	"github.com/fobus89/dsl/token"
	"github.com/fobus89/dsl/value"
)

type (
	StmtLookupType = Handler[token.TokenType, StmtHandlerType]
	NudLookupType  = Handler[token.TokenType, NudHandlerType]
	LedLookupType  = Handler[token.TokenType, LedHandlerType]
	BpLookupType   = Handler[token.TokenType, BindingPower]
)

type Parser interface {
	ParseExpr(bp BindingPower) (ast.Expr, error)
	ParseStmt() (ast.Expr, error)
	Next() token.Token
	HasToken() bool
	Peek(offset int) token.Token
	CurrentToken() token.Token
	MatchAllNext(tokens ...token.TokenType) bool
	MatchAll(tokens ...token.TokenType) bool
	MatchAny(tokens ...token.TokenType) bool
	MatchAnyNext(tokens ...token.TokenType) bool
	Match(tok token.TokenType) bool
	MatchNext(tok token.TokenType) bool
	MatchNextN(tok token.TokenType, n int) bool
	NudRegister(kind token.TokenType, nudHander NudHandlerType)
	LedRegister(kind token.TokenType, bp BindingPower, ledHander LedHandlerType)
	StmtRegister(kind token.TokenType, stmtHander StmtHandlerType)
	StmtOrNone(kind token.TokenType) (StmtHandlerType, bool)
	BpOrNone(kind token.TokenType) (BindingPower, bool)
	NudOrNone(kind token.TokenType) (NudHandlerType, bool)
	LedOrNone(kind token.TokenType) (LedHandlerType, bool)
	Stmt(kind token.TokenType) StmtHandlerType
	Bp(kind token.TokenType) BindingPower
	Nud(kind token.TokenType) NudHandlerType
	Led(kind token.TokenType) LedHandlerType

	//ctx
	SetValue(key string, val value.Type)
	GetValue(key string) (value.Type, bool)
	SetFunc(key string, val functype)
	GetFunc(key string) (functype, bool)
}

var _ Parser = (*parser)(nil)

type parser struct {
	Nodes      []token.Token
	stmtLookup StmtLookupType
	nudLookup  NudLookupType
	ledLookup  LedLookupType
	bpLookup   BpLookupType
	ctx        *scope
	pos        int
}

func (p *parser) GetFunc(key string) (functype, bool) {
	return p.ctx.GetFunc(key)
}

func (p *parser) SetFunc(key string, val functype) {
	p.ctx.SetFunc(key, val)
}

func (p *parser) GetValue(key string) (value.Type, bool) {
	return p.ctx.GetValue(key)
}

func (p *parser) SetValue(key string, val value.Type) {
	p.ctx.SetValue(key, val)
}

func NewParser(s string) *parser {
	l := lexer.NewLexer(s)
	return &parser{
		Nodes:      l.Tokens(),
		stmtLookup: StmtLookupType{},
		nudLookup:  NudLookupType{},
		ledLookup:  LedLookupType{},
		bpLookup:   BpLookupType{},
		ctx:        NewCtx(),
		pos:        0,
	}
}

func (p *parser) Parse() ([]ast.Expr, error) {
	var exprs []ast.Expr

	for p.HasToken() {
		stmt, err := p.ParseStmt()
		{
			if err != nil {
				return nil, err
			}
		}

		exprs = append(exprs, stmt)
	}

	return exprs, nil
}

func (p *parser) CurrentToken() token.Token {
	return p.Peek(0)
}

func (p *parser) Peek(offset int) token.Token {
	pos := p.pos + offset
	if pos < 0 || pos >= len(p.Nodes) {
		return token.NewToken(token.EOF, "")
	}
	return p.Nodes[pos]
}

func (p *parser) Next() token.Token {
	return p.NextN(1)
}

func (p *parser) NextN(n int) token.Token {
	tok := p.Peek(0)

	if tok.Type == token.EOF {
		return tok
	}

	p.pos += n

	return tok
}

func (p *parser) CurrentTokenKind() token.TokenType {
	return p.CurrentToken().Type
}

func (p *parser) HasToken() bool {
	return p.pos < len(p.Nodes) && p.CurrentTokenKind() != token.EOF
}

func (p *parser) MatchAll(tokens ...token.TokenType) bool {
	for i, tok := range tokens {
		if !p.MatchN(tok, i) {
			return false
		}
	}

	return true
}

func (p *parser) MatchAllNext(tokens ...token.TokenType) bool {
	if !p.MatchAll(tokens...) {
		return false
	}

	p.NextN(len(tokens))

	return true
}

func (p *parser) MatchAny(tokens ...token.TokenType) bool {
	return slices.ContainsFunc(tokens, p.Match)
}

func (p *parser) MatchAnyNext(tokens ...token.TokenType) bool {
	return slices.ContainsFunc(tokens, p.MatchNext)
}

func (p *parser) Match(tok token.TokenType) bool {
	return p.CurrentTokenKind() == tok
}

func (p *parser) MatchN(tok token.TokenType, pos int) bool {
	return p.Peek(pos).Type == tok
}

func (p *parser) MatchNext(tok token.TokenType) bool {
	return p.MatchNextN(tok, 1)
}

func (p *parser) MatchNextN(tok token.TokenType, n int) bool {
	if p.Match(tok) {
		p.NextN(n)
		return true
	}

	return false
}
