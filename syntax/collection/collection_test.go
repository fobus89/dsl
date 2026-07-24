package collection_parser_test

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"go/types"
	"reflect"
	"strings"
	"testing"

	dslast "github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	collection_parser "github.com/fobus89/dsl/syntax/collection"
	let_parser "github.com/fobus89/dsl/syntax/let"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
)

type testParser interface {
	parser.Parser
	Parse() ([]dslast.Expr, error)
}

func newCollectionParser(input string) testParser {
	p := parser.NewParser(input)
	literal_parser.RegisterParser(p)
	collection_parser.RegisterParser(p)
	let_parser.RegisterParser(p)
	return p
}

func TestArrayAndSliceLiterals(t *testing.T) {
	p := newCollectionParser(`
		let array = [3]int{1, 2, 3}
		let slice = []int{1, 2, 3}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	for _, expr := range exprs {
		if _, err := expr.Eval(p.Ctx()); err != nil {
			t.Fatal(err)
		}
	}

	array, _ := p.Ctx().GetValue("array")
	if got, want := array.TypeName(), "[3]int"; got != want {
		t.Fatalf("array type = %q, want %q", got, want)
	}
	if got, want := array.Any(), []any{int64(1), int64(2), int64(3)}; !reflect.DeepEqual(got, want) {
		t.Fatalf("array = %#v, want %#v", got, want)
	}

	slice, _ := p.Ctx().GetValue("slice")
	if got, want := slice.TypeName(), "[]int"; got != want {
		t.Fatalf("slice type = %q, want %q", got, want)
	}
}

func TestArbitraryDimensionAndMixedCollections(t *testing.T) {
	p := newCollectionParser(`
		let matrix = [1][][][]int{{{{1, 2}, {3}}}}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	matrix, _ := p.Ctx().GetValue("matrix")
	if got, want := matrix.TypeName(), "[1][][][]int"; got != want {
		t.Fatalf("matrix type = %q, want %q", got, want)
	}

	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(generated, "matrix := [1][][][]int{") {
		t.Fatalf("unexpected generated code:\n%s", generated)
	}
	typeCheckGenerated(t, generated)
}

func TestArrayRejectsTooManyValues(t *testing.T) {
	p := newCollectionParser(`let items = [2]int{1, 2, 3}`)
	_, err := p.Parse()
	if err == nil || !strings.Contains(err.Error(), "3 values") {
		t.Fatalf("expected array length error, got %v", err)
	}
}

func TestArrayMissingValuesUseZeroValue(t *testing.T) {
	p := newCollectionParser(`let items = [3]int{7}`)
	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	items, _ := p.Ctx().GetValue("items")
	if got, want := items.Any(), []any{int64(7), int64(0), int64(0)}; !reflect.DeepEqual(got, want) {
		t.Fatalf("items = %#v, want %#v", got, want)
	}
}

func TestCollectionRejectsWrongElementType(t *testing.T) {
	p := newCollectionParser(`let items = []int{1, "wrong"}`)
	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	_, err = exprs[0].Eval(p.Ctx())
	if err == nil || !strings.Contains(err.Error(), "expected int") {
		t.Fatalf("expected element type error, got %v", err)
	}

	_, err = exprs[0].PrintGO(p.Ctx())
	if err == nil || !strings.Contains(err.Error(), "expected int") {
		t.Fatalf("expected PrintGO element type error, got %v", err)
	}
}

func typeCheckGenerated(t *testing.T, generated string) {
	t.Helper()

	source := "package generated\nfunc test() {\n" +
		generated + "\n_ = matrix\n}"
	files := token.NewFileSet()
	file, err := goparser.ParseFile(files, "generated.go", source, 0)
	if err != nil {
		t.Fatalf("generated invalid Go syntax: %v\n%s", err, source)
	}
	config := types.Config{}
	if _, err := config.Check(
		"generated",
		files,
		[]*ast.File{file},
		nil,
	); err != nil {
		t.Fatalf("generated Go does not type-check: %v\n%s", err, source)
	}
}
