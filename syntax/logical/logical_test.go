package logical_parser_test

import (
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	logical_parser "github.com/fobus89/dsl/syntax/logical"
	"github.com/fobus89/dsl/value"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newLogicalTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	logical_parser.RegisterParser(p)

	return p
}

func TestLogicalOrCastsAnyValueToBool(t *testing.T) {
	p := newLogicalTestParser(`r || true`)
	p.Ctx().SetValue("r", value.NewType(map[string]any{}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if !got.UnsafeCastBool() {
		t.Fatalf("expected truthy map || true")
	}
}

func TestLogicalOrWithFalsyValues(t *testing.T) {
	p := newLogicalTestParser(`r || false`)
	p.Ctx().SetValue("r", value.NewTypeUndefined())

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if got.UnsafeCastBool() {
		t.Fatalf("expected undefined || false to be false")
	}
}
