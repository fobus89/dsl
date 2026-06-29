package all_parser_test

import (
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	all_parser "github.com/fobus89/dsl/syntax/all"
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newAllTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	all_parser.RegisterParser(p)
	assignment_parser.RegisterParser(p)

	return p
}

func evalProgram(t *testing.T, p testParser) {
	t.Helper()

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	for _, expr := range exprs {
		if _, err := expr.Eval(p.Ctx()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAllReturnsMatchedMap(t *testing.T) {
	p := newAllTestParser(`r = sample all target`)
	p.Ctx().SetValue("sample", value.NewType(map[string]any{
		"id":   1,
		"name": "user",
	}))
	p.Ctx().SetValue("target", value.NewType(map[string]any{
		"id":   1,
		"name": "user",
		"age":  20,
	}))

	evalProgram(t, p)

	got, ok := p.Ctx().GetValue("r")
	if !ok {
		t.Fatal("expected r in context")
	}

	m := got.Any().(map[string]any)
	if m["age"] != 20 {
		t.Fatalf("expected right map, got %#v", m)
	}
}

func TestAllReturnsNilWhenLeftHasExtraKeys(t *testing.T) {
	p := newAllTestParser(`r = sample all users`)
	p.Ctx().SetValue("sample", value.NewType(map[string]any{
		"id": int64(1),
		"a":  int64(2),
		"b":  int64(3),
		"c":  int64(4),
	}))
	p.Ctx().SetValue("users", value.NewType([]map[string]any{
		{
			"id":   float64(1),
			"name": "Leanne Graham",
		},
		{
			"id":   float64(2),
			"name": "Ervin Howell",
		},
	}))

	evalProgram(t, p)

	got, ok := p.Ctx().GetValue("r")
	if !ok {
		t.Fatal("expected r in context")
	}

	if !got.IsNil() {
		t.Fatalf("expected nil because all keys must match, got %#v", got.Any())
	}
}

func TestAllPrimitiveReturnsMatchedValue(t *testing.T) {
	p := newAllTestParser(`r = needle all nums`)
	p.Ctx().SetValue("needle", value.NewType(2))
	p.Ctx().SetValue("nums", value.NewType([]int{1, 2, 3}))

	evalProgram(t, p)

	got, ok := p.Ctx().GetValue("r")
	if !ok {
		t.Fatal("expected r in context")
	}

	if got.UnsafeCastInt() != 2 {
		t.Fatalf("expected 2, got %#v", got.Any())
	}
}
