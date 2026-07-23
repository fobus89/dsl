package forstmt_parser_test

import (
	"strings"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	call_parser "github.com/fobus89/dsl/syntax/call"
	comparison_parser "github.com/fobus89/dsl/syntax/comparison"
	forstmt_parser "github.com/fobus89/dsl/syntax/for_stmt"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newForParser(input string) testParser {
	p := parser.NewParser(input)
	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	comparison_parser.RegisterParser(p)
	assignment_parser.RegisterParser(p)
	call_parser.RegisterParser(p)
	forstmt_parser.RegisterParser(p)
	return p
}

func TestForInEval(t *testing.T) {
	p := newForParser(`
		sum = 0
		for item in items {
			sum = sum + item
		}
	`)
	p.Ctx().SetValue("items", value.NewType([]int{1, 2, 3}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	for _, expr := range exprs {
		if _, err := expr.Eval(p.Ctx()); err != nil {
			t.Fatal(err)
		}
	}

	sum, ok := p.Ctx().GetValue("sum")
	if !ok {
		t.Fatal("sum not found")
	}
	if got := sum.UnsafeCastInt(); got != 6 {
		t.Fatalf("sum = %d, want 6", got)
	}
}

func TestForPrintGO(t *testing.T) {
	p := newForParser(`
		for item in items {
			log(item)
		}
		for ready {
			log("waiting")
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	first, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(first, "for _, item := range items {") {
		t.Fatalf("unexpected for-in code:\n%s", first)
	}

	second, err := exprs[1].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(second, "for ready {") {
		t.Fatalf("unexpected condition for code:\n%s", second)
	}
}
