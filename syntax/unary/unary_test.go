package unary_parser_test

import (
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	logical_parser "github.com/fobus89/dsl/syntax/logical"
	unary_parser "github.com/fobus89/dsl/syntax/unary"
	"github.com/fobus89/dsl/value"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newUnaryTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	unary_parser.RegisterParser(p)
	logical_parser.RegisterParser(p)

	return p
}

func newUnaryProgramTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	assignment_parser.RegisterParser(p)
	unary_parser.RegisterParser(p)
	logical_parser.RegisterParser(p)

	return p
}

func TestUnaryMinus(t *testing.T) {
	p := newUnaryTestParser(`-5`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if got.UnsafeCastFloat64() != -5 {
		t.Fatalf("expected -5, got %#v", got.Any())
	}
}

func TestUnaryBang(t *testing.T) {
	p := newUnaryTestParser(`!true`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if got.UnsafeCastBool() {
		t.Fatalf("expected false, got %#v", got.Any())
	}
}

func TestUnaryBangChain(t *testing.T) {
	p := newUnaryTestParser(`!!!!!true`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if got.UnsafeCastBool() {
		t.Fatalf("expected false, got %#v", got.Any())
	}
}

func TestUnaryBangChainEven(t *testing.T) {
	p := newUnaryTestParser(`!!!!true`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if !got.UnsafeCastBool() {
		t.Fatalf("expected true, got %#v", got.Any())
	}
}

func TestUnaryPrecedence(t *testing.T) {
	p := newUnaryTestParser(`-1 * 2`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if got.UnsafeCastFloat64() != -2 {
		t.Fatalf("expected -2, got %#v", got.Any())
	}
}

func TestUnaryMinusAfterBinaryStatement(t *testing.T) {
	p := newUnaryProgramTestParser(`
		user1 any users
		r = -1
	`)

	p.Ctx().SetValue("user1", value.NewType(map[string]any{"id": 1}))
	p.Ctx().SetValue("users", value.NewType([]map[string]any{{"id": 1}}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	for _, expr := range exprs {
		_, err := expr.Eval(p.Ctx())
		if err != nil {
			t.Fatal(err)
		}
	}

	got, ok := p.Ctx().GetValue("r")
	if !ok {
		t.Fatal("expected r in context")
	}

	if got.UnsafeCastFloat64() != -1 {
		t.Fatalf("expected -1, got %#v", got.Any())
	}
}
