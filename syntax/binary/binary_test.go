package binary_parser_test

import (
	"math"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	map_parser "github.com/fobus89/dsl/syntax/map"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newBinaryTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	map_parser.RegisterParser(p)

	return p
}

func TestUnsupportedBinaryMathReturnsNan(t *testing.T) {
	tests := []string{
		`1 / ""`,
		`1 + {a: 1}`,
	}

	for _, input := range tests {
		p := newBinaryTestParser(input)

		exprs, err := p.Parse()
		if err != nil {
			t.Fatal(err)
		}

		got, err := exprs[0].Eval(p.Ctx())
		if err != nil {
			t.Fatal(err)
		}

		if !math.IsNaN(got.UnsafeCastFloat64()) {
			t.Fatalf("expected %q to return nan, got %#v", input, got.Any())
		}
	}
}

func TestPlusWithStringConcats(t *testing.T) {
	p := newBinaryTestParser(`1 + ""`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if got.Any() != "1" {
		t.Fatalf("expected string concat, got %#v", got.Any())
	}
}
