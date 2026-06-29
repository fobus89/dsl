package any_parser_test

import (
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	any_parser "github.com/fobus89/dsl/syntax/any"
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newAnyTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	any_parser.RegisterParser(p)
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

func TestAnyReturnsMatchedMap(t *testing.T) {
	p := newAnyTestParser(`r = sample any target`)
	p.Ctx().SetValue("sample", value.NewType(map[string]any{
		"id":   int64(1),
		"name": "user",
	}))
	p.Ctx().SetValue("target", value.NewType(map[string]any{
		"id":   float64(1),
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

func TestAnyFindsMapInsideSliceWithJSONNumber(t *testing.T) {
	p := newAnyTestParser(`r = sample any users`)
	p.Ctx().SetValue("sample", value.NewType(map[string]any{
		"id": int64(1),
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

	m := got.Any().(map[string]any)
	if m["name"] != "Leanne Graham" {
		t.Fatalf("expected first user, got %#v", m)
	}
}

func TestAnyPrimitiveReturnsMatchedValue(t *testing.T) {
	p := newAnyTestParser(`r = needle any nums`)
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

func TestAnyReturnsNilWhenNotFound(t *testing.T) {
	p := newAnyTestParser(`r = sample any target`)
	p.Ctx().SetValue("sample", value.NewType(map[string]any{
		"id": 2,
	}))
	p.Ctx().SetValue("target", value.NewType(map[string]any{
		"id": 1,
	}))

	evalProgram(t, p)

	got, ok := p.Ctx().GetValue("r")
	if !ok {
		t.Fatal("expected r in context")
	}

	if !got.IsNil() {
		t.Fatalf("expected nil, got %#v", got.Any())
	}
}
