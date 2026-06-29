package call_parser_test

import (
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	call_parser "github.com/fobus89/dsl/syntax/call"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newCallTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	call_parser.RegisterParser(p)

	return p
}

func TestCallInvokesFunctionWithArgs(t *testing.T) {
	p := newCallTestParser(`sum(1, 2 + 3)`)
	p.Ctx().SetFunc("sum", func(vals ...value.Type) (value.Type, error) {
		var total float64

		for _, val := range vals {
			total += val.UnsafeCastFloat64()
		}

		return value.NewType(total), nil
	})

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if got.UnsafeCastFloat64() != 6 {
		t.Fatalf("expected sum result to be 6, got %#v", got.Any())
	}
}

func TestCallReturnsErrorForMissingFunction(t *testing.T) {
	p := newCallTestParser(`missing()`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := exprs[0].Eval(p.Ctx()); err == nil {
		t.Fatal("expected missing function error")
	}
}
