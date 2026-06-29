package assignment_parser_test

import (
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newAssignmentTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	assignment_parser.RegisterParser(p)

	return p
}

func TestAssignmentSetsValue(t *testing.T) {
	p := newAssignmentTestParser(`answer = 1 + 2`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	got, ok := p.Ctx().GetValue("answer")
	if !ok {
		t.Fatal("expected answer in context")
	}

	if got.UnsafeCastFloat64() != 3 {
		t.Fatalf("expected answer to be 3, got %#v", got.Any())
	}
}

func TestAssignmentRequiresIdentLeftSide(t *testing.T) {
	p := newAssignmentTestParser(`1 = 2`)

	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error")
	}
}
