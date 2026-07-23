package let_parser_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	comparison_parser "github.com/fobus89/dsl/syntax/comparison"
	flow_parser "github.com/fobus89/dsl/syntax/flow"
	forstmt_parser "github.com/fobus89/dsl/syntax/for_stmt"
	ifstmt_parser "github.com/fobus89/dsl/syntax/if_stmt"
	let_parser "github.com/fobus89/dsl/syntax/let"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	logical_parser "github.com/fobus89/dsl/syntax/logical"
	"github.com/fobus89/dsl/value"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newLetParser(input string) testParser {
	p := parser.NewParser(input)
	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	comparison_parser.RegisterParser(p)
	logical_parser.RegisterParser(p)
	flow_parser.RegisterParser(p)
	ifstmt_parser.RegisterParser(p)
	forstmt_parser.RegisterParser(p)
	let_parser.RegisterParser(p)
	return p
}

func TestLetIfExpression(t *testing.T) {
	p := newLetParser(`
		let result = if true: "yes" else: "no"
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	result, ok := p.Ctx().GetValue("result")
	if !ok || result.UnsafeCastString() != "yes" {
		t.Fatalf("result = %#v, want yes", result.Any())
	}

	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(generated, "result := func() any") ||
		!strings.Contains(generated, `return "yes"`) {
		t.Fatalf("unexpected generated if expression:\n%s", generated)
	}
}

func TestLetForYieldBreakContinue(t *testing.T) {
	p := newLetParser(`
		let result = for item in items {
			if item == 2 {
				continue
			}
			if item == 4 {
				break
			}
			yield item
		}
	`)
	p.Ctx().SetValue("items", value.NewType([]int{1, 2, 3, 4, 5}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	result, ok := p.Ctx().GetValue("result")
	if !ok {
		t.Fatal("result not found")
	}
	if got, want := result.Any(), []any{1, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}

	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"result := make([]any, 0)",
		"continue",
		"break",
		"result = append(result, item)",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated code misses %q:\n%s", expected, generated)
		}
	}
}

func TestSingleExpressionFor(t *testing.T) {
	p := newLetParser(`
		let result = for item in items: yield item
	`)
	p.Ctx().SetValue("items", value.NewType([]string{"a", "b"}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	result, _ := p.Ctx().GetValue("result")
	if got, want := result.Any(), []any{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
}
