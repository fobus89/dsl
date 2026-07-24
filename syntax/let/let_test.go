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
	if !strings.Contains(generated, "result := func() string") ||
		!strings.Contains(generated, `return "yes"`) {
		t.Fatalf("unexpected generated if expression:\n%s", generated)
	}
}

func TestLetIfRejectsIncompatibleBranchTypes(t *testing.T) {
	p := newLetParser(`
		let result = if true: "yes" else: 1
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	_, err = exprs[0].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(
			err.Error(),
			"if branches return incompatible types string and int",
		) {
		t.Fatalf("expected incompatible branch error, got %v", err)
	}
}

func TestLetIfWithoutElseUsesTypedZeroValue(t *testing.T) {
	p := newLetParser(`
		let x = if false { mixed }
	`)
	p.Ctx().SetValue(
		"mixed",
		value.NewTypeWithExplicit(
			[2][][]any{},
			"[2][][]User",
		),
	)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	x, _ := p.Ctx().GetValue("x")
	if got := x.TypeName(); got != "[2][][]User" {
		t.Fatalf("x type = %q, want [2][][]User", got)
	}

	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"x := func() [2][][]User",
		"return mixed",
		"var zero [2][][]User",
		"return zero",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf(
				"typed if IIFE misses %q:\n%s",
				expected,
				generated,
			)
		}
	}
	if strings.Contains(generated, "func() any") ||
		strings.Contains(generated, "return nil") {
		t.Fatalf("typed if fell back to any/nil:\n%s", generated)
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
		"result := func() []int",
		"result := make([]int, 0)",
		"continue",
		"break",
		"result = append(result, item)",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated code misses %q:\n%s", expected, generated)
		}
	}
}

func TestForYieldPreservesNestedElementType(t *testing.T) {
	p := newLetParser(`
		let result = for item in mixed {
			yield item
		}
	`)
	p.Ctx().SetValue(
		"mixed",
		value.NewTypeWithExplicit(
			[2][][]any{},
			"[2][][]User",
		),
	)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(
		generated,
		"result := make([][][]User, 0)",
	) {
		t.Fatalf("unexpected generated accumulator:\n%s", generated)
	}
	if !strings.Contains(generated, "result := func() [][][]User") ||
		!strings.Contains(generated, "return result") {
		t.Fatalf("yield result is not isolated in IIFE:\n%s", generated)
	}
}

func TestConditionalYield(t *testing.T) {
	p := newLetParser(`
		let result = for item in items {
			yield if item > 1 { item }
		}
	`)
	p.Ctx().SetValue("items", value.NewType([]int{1, 2, 3}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	result, _ := p.Ctx().GetValue("result")
	if got, want := result.Any(), []any{2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}

	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"result := func() []int",
		"if (item > 1)",
		"result = append(result, item)",
		"return result",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf(
				"conditional yield misses %q:\n%s",
				expected,
				generated,
			)
		}
	}
}

func TestConditionalYieldElse(t *testing.T) {
	p := newLetParser(`
		let result = for item in items {
			yield if item > 1 { "large" } else { "small" }
		}
	`)
	p.Ctx().SetValue("items", value.NewType([]int{1, 2}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	result, _ := p.Ctx().GetValue("result")
	if got, want := result.Any(), []any{"small", "large"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
}

func TestForYieldRejectsIncompatibleTypes(t *testing.T) {
	p := newLetParser(`
		let result = for item in items {
			if true {
				yield item
			}
			yield "wrong"
		}
	`)
	p.Ctx().SetValue("items", value.NewType([]int{1, 2}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	_, err = exprs[0].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(err.Error(), "incompatible types int and string") {
		t.Fatalf("expected incompatible yield error, got %v", err)
	}
}

func TestForBreakReturnsScalarValue(t *testing.T) {
	p := newLetParser(`
		let result = for item in items {
			if item == 3 {
				break item
			}
		}
	`)
	p.Ctx().SetValue("items", value.NewType([]int{1, 2, 3, 4}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	result, _ := p.Ctx().GetValue("result")
	if got := result.UnsafeCastInt(); got != 3 {
		t.Fatalf("result = %d, want 3", got)
	}
	if got := result.TypeName(); got != "int" {
		t.Fatalf("result type = %q, want int", got)
	}

	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"result := func() int",
		"return item",
		"var zero int",
		"return zero",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf(
				"generated scalar loop misses %q:\n%s",
				expected,
				generated,
			)
		}
	}
	if strings.Contains(generated, "make([]") ||
		strings.Contains(generated, "append(") ||
		strings.Contains(generated, "var result") {
		t.Fatalf("scalar break generated an accumulator:\n%s", generated)
	}
}

func TestForBreakScalarUsesZeroValueWhenNotFound(t *testing.T) {
	p := newLetParser(`
		let result = for item in items {
			if item == 9 {
				break item
			}
		}
	`)
	p.Ctx().SetValue("items", value.NewType([]int{1, 2, 3}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	result, _ := p.Ctx().GetValue("result")
	if got := result.UnsafeCastInt(); got != 0 {
		t.Fatalf("result = %d, want zero value", got)
	}
}

func TestForBreakNestedTypeUsesIIFE(t *testing.T) {
	p := newLetParser(`
		let result = for item in mixed {
			break item
		}
	`)
	p.Ctx().SetValue(
		"mixed",
		value.NewTypeWithExplicit(
			[2][][]any{},
			"[2][][]User",
		),
	)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"result := func() [][]User",
		"return item",
		"var zero [][]User",
		"return zero",
		"}()",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf(
				"generated IIFE misses %q:\n%s",
				expected,
				generated,
			)
		}
	}
}

func TestConditionalBreakReturnsScalar(t *testing.T) {
	p := newLetParser(`
		let result = for item in items {
			break if item == 3 { item }
		}
	`)
	p.Ctx().SetValue("items", value.NewType([]int{1, 2, 3, 4}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	result, _ := p.Ctx().GetValue("result")
	if got := result.UnsafeCastInt(); got != 3 {
		t.Fatalf("result = %d, want 3", got)
	}

	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"result := func() int",
		"if (item == 3)",
		"return item",
		"var zero int",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf(
				"conditional break misses %q:\n%s",
				expected,
				generated,
			)
		}
	}
}

func TestConditionalBreakPreservesNestedType(t *testing.T) {
	p := newLetParser(`
		let result = for item in mixed {
			break if true { item }
		}
	`)
	p.Ctx().SetValue(
		"mixed",
		value.NewTypeWithExplicit(
			[2][][]any{},
			"[2][][]User",
		),
	)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(generated, "result := func() [][]User") ||
		!strings.Contains(generated, "return item") {
		t.Fatalf("unexpected conditional break IIFE:\n%s", generated)
	}
}

func TestForRejectsYieldAndBreakValueTogether(t *testing.T) {
	p := newLetParser(`
		let result = for item in items {
			if item == 2 {
				break item
			}
			yield item
		}
	`)
	p.Ctx().SetValue("items", value.NewType([]int{1, 2, 3}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	_, err = exprs[0].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(err.Error(), "cannot mix yield with break value") {
		t.Fatalf("expected mixed result mode error, got %v", err)
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
