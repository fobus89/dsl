package lexer

import (
	"slices"
	"unicode"
)

func (l *Lexer) isEOF() bool {
	return l.pos >= l.len
}

func (l *Lexer) Peek(n int) rune {
	idx := l.pos + n

	if idx >= l.len || idx < 0 {
		return 0
	}

	return l.input[idx]
}

// FIXME: cursor( line | col )
func (l *Lexer) advance() rune {
	if l.isEOF() {
		return 0
	}

	ch := l.input[l.pos]
	l.pos++

	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}

	return ch
}

func (l *Lexer) Current() rune {
	return l.Peek(0)
}

func (l *Lexer) Match(symbol rune) bool {
	return l.Current() == symbol
}

func (l *Lexer) MatchN(symbol rune, n int) bool {
	return l.Peek(n) == symbol
}

func (l *Lexer) NextN(n int) rune {
	var ch rune

	for range n {
		ch = l.advance()
	}

	return ch
}

func (l *Lexer) Next() rune {
	return l.advance()
}

func (l *Lexer) MatchNAndNext(symbol rune, pos int) bool {
	if l.MatchN(symbol, pos) {
		l.NextN(pos)
		return true
	}

	return false
}

func (l *Lexer) MatchAndNext(symbol rune) bool {
	filter := func(r rune) bool {
		return r == symbol
	}

	return l.MatchFuncNext(filter)
}

func (l *Lexer) MatchAll(symbols ...rune) bool {
	for i, symbol := range symbols {
		if !l.MatchN(symbol, i) {
			return false
		}
	}

	return true
}

func (l *Lexer) MatchAllNext(symbols ...rune) bool {
	if !l.MatchAll(symbols...) {
		return false
	}

	l.NextN(len(symbols))

	return true
}

func (l *Lexer) MatchAny(symbols ...rune) bool {
	filter := func(r rune) bool {
		return slices.Contains(symbols, r)
	}

	return l.MatchFunc(filter)
}

func (l *Lexer) MatchAnyNext(symbols ...rune) bool {
	filter := func(r rune) bool {
		return slices.Contains(symbols, r)
	}

	return l.MatchFuncNext(filter)
}

func (l *Lexer) MatchFunc(f func(rune) bool) bool {
	return f(l.Current())
}

func (l *Lexer) MatchFuncNext(f func(rune) bool) bool {
	if l.MatchFunc(f) {
		l.Next()
		return true
	}

	return false
}

func (l *Lexer) SkipSpace() {
	for unicode.IsSpace(l.Current()) {
		l.Next()
	}
}
