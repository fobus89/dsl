package literal_parser_test

import (
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	comparison_parser "github.com/fobus89/dsl/syntax/comparison"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newLiteralTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	comparison_parser.RegisterParser(p)

	return p
}

func TestNullLiteral(t *testing.T) {
	p := newLiteralTestParser(`null`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if !got.IsNil() {
		t.Fatalf("expected nil, got %#v", got.Any())
	}
}

func TestNullEqualsNil(t *testing.T) {
	p := newLiteralTestParser(`null == nil`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if !got.UnsafeCastBool() {
		t.Fatalf("expected null == nil")
	}
}

func TestNanEqualsNan(t *testing.T) {
	p := newLiteralTestParser(`nan == nan`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if !got.UnsafeCastBool() {
		t.Fatalf("expected nan == nan in DSL")
	}
}

func TestUndefinedLiteral(t *testing.T) {
	p := newLiteralTestParser(`undefined == undefind`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if !got.UnsafeCastBool() {
		t.Fatalf("expected undefined aliases to be equal")
	}
}
