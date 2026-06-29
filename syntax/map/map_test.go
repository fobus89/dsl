package map_parser_test

import (
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	map_parser "github.com/fobus89/dsl/syntax/map"
	member_parser "github.com/fobus89/dsl/syntax/member"
	"github.com/fobus89/dsl/value"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newMapTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	assignment_parser.RegisterParser(p)
	member_parser.RegisterParser(p)
	map_parser.RegisterParser(p)

	return p
}

func TestMapLiteralAssignment(t *testing.T) {
	p := newMapTestParser(`
		somevar = {
			name: "",
			age: r.age,
			sum: 1 + 2,
			nested: {
				ok: true
			}
		}
	`)

	p.Ctx().SetValue("r", value.NewType(map[string]any{
		"age": 30,
	}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	for _, expr := range exprs {
		if _, err := expr.Eval(p.Ctx()); err != nil {
			t.Fatal(err)
		}
	}

	got, ok := p.Ctx().GetValue("somevar")
	if !ok {
		t.Fatal("expected somevar in context")
	}

	m := got.Any().(map[string]any)
	if m["name"] != "" {
		t.Fatalf("expected empty name, got %#v", m)
	}

	if m["age"] != 30 {
		t.Fatalf("expected age from member, got %#v", m)
	}

	if m["sum"] != float64(3) {
		t.Fatalf("expected expression value, got %#v", m)
	}

	nested := m["nested"].(map[string]any)
	if nested["ok"] != true {
		t.Fatalf("expected nested map, got %#v", m)
	}
}
