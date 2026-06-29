package member_parser_test

import (
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	map_parser "github.com/fobus89/dsl/syntax/map"
	member_parser "github.com/fobus89/dsl/syntax/member"
	"github.com/fobus89/dsl/value"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newMemberTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	map_parser.RegisterParser(p)
	member_parser.RegisterParser(p)

	return p
}

func TestMissingDeepMemberReturnsNil(t *testing.T) {
	p := newMemberTestParser(`somevar.a.ida`)
	p.Ctx().SetValue("somevar", value.NewType(map[string]any{
		"a": map[string]any{
			"id": 1,
		},
	}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if !got.IsNil() {
		t.Fatalf("expected nil, got %#v", got.Any())
	}
}

func TestMemberOnNonObjectReturnsNil(t *testing.T) {
	p := newMemberTestParser(`somevar.a.ida`)
	p.Ctx().SetValue("somevar", value.NewType(map[string]any{
		"a": 1,
	}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if !got.IsNil() {
		t.Fatalf("expected nil, got %#v", got.Any())
	}
}
