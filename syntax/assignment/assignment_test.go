package assignment_parser_test

import (
	"strings"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	let_parser "github.com/fobus89/dsl/syntax/let"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
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
	let_parser.RegisterParser(p)

	return p
}

func TestAssignmentSetsValue(t *testing.T) {
	p := newAssignmentTestParser(`answer = 1 + 2`)
	p.Ctx().SetValue("answer", value.NewType(0))

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

	printed, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(printed, ":=") || printed != "answer = (1 + 2)" {
		t.Fatalf("unexpected assignment Go: %s", printed)
	}
}

func TestAssignmentRequiresIdentLeftSide(t *testing.T) {
	p := newAssignmentTestParser(`1 = 2`)

	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestAssignmentRequiresDeclaredValue(t *testing.T) {
	p := newAssignmentTestParser(`missing = 1`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err == nil ||
		!strings.Contains(err.Error(), "undeclared value missing") {
		t.Fatalf("expected undeclared value error, got %v", err)
	}
	if _, err := exprs[0].PrintGO(p.Ctx()); err == nil ||
		!strings.Contains(err.Error(), "undeclared value missing") {
		t.Fatalf("expected PrintGO undeclared value error, got %v", err)
	}
}

func TestAssignmentBeforeDeclarationIsRejected(t *testing.T) {
	p := newAssignmentTestParser(`
		late = 1
		let late = 2
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err == nil ||
		!strings.Contains(err.Error(), "undeclared value late") {
		t.Fatalf("expected forward assignment error, got %v", err)
	}
}
