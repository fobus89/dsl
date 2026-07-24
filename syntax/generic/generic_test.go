package generic_parser_test

import (
	astgo "go/ast"
	goparser "go/parser"
	gotoken "go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	call_parser "github.com/fobus89/dsl/syntax/call"
	collection_parser "github.com/fobus89/dsl/syntax/collection"
	funcdecl_parser "github.com/fobus89/dsl/syntax/func_decl"
	generic_parser "github.com/fobus89/dsl/syntax/generic"
	let_parser "github.com/fobus89/dsl/syntax/let"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	member_parser "github.com/fobus89/dsl/syntax/member"
	typedecl_parser "github.com/fobus89/dsl/syntax/type_decl"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newGenericParser(input string) testParser {
	p := parser.NewParser(input)
	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	call_parser.RegisterParser(p)
	collection_parser.RegisterParser(p)
	generic_parser.RegisterParser(p)
	member_parser.RegisterParser(p)
	typedecl_parser.RegisterParser(p)
	funcdecl_parser.RegisterParser(p)
	let_parser.RegisterParser(p)
	return p
}

func TestGenericTypeFunctionAndMethods(t *testing.T) {
	p := newGenericParser(`
		type Box[T any] struct {
			value: T
		}

		fn identity[T any](value: T) T {
			return value
		}

		fn (b: Box[T]) get() T {
			return b.value
		}

		fn (b: Box[T]) replace[R any](value: R) R {
			return value
		}

		let box = Box[int]{value: 10}
		let textBox = Box[string]{value: "text"}
		let first = identity[int](box.get())
		let text = identity[string](textBox.get())
		let second = box.replace[string]("ok")
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
	first, _ := p.Ctx().GetValue("first")
	second, _ := p.Ctx().GetValue("second")
	text, _ := p.Ctx().GetValue("text")
	if first.UnsafeCastInt() != 10 ||
		second.UnsafeCastString() != "ok" ||
		text.UnsafeCastString() != "text" {
		t.Fatalf(
			"generic results = %#v, %#v",
			first.Any(),
			second.Any(),
		)
	}

	parts := make([]string, 0, len(exprs))
	for _, expr := range exprs {
		printed, err := expr.PrintGO(p.Ctx())
		if err != nil {
			t.Fatal(err)
		}
		parts = append(parts, printed)
	}
	joined := strings.Join(parts, "\n\n")
	for _, expected := range []string{
		"type Box_int struct",
		"type Box_string struct",
		"func identity_int(value int) int",
		"func identity_string(value string) string",
		"func (b Box_int) get() int",
		"func (b Box_string) get() string",
		"func Box_replace_int_string(b Box_int, value string) string",
		"box := Box_int{value: 10}",
		`textBox := Box_string{value: "text"}`,
		"first := identity_int(box.get())",
		"text := identity_string(textBox.get())",
		`second := Box_replace_int_string(box, "ok")`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf(
				"generated generics miss %q:\n%s",
				expected,
				joined,
			)
		}
	}
	if strings.Contains(joined, "[T ") ||
		strings.Contains(joined, "[R ") {
		t.Fatalf("generated Go still contains generics:\n%s", joined)
	}

	source := "package generated\n\n" +
		strings.Join(parts[:4], "\n\n") +
		"\n\nfunc run() {\n" +
		strings.Join(parts[4:], "\n") +
		"\n_ = first\n_ = second\n_ = text\n}"
	files := gotoken.NewFileSet()
	file, err := goparser.ParseFile(
		files,
		"generated.go",
		source,
		goparser.AllErrors,
	)
	if err != nil {
		t.Fatalf("generated Go does not parse: %v\n%s", err, source)
	}
	if _, err := (&types.Config{}).Check(
		"generated",
		files,
		[]*astgo.File{file},
		nil,
	); err != nil {
		t.Fatalf(
			"generated Go does not type-check: %v\n%s",
			err,
			source,
		)
	}
}

func TestGenericFunctionGeneratesOnlyConcreteGo(t *testing.T) {
	p := newGenericParser(`
		fn identity[T any](value: T) T { return value }
		let value = identity[int](1)
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	var parts []string
	for _, expr := range exprs {
		printed, err := expr.PrintGO(p.Ctx())
		if err != nil {
			t.Fatal(err)
		}
		parts = append(parts, printed)
	}
	generated := strings.Join(parts, "\n")

	if !strings.Contains(
		generated,
		"func identity_int(value int) int",
	) {
		t.Fatalf("concrete function was not generated:\n%s", generated)
	}
	if !strings.Contains(generated, "value := identity_int(1)") {
		t.Fatalf("concrete call was not generated:\n%s", generated)
	}
	if strings.Contains(generated, "identity[") ||
		strings.Contains(generated, "[T any]") {
		t.Fatalf("Go generics leaked into generated code:\n%s", generated)
	}
}

func TestGenericArityAndConstraintsAreChecked(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error string
	}{
		{
			name: "function arity",
			input: `
				fn identity[T any](value: T) T { return value }
				let value = identity[int, string](1)
			`,
			error: "expects 1 type arguments, got 2",
		},
		{
			name: "comparable constraint",
			input: `
				fn same[T comparable](value: T) T { return value }
				let value = same[[]int]([]int{1})
			`,
			error: "does not satisfy comparable",
		},
		{
			name: "generic struct field",
			input: `
				type Box[T any] struct { value: T }
				let box = Box[int]{value: "wrong"}
			`,
			error: "expects int, got string",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := newGenericParser(test.input)
			exprs, err := p.Parse()
			if err != nil {
				t.Fatal(err)
			}
			for _, expr := range exprs {
				if _, err = expr.PrintGO(p.Ctx()); err != nil {
					break
				}
			}
			if err == nil ||
				!strings.Contains(err.Error(), test.error) {
				t.Fatalf(
					"expected %q, got %v",
					test.error,
					err,
				)
			}
		})
	}
}
