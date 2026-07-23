package ifstmt_parser_test

import (
	"strings"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	call_parser "github.com/fobus89/dsl/syntax/call"
	comparison_parser "github.com/fobus89/dsl/syntax/comparison"
	funcdecl_parser "github.com/fobus89/dsl/syntax/func_decl"
	ifstmt_parser "github.com/fobus89/dsl/syntax/if_stmt"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newIfParser(input string) testParser {
	p := parser.NewParser(input)
	literal_parser.RegisterParser(p)
	comparison_parser.RegisterParser(p)
	call_parser.RegisterParser(p)
	ifstmt_parser.RegisterParser(p)
	funcdecl_parser.RegisterParser(p)
	return p
}

func TestIfElseEvalAndPrintGO(t *testing.T) {
	p := newIfParser(`
		fn choose(ok: bool) string {
			if ok {
				return "yes"
			} else {
				return "no"
			}
		}
		choose(true)
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	result, err := exprs[1].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if got := result.UnsafeCastString(); got != "yes" {
		t.Fatalf("choose(true) = %q, want yes", got)
	}

	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"if ok {", "} else {", `return "yes"`} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated code misses %q:\n%s", expected, generated)
		}
	}
}

func TestElseIfParses(t *testing.T) {
	p := newIfParser(`
		if score > 90 {
			"excellent"
		} else if score > 70 {
			"good"
		} else {
			"retry"
		}
	`)

	p.Ctx().SetValue("score", value.NewType(80))
	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}
}

func TestNestedReturnTypeIsValidated(t *testing.T) {
	p := newIfParser(`
		fn bad(ok: bool) string {
			if ok {
				return 123
			}
			return "ok"
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].PrintGO(p.Ctx()); err == nil {
		t.Fatal("expected nested return type error")
	}
}
