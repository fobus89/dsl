// Package lexer implements lexical analysis for the foo_lang programming language.
// It tokenizes source code into a stream of tokens for parsing.
package lexer

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/fobus89/dsl/token"
)

type Lexer struct {
	input []rune
	len   int
	pos   int
	line  int
	col   int
}

func NewLexer(s string) *Lexer {
	input := []rune(s)

	return &Lexer{
		input: input,
		len:   len(input),
		pos:   0,
		line:  0,
		col:   0,
	}
}

func IsSymbol(r rune) bool {
	_, ok := token.Symbol(string(r))
	return ok
}

func (l *Lexer) Get(startPos, endPos int) []rune {
	return l.input[startPos:endPos]
}

func (l *Lexer) Token() token.Token {
	l.SkipSpace()

	ch := l.Current()

	switch {

	case ch == '$':
		return l.readCompileTime()
	case unicode.IsDigit(ch):
		return l.ReadNumber()
	case ch == '"':
		return l.ReadStringFormat()
	case unicode.IsLetter(ch) || ch == '_':
		return l.ReadReservedToken()
	case IsSymbol(ch):
		return l.ReadSymbol()
	}

	if ch == 0x00 {
		return token.NewEOF(l.line, l.col)
	}

	return token.NewTokenPositioned(token.ILLEGAL, string(l.Next()), l.line, l.col)
}

func (l *Lexer) ReadStringFormat() token.Token {
	var (
		line = l.line
		col  = l.col
	)

	l.Next()

	var (
		buffer           strings.Builder
		hasInterpolation bool
		tokens           []token.Token
	)

	for !l.isEOF() {
		if l.MatchAndNext('"') {

			tokens = append(tokens, token.NewTokenPositioned(token.STRING_LITERAL, buffer.String(), l.line, l.col))
			buffer.Reset()

			break
		}

		switch {

		case l.MatchAllNext('\\', 'r'):
			buffer.WriteByte('\r')

		case l.MatchAllNext('\\', 'n'):
			buffer.WriteByte('\n')

		case l.MatchAllNext('\\', 't'):
			buffer.WriteByte('\t')

		case l.MatchAllNext('\\', '\\'):
			buffer.WriteByte('\\')

		case l.MatchAllNext('\\', '{'):
			buffer.WriteByte('{')

		case l.MatchAllNext('\\', '}'):
			buffer.WriteByte('}')

		case l.MatchAllNext('\\', '"'):
			buffer.WriteByte('"')
		case l.MatchAndNext('{'):
			hasInterpolation = true

			if buffer.Len() > 0 {
				tokens = append(tokens, token.NewTokenPositioned(token.STRING_LITERAL, buffer.String(), l.line, l.col))
				buffer.Reset()
			}

			braceCount := 1

			for !l.isEOF() && braceCount > 0 {
				t := l.Token()

				if t.Type == token.LBRACE {
					braceCount++
				} else if t.Type == token.RBRACE {
					braceCount--
					if braceCount == 0 {
						break
					}
				}

				tokens = append(tokens, token.NewTokenPositioned(t.Type, t.Literal, l.line, l.col))
			}

		default:
			buffer.WriteRune(l.Next())
		}
	}

	if !hasInterpolation {
		return tokens[0]
	}

	stringTok := token.NewTokenPositioned(token.STRING_FORMAT, "", line, col, tokens...)

	stringTok.Add(token.NewTokenPositioned(token.STRING_FORMAT, "", l.line, l.col))

	return stringTok
}

func (l *Lexer) readCompileTime() token.Token {
	ch := l.Next() // consume '$'

	if l.Match('"') {
		// $"i{f}" → COMPTIME_IDENT with string interpolation
		tok := l.ReadStringFormat()

		comptime := token.NewTokenPositioned(token.COMPTIME_IDENT, string(ch), l.line, l.col, tok)

		// comptime.Tokens = append(comptime.Tokens, tokens...)

		return comptime
	}

	// Read identifier or keyword after $
	if unicode.IsLetter(l.Current()) || l.Current() == '_' {
		comptime := token.NewTokenPositioned(token.COMPTIME_IDENT, string(ch), l.line, l.col, l.ReadReservedToken())
		return comptime
	}

	return token.NewTokenPositioned(token.ILLEGAL, string(ch), l.line, l.col)
}

func (l *Lexer) ReadReservedToken() token.Token {
	startPos := l.pos

	filter := func(r rune) bool {
		return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
	}

	for l.MatchFuncNext(filter) {
	}

	symbols := string(l.Get(startPos, l.pos))
	{
		if len(symbols) == 0 {
			return token.NewTokenPositioned(token.ILLEGAL, "", l.line, l.col)
		}
	}

	tok, ok := token.LookupReservedToken(symbols)
	{
		if !ok {
			return token.NewTokenPositioned(token.IDENT, symbols, l.line, l.col)
		}
	}

	value := l.Get(startPos, l.pos)

	return token.NewTokenPositioned(tok, string(value), l.line, l.col)
}

func (l *Lexer) ReadNumber() token.Token {
	startPos := l.pos

	isBin := func(r rune) bool {
		return (r == 'b' || r == 'B') || r == '0' || r == '1' || r == '_'
	}

	isOct := func(r rune) bool {
		return (r == 'o' || r == 'O') || (r >= '0' && r <= '7') || r == '_'
	}

	isHex := func(r rune) bool {
		return (r == 'x' || r == 'X') ||
			(r >= '0' && r <= '9') ||
			(r >= 'a' && r <= 'f') ||
			(r >= 'A' && r <= 'F')
	}

	var hasDot bool

	isDec := func(r rune) bool {
		if r == '.' && l.Peek(1) == '.' {
			return false
		}

		if r == '.' && unicode.IsLetter(l.Peek(1)) {
			return false
		}

		if r == '.' {
			hasDot = true
		}

		return (r >= '0' && r <= '9') || r == '.' || r == 'e' || r == 'E' || r == '_'
	}

	var filter func(rune) bool
	var is string

	switch {
	case l.MatchAll('0', 'b') || l.MatchAll('o', 'B'):
		filter = isBin
		is = "bin"
	case l.MatchAll('0', 'o') || l.MatchAll('o', 'O'):
		filter = isOct
		is = "oct"
	case l.MatchAll('0', 'x') || l.MatchAll('o', 'X'):
		filter = isHex
		is = "hex"
	default:
		filter = isDec
		is = "dec"
	}

	for l.MatchFuncNext(filter) {
	}

	value := string(l.Get(startPos, l.pos))
	tok := token.ILLEGAL

	switch is {
	case "bin":
		if _, err := strconv.ParseInt(value, 0, 64); err == nil {
			tok = token.INT_LITERAL
		}

	case "oct":
		if _, err := strconv.ParseInt(value, 0, 64); err == nil {
			tok = token.INT_LITERAL
		}

	case "hex":
		if _, err := strconv.ParseInt(value, 0, 64); err == nil {
			tok = token.INT_LITERAL
		}

	case "dec":
		if hasDot {
			if _, err := strconv.ParseFloat(value, 64); err == nil {
				tok = token.FLOAT_LITERAL
			}
		} else {
			if _, err := strconv.ParseInt(value, 0, 64); err == nil {
				tok = token.INT_LITERAL
			}
		}
	}

	return token.NewTokenPositioned(tok, value, l.line, l.col)
}

func (l *Lexer) Tokens() []token.Token {
	var tokens []token.Token

	for !l.isEOF() {
		tok := l.Token()

		if tok.Type == token.EOF {
			break
		}

		if tok.Len() > 0 {
			tokens = append(tokens, tok.Join()...)
		} else {
			tokens = append(tokens, tok)
		}
	}

	return append(tokens, token.NewEOF(l.line, l.col))
}

func (l *Lexer) ReadSymbol() token.Token {
	for _, symbol := range token.Symbols() {
		if l.MatchAllNext([]rune(symbol)...) {
			return token.NewTokenPositioned(token.StringToToken(symbol), symbol, l.line, l.col)
		}
	}

	return token.NewTokenPositioned(token.ILLEGAL, "", l.line, l.col)
}
