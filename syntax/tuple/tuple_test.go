package tuple_parser_test

import (
	"strings"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	call_parser "github.com/fobus89/dsl/syntax/call"
	funcdecl_parser "github.com/fobus89/dsl/syntax/func_decl"
	let_parser "github.com/fobus89/dsl/syntax/let"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	member_parser "github.com/fobus89/dsl/syntax/member"
	typedecl_parser "github.com/fobus89/dsl/syntax/type_decl"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newTupleParser(input string) testParser {
	p := parser.NewParser(input)
	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	call_parser.RegisterParser(p)
	funcdecl_parser.RegisterParser(p)
	member_parser.RegisterParser(p)
	typedecl_parser.RegisterParser(p)
	let_parser.RegisterParser(p)
	return p
}

func TestTupleAliasReceiverMethod(t *testing.T) {
	p := newTupleParser(`
		type Point = (int, int)

		fn (p: Point) sum() int {
			return p.0 + p.1
		}

		let point = Point((10, 20))
		let result = point.sum()
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
	result, ok := p.Ctx().GetValue("result")
	if !ok || result.UnsafeCastInt() != 30 {
		t.Fatalf("result = %#v, want 30", result.Any())
	}

	method, err := exprs[1].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(
		method,
		"func (p Point) sum() int",
	) || !strings.Contains(method, "p.V0") ||
		!strings.Contains(method, "p.V1") {
		t.Fatalf("unexpected Point method:\n%s", method)
	}

	constructor, err := exprs[2].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(constructor, "Point(struct { V0 int; V1 int }") {
		t.Fatalf("unexpected Point construction:\n%s", constructor)
	}
}

func TestTupleTypeDeclaration(t *testing.T) {
	p := newTupleParser(`type Point = (int, int)`)
	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if generated !=
		"type Point struct { V0 int; V1 int }" {
		t.Fatalf("unexpected tuple type:\n%s", generated)
	}
}

func TestTupleAccessAndDestructuring(t *testing.T) {
	p := newTupleParser(`
		let pair = (10, "hello")
		let (number, text) = pair
		pair.0
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
	number, _ := p.Ctx().GetValue("number")
	text, _ := p.Ctx().GetValue("text")
	if number.Any() != int64(10) ||
		text.UnsafeCastString() != "hello" {
		t.Fatalf(
			"destructured values = %#v, %#v",
			number.Any(),
			text.Any(),
		)
	}

	first, err := exprs[2].Eval(p.Ctx())
	if err != nil || first.Any() != int64(10) {
		t.Fatalf("pair.0 = %#v, err=%v", first.Any(), err)
	}

	pairGO, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	destructureGO, err := exprs[1].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(pairGO, "struct { V0 int; V1 string }") ||
		!strings.Contains(destructureGO, "__tuple.V0") ||
		!strings.Contains(destructureGO, "__tuple.V1") {
		t.Fatalf(
			"unexpected tuple Go:\n%s\n%s",
			pairGO,
			destructureGO,
		)
	}
}

func TestTupleDestructuringArityIsChecked(t *testing.T) {
	p := newTupleParser(`let (a, b, c) = (1, 2)`)
	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].PrintGO(p.Ctx()); err == nil ||
		!strings.Contains(err.Error(), "expects 3") {
		t.Fatalf("expected tuple arity error, got %v", err)
	}
}
