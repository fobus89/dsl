package comparison_parser_test

import (
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	comparison_parser "github.com/fobus89/dsl/syntax/comparison"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newComparisonTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	comparison_parser.RegisterParser(p)

	return p
}

func TestComparison(t *testing.T) {
	p := newComparisonTestParser(`1 < 2`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if !got.UnsafeCastBool() {
		t.Fatalf("expected true")
	}
}

func TestComparisonEquality(t *testing.T) {
	p := newComparisonTestParser(`1 == 1.0`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if !got.UnsafeCastBool() {
		t.Fatalf("expected true")
	}
}
