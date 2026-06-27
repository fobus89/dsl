package lexer

import (
	"testing"

	"github.com/fobus89/dsl/token"
)

func tokensEqual(a, b []token.Token) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Type != b[i].Type ||
			a[i].Literal != b[i].Literal {
			return false
		}
	}
	return true
}

func Test_Lexer_StringFormat(t *testing.T) {
	input := `
		"{1+2+3}"
	`
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.STRING_FORMAT, ""),

		token.NewToken(token.INT_LITERAL, "1"),
		token.NewToken(token.PLUS, "+"),
		token.NewToken(token.INT_LITERAL, "2"),
		token.NewToken(token.PLUS, "+"),
		token.NewToken(token.INT_LITERAL, "3"),

		token.NewToken(token.STRING_LITERAL, ""),

		token.NewToken(token.STRING_FORMAT, ""),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_SimpleString(t *testing.T) {
	input := `"hello world"`
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.STRING_LITERAL, "hello world"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_StringWithEscapes(t *testing.T) {
	input := `"hello\nworld\t\"test\""`
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.STRING_LITERAL, "hello\nworld\t\"test\""),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_Numbers_Decimal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected token.TokenType
		literal  string
	}{
		{"integer", "123", token.INT_LITERAL, "123"},
		{"float", "123.456", token.FLOAT_LITERAL, "123.456"},
		{"with_underscore", "1_000_000", token.INT_LITERAL, "1_000_000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			got := l.Tokens()

			if len(got) != 2 {
				t.Fatalf("expected 2 tokens, got %d", len(got))
			}

			if got[0].Type != tt.expected {
				t.Errorf("expected type %v, got %v", tt.expected, got[0].Type)
			}

			if got[0].Literal != tt.literal {
				t.Errorf("expected literal %q, got %q", tt.literal, got[0].Literal)
			}
		})
	}
}

func Test_Lexer_Numbers_Binary(t *testing.T) {
	input := "0b1010"
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.INT_LITERAL, "0b1010"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_Numbers_Octal(t *testing.T) {
	input := "0o755"
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.INT_LITERAL, "0o755"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_Numbers_Hex(t *testing.T) {
	input := "0xFF"
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.INT_LITERAL, "0xFF"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_Identifiers(t *testing.T) {
	input := "foo bar _test test123"
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.IDENT, "foo"),
		token.NewToken(token.IDENT, "bar"),
		token.NewToken(token.IDENT, "_test"),
		token.NewToken(token.IDENT, "test123"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_Operators(t *testing.T) {
	input := "+ - * / == != < > <= >="
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.PLUS, "+"),
		token.NewToken(token.MINUS, "-"),
		token.NewToken(token.STAR, "*"),
		token.NewToken(token.SLASH, "/"),
		token.NewToken(token.EQ_EQ, "=="),
		token.NewToken(token.BANG_EQ, "!="),
		token.NewToken(token.LT, "<"),
		token.NewToken(token.GT, ">"),
		token.NewToken(token.LT_EQ, "<="),
		token.NewToken(token.GT_EQ, ">="),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_CompileTime_Ident(t *testing.T) {
	input := "$foo"
	l := NewLexer(input)

	tok := l.Token()

	if tok.Type != token.COMPTIME_IDENT {
		t.Errorf("expected COMPTIME_IDENT, got %v", tok.Type)
	}

	if tok.Literal != "$" {
		t.Errorf("expected literal '$', got %q", tok.Literal)
	}

	if tok.Len() != 1 {
		t.Errorf("expected 1 child token, got %d", tok.Len())
	}
}

func Test_Lexer_CompileTime_String(t *testing.T) {
	input := `$"test"`
	l := NewLexer(input)

	tok := l.Token()

	if tok.Type != token.COMPTIME_IDENT {
		t.Errorf("expected COMPTIME_IDENT, got %v", tok.Type)
	}

	if tok.Len() != 1 {
		t.Errorf("expected 1 child token, got %d", tok.Len())
	}
}

func Test_Lexer_MixedExpression(t *testing.T) {
	input := `x = 10 + 20 * 30`
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.IDENT, "x"),
		token.NewToken(token.EQ, "="),
		token.NewToken(token.INT_LITERAL, "10"),
		token.NewToken(token.PLUS, "+"),
		token.NewToken(token.INT_LITERAL, "20"),
		token.NewToken(token.STAR, "*"),
		token.NewToken(token.INT_LITERAL, "30"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_Braces(t *testing.T) {
	input := "{ [ ( ) ] }"
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.LBRACE, "{"),
		token.NewToken(token.LBRACKET, "["),
		token.NewToken(token.LPARENT, "("),
		token.NewToken(token.RPARENT, ")"),
		token.NewToken(token.RBRACKET, "]"),
		token.NewToken(token.RBRACE, "}"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_EmptyInput(t *testing.T) {
	input := ""
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_WhitespaceOnly(t *testing.T) {
	input := "   \n\t  \n  "
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_ComplexExpression(t *testing.T) {
	input := `1+2*(2+3-(1+2+foo + bar()))`
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.INT_LITERAL, "1"),
		token.NewToken(token.PLUS, "+"),
		token.NewToken(token.INT_LITERAL, "2"),
		token.NewToken(token.STAR, "*"),
		token.NewToken(token.LPARENT, "("),
		token.NewToken(token.INT_LITERAL, "2"),
		token.NewToken(token.PLUS, "+"),
		token.NewToken(token.INT_LITERAL, "3"),
		token.NewToken(token.MINUS, "-"),
		token.NewToken(token.LPARENT, "("),
		token.NewToken(token.INT_LITERAL, "1"),
		token.NewToken(token.PLUS, "+"),
		token.NewToken(token.INT_LITERAL, "2"),
		token.NewToken(token.PLUS, "+"),
		token.NewToken(token.IDENT, "foo"),
		token.NewToken(token.PLUS, "+"),
		token.NewToken(token.IDENT, "bar"),
		token.NewToken(token.LPARENT, "("),
		token.NewToken(token.RPARENT, ")"),
		token.NewToken(token.RPARENT, ")"),
		token.NewToken(token.RPARENT, ")"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_ForLoop(t *testing.T) {
	input := `for let x = 1; x < 10; x++ { }`
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.FOR, "for"),
		token.NewToken(token.LET, "let"),
		token.NewToken(token.IDENT, "x"),
		token.NewToken(token.EQ, "="),
		token.NewToken(token.INT_LITERAL, "1"),
		token.NewToken(token.SEMICOLON, ";"),
		token.NewToken(token.IDENT, "x"),
		token.NewToken(token.LT, "<"),
		token.NewToken(token.INT_LITERAL, "10"),
		token.NewToken(token.SEMICOLON, ";"),
		token.NewToken(token.IDENT, "x"),
		token.NewToken(token.PLUS_PLUS, "++"),
		token.NewToken(token.LBRACE, "{"),
		token.NewToken(token.RBRACE, "}"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_IfStatement(t *testing.T) {
	input := `if x > 5 { return true } else { return false }`
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.IF, "if"),
		token.NewToken(token.IDENT, "x"),
		token.NewToken(token.GT, ">"),
		token.NewToken(token.INT_LITERAL, "5"),
		token.NewToken(token.LBRACE, "{"),
		token.NewToken(token.RETURN, "return"),
		token.NewToken(token.TRUE, "true"),
		token.NewToken(token.RBRACE, "}"),
		token.NewToken(token.ELSE, "else"),
		token.NewToken(token.LBRACE, "{"),
		token.NewToken(token.RETURN, "return"),
		token.NewToken(token.FALSE, "false"),
		token.NewToken(token.RBRACE, "}"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_FunctionDeclaration(t *testing.T) {
	input := `fn add(a, b) { return a + b }`
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.FN, "fn"),
		token.NewToken(token.IDENT, "add"),
		token.NewToken(token.LPARENT, "("),
		token.NewToken(token.IDENT, "a"),
		token.NewToken(token.COMMA, ","),
		token.NewToken(token.IDENT, "b"),
		token.NewToken(token.RPARENT, ")"),
		token.NewToken(token.LBRACE, "{"),
		token.NewToken(token.RETURN, "return"),
		token.NewToken(token.IDENT, "a"),
		token.NewToken(token.PLUS, "+"),
		token.NewToken(token.IDENT, "b"),
		token.NewToken(token.RBRACE, "}"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_NestedStructures(t *testing.T) {
	input := `fn calc() { for let i = 0; i < 10; i++ { if i > 5 { return i } } }`
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.FN, "fn"),
		token.NewToken(token.IDENT, "calc"),
		token.NewToken(token.LPARENT, "("),
		token.NewToken(token.RPARENT, ")"),
		token.NewToken(token.LBRACE, "{"),
		token.NewToken(token.FOR, "for"),
		token.NewToken(token.LET, "let"),
		token.NewToken(token.IDENT, "i"),
		token.NewToken(token.EQ, "="),
		token.NewToken(token.INT_LITERAL, "0"),
		token.NewToken(token.SEMICOLON, ";"),
		token.NewToken(token.IDENT, "i"),
		token.NewToken(token.LT, "<"),
		token.NewToken(token.INT_LITERAL, "10"),
		token.NewToken(token.SEMICOLON, ";"),
		token.NewToken(token.IDENT, "i"),
		token.NewToken(token.PLUS_PLUS, "++"),
		token.NewToken(token.LBRACE, "{"),
		token.NewToken(token.IF, "if"),
		token.NewToken(token.IDENT, "i"),
		token.NewToken(token.GT, ">"),
		token.NewToken(token.INT_LITERAL, "5"),
		token.NewToken(token.LBRACE, "{"),
		token.NewToken(token.RETURN, "return"),
		token.NewToken(token.IDENT, "i"),
		token.NewToken(token.RBRACE, "}"),
		token.NewToken(token.RBRACE, "}"),
		token.NewToken(token.RBRACE, "}"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_struct(t *testing.T) {
	input := `
struct Node {
value: Int,
next: Node
}`

	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		token.NewToken(token.Struct, "struct"),

		token.NewToken(token.IDENT, "Node"),
		token.NewToken(token.LBRACE, "{"),

		token.NewToken(token.IDENT, "value"),
		token.NewToken(token.COLON, ":"),
		token.NewToken(token.IDENT, "Int"),

		token.NewToken(token.COMMA, ","),

		token.NewToken(token.IDENT, "next"),
		token.NewToken(token.COLON, ":"),
		token.NewToken(token.IDENT, "Node"),

		token.NewToken(token.RBRACE, "}"),
		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}

func Test_Lexer_LinkedListProgram(t *testing.T) {
	input := `
struct Node {
	value: Int,
	next: Node
}

fn create_node(val: Int) -> Node {
	return Node { value: val, next: nil }
}

fn append(head: Node, val: Int) {
	let mut current = head
	while current.next != nil {
		current = current.next
	}
	current.next = create_node(val)
}

fn print_list(head: Node) {
	let mut current = head
	while current != nil {
		println(current.value)
		current = current.next
	}
}
`
	l := NewLexer(input)

	got := l.Tokens()

	expected := []token.Token{
		// struct Node {
		token.NewToken(token.Struct, "struct"),
		token.NewToken(token.IDENT, "Node"),
		token.NewToken(token.LBRACE, "{"),
		// value: Int,
		token.NewToken(token.IDENT, "value"),
		token.NewToken(token.COLON, ":"),
		token.NewToken(token.IDENT, "Int"),
		token.NewToken(token.COMMA, ","),
		// next: Node
		token.NewToken(token.IDENT, "next"),
		token.NewToken(token.COLON, ":"),
		token.NewToken(token.IDENT, "Node"),
		// }
		token.NewToken(token.RBRACE, "}"),

		// fn create_node(val: Int) -> Node {
		token.NewToken(token.FN, "fn"),
		token.NewToken(token.IDENT, "create_node"),
		token.NewToken(token.LPARENT, "("),
		token.NewToken(token.IDENT, "val"),
		token.NewToken(token.COLON, ":"),
		token.NewToken(token.IDENT, "Int"),
		token.NewToken(token.RPARENT, ")"),
		token.NewToken(token.MINUS_GT, "->"),
		token.NewToken(token.IDENT, "Node"),
		token.NewToken(token.LBRACE, "{"),
		// return Node { value: val, next: nil }
		token.NewToken(token.RETURN, "return"),
		token.NewToken(token.IDENT, "Node"),
		token.NewToken(token.LBRACE, "{"),
		token.NewToken(token.IDENT, "value"),
		token.NewToken(token.COLON, ":"),
		token.NewToken(token.IDENT, "val"),
		token.NewToken(token.COMMA, ","),
		token.NewToken(token.IDENT, "next"),
		token.NewToken(token.COLON, ":"),
		token.NewToken(token.NIL, "nil"),
		token.NewToken(token.RBRACE, "}"),
		// }
		token.NewToken(token.RBRACE, "}"),

		// fn append(head: Node, val: Int) {
		token.NewToken(token.FN, "fn"),
		token.NewToken(token.IDENT, "append"),
		token.NewToken(token.LPARENT, "("),
		token.NewToken(token.IDENT, "head"),
		token.NewToken(token.COLON, ":"),
		token.NewToken(token.IDENT, "Node"),
		token.NewToken(token.COMMA, ","),
		token.NewToken(token.IDENT, "val"),
		token.NewToken(token.COLON, ":"),
		token.NewToken(token.IDENT, "Int"),
		token.NewToken(token.RPARENT, ")"),
		token.NewToken(token.LBRACE, "{"),
		// let mut current = head
		token.NewToken(token.LET, "let"),
		token.NewToken(token.MUT, "mut"),
		token.NewToken(token.IDENT, "current"),
		token.NewToken(token.EQ, "="),
		token.NewToken(token.IDENT, "head"),
		// while current.next != nil {
		token.NewToken(token.WHILE, "while"),
		token.NewToken(token.IDENT, "current"),
		token.NewToken(token.DOT, "."),
		token.NewToken(token.IDENT, "next"),
		token.NewToken(token.BANG_EQ, "!="),
		token.NewToken(token.NIL, "nil"),
		token.NewToken(token.LBRACE, "{"),
		// current = current.next
		token.NewToken(token.IDENT, "current"),
		token.NewToken(token.EQ, "="),
		token.NewToken(token.IDENT, "current"),
		token.NewToken(token.DOT, "."),
		token.NewToken(token.IDENT, "next"),
		// }
		token.NewToken(token.RBRACE, "}"),
		// current.next = create_node(val)
		token.NewToken(token.IDENT, "current"),
		token.NewToken(token.DOT, "."),
		token.NewToken(token.IDENT, "next"),
		token.NewToken(token.EQ, "="),
		token.NewToken(token.IDENT, "create_node"),
		token.NewToken(token.LPARENT, "("),
		token.NewToken(token.IDENT, "val"),
		token.NewToken(token.RPARENT, ")"),
		// }
		token.NewToken(token.RBRACE, "}"),

		// fn print_list(head: Node) {
		token.NewToken(token.FN, "fn"),
		token.NewToken(token.IDENT, "print_list"),
		token.NewToken(token.LPARENT, "("),
		token.NewToken(token.IDENT, "head"),
		token.NewToken(token.COLON, ":"),
		token.NewToken(token.IDENT, "Node"),
		token.NewToken(token.RPARENT, ")"),
		token.NewToken(token.LBRACE, "{"),
		// let mut current = head
		token.NewToken(token.LET, "let"),
		token.NewToken(token.MUT, "mut"),
		token.NewToken(token.IDENT, "current"),
		token.NewToken(token.EQ, "="),
		token.NewToken(token.IDENT, "head"),
		// while current != nil {
		token.NewToken(token.WHILE, "while"),
		token.NewToken(token.IDENT, "current"),
		token.NewToken(token.BANG_EQ, "!="),
		token.NewToken(token.NIL, "nil"),
		token.NewToken(token.LBRACE, "{"),
		// println(current.value)
		token.NewToken(token.PRINTLN, "println"),
		token.NewToken(token.LPARENT, "("),
		token.NewToken(token.IDENT, "current"),
		token.NewToken(token.DOT, "."),
		token.NewToken(token.IDENT, "value"),
		token.NewToken(token.RPARENT, ")"),
		// current = current.next
		token.NewToken(token.IDENT, "current"),
		token.NewToken(token.EQ, "="),
		token.NewToken(token.IDENT, "current"),
		token.NewToken(token.DOT, "."),
		token.NewToken(token.IDENT, "next"),
		// }
		token.NewToken(token.RBRACE, "}"),
		// }
		token.NewToken(token.RBRACE, "}"),

		token.NewToken(token.EOF, ""),
	}

	if !tokensEqual(got, expected) {
		t.Fatalf("tokens mismatch:\n got=%#v\n want=%#v", got, expected)
	}
}
