package comparison_parser_test

import (
	"strings"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	comparison_parser "github.com/fobus89/dsl/syntax/comparison"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func TestComparisonRejectsSliceAndNumber(t *testing.T) {
	p := newComparisonTestParser(`item == 2`)
	p.Ctx().SetValue(
		"item",
		value.NewTypeWithExplicit(nil, "[][]User"),
	)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	_, err = exprs[0].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(
			err.Error(),
			"cannot compare [][]User with int",
		) {
		t.Fatalf("expected incompatible comparison error, got %v", err)
	}
}

func TestComparisonRejectsSlicesOfSameType(t *testing.T) {
	p := newComparisonTestParser(`a == b`)
	p.Ctx().SetValue(
		"a",
		value.NewTypeWithExplicit(nil, "[]int"),
	)
	p.Ctx().SetValue(
		"b",
		value.NewTypeWithExplicit(nil, "[]int"),
	)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	_, err = exprs[0].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(err.Error(), "requires comparable types") {
		t.Fatalf("expected non-comparable slice error, got %v", err)
	}
}

func TestComparisonAllowsComparableArrays(t *testing.T) {
	p := newComparisonTestParser(`a == b`)
	p.Ctx().SetValue(
		"a",
		value.NewTypeWithExplicit(nil, "[2]int"),
	)
	p.Ctx().SetValue(
		"b",
		value.NewTypeWithExplicit(nil, "[2]int"),
	)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].PrintGO(p.Ctx()); err != nil {
		t.Fatal(err)
	}
}

func TestOrderedComparisonRejectsBool(t *testing.T) {
	p := newComparisonTestParser(`true > false`)
	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	_, err = exprs[0].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(err.Error(), "requires numbers or strings") {
		t.Fatalf("expected ordered type error, got %v", err)
	}
}

func newComparisonTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	comparison_parser.RegisterParser(p)

	return p
}

func TestComparison(t *testing.T) {
	p := newComparisonTestParser(`1 < 2`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if !got.UnsafeCastBool() {
		t.Fatalf("expected true")
	}
}

func TestComparisonEquality(t *testing.T) {
	p := newComparisonTestParser(`1 == 1.0`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if !got.UnsafeCastBool() {
		t.Fatalf("expected true")
	}
}
