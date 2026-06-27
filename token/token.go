package token

import (
	"fmt"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
	tokens  []Token
}

func (t Token) String() string {
	return fmt.Sprintf(
		"%s{ (%d:%d) %s }",
		t.Literal,
		t.Line,
		t.Col,
		t.Type,
	)
}

func NewEOF(line, col int) Token {
	return Token{
		Type:    EOF,
		Literal: "",
		Line:    line,
		Col:     col,
	}
}

func NewToken(t TokenType, value string) Token {
	return Token{
		Type:    t,
		Literal: value,
	}
}

func NewTokenPositioned(t TokenType, value string, line, col int, tokens ...Token) Token {
	return Token{
		Type:    t,
		Literal: value,
		Line:    line,
		Col:     col,
		tokens:  tokens,
	}
}

func (t *Token) Len() int {
	return len(t.tokens)
}

func (t *Token) Add(child Token) {
	t.tokens = append(t.tokens, child)
}

func (t Token) Join() []Token {
	tokens := t.tokens
	t.tokens = nil
	return append([]Token{t}, tokens...)
}
