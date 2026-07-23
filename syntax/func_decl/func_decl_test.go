package funcdecl_parser_test

import (
	"strings"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	call_parser "github.com/fobus89/dsl/syntax/call"
	funcdecl_parser "github.com/fobus89/dsl/syntax/func_decl"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	map_parser "github.com/fobus89/dsl/syntax/map"
	member_parser "github.com/fobus89/dsl/syntax/member"
	typedecl_parser "github.com/fobus89/dsl/syntax/type_decl"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newFuncDeclTestParser(input string) testParser {
	p := parser.NewParser(input)
	literal_parser.RegisterParser(p)
	assignment_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	call_parser.RegisterParser(p)
	map_parser.RegisterParser(p)
	member_parser.RegisterParser(p)
	typedecl_parser.RegisterParser(p)
	funcdecl_parser.RegisterParser(p)
	return p
}

func TestParseFunctionDeclaration(t *testing.T) {
	p := newFuncDeclTestParser(`fn add(a: Int, b: Int) Int { return a + b }`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	decl, ok := exprs[0].(*funcdecl_parser.FuncDecl)
	if !ok {
		t.Fatalf("expected *FuncDecl, got %T", exprs[0])
	}
	if decl.IsMethod() {
		t.Fatal("ordinary function must not be a method")
	}
	if decl.Name != "add" {
		t.Fatalf("expected add, got %s", decl.Name)
	}
	if len(decl.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(decl.Params))
	}
	if decl.ReturnType == nil {
		t.Fatal("expected return type")
	}
	if len(decl.Body) != 1 {
		t.Fatalf("expected one body statement, got %d", len(decl.Body))
	}
}

func TestParseMethodDeclaration(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		fn (u: User) name() String { return u.name }
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	decl, ok := exprs[1].(*funcdecl_parser.FuncDecl)
	if !ok {
		t.Fatalf("expected *FuncDecl, got %T", exprs[1])
	}
	if !decl.IsMethod() {
		t.Fatal("declaration with receiver must be a method")
	}
	if decl.Recv.Name != "u" {
		t.Fatalf("expected receiver u, got %s", decl.Recv.Name)
	}
	if decl.Recv.Type.IsPtr {
		t.Fatal("expected value receiver")
	}
	if decl.Name != "name" {
		t.Fatalf("expected name, got %s", decl.Name)
	}
}

func TestParseAndPrintPointerMethodReceiver(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		fn (u: *User) Name() String { return u.name }
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	decl, ok := exprs[1].(*funcdecl_parser.FuncDecl)
	if !ok {
		t.Fatalf("expected *FuncDecl, got %T", exprs[1])
	}
	if !decl.Recv.Type.IsPtr {
		t.Fatal("expected pointer receiver")
	}

	got, err := decl.PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	want := "func (u *User) Name() string {\n\treturn u.name\n}"
	if got != want {
		t.Fatalf("PrintGO() = %q, want %q", got, want)
	}
}

func TestPointerParameterAndReturnType(t *testing.T) {
	p := newFuncDeclTestParser(`
		fn copy(value: *String) *String {
			return value
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	decl := exprs[0].(*funcdecl_parser.FuncDecl)
	if decl.Params[0].Type.Name != "String" ||
		!decl.Params[0].Type.IsPtr {
		t.Fatalf("unexpected parameter type: %#v", decl.Params[0].Type)
	}
	if decl.ReturnType.Name != "String" || !decl.ReturnType.IsPtr {
		t.Fatalf("unexpected return type: %#v", decl.ReturnType)
	}

	got, err := decl.PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	want := "func copy(value *string) *string {\n\treturn value\n}"
	if got != want {
		t.Fatalf("PrintGO() = %q, want %q", got, want)
	}
}

func TestFunctionIsAvailableBeforeItsDeclaration(t *testing.T) {
	p := newFuncDeclTestParser(`
		add(2, 3)
		fn add(a: Int, b: Int) Int { return a + b }
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	result, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if got := result.UnsafeCastInt(); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestCallMethodThroughReceiver(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		user = User {name: "Bob"}
		user.name()
		fn (u: User) name() String { return u.name }
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := exprs[1].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	result, err := exprs[2].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if got := result.UnsafeCastString(); got != "Bob" {
		t.Fatalf("expected Bob, got %q", got)
	}
}

func TestMethodRejectsReceiverWithoutDeclaredType(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		plain = {name: "Bob"}
		plain.name()
		fn (u: User) name() String { return u.name }
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[1].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	_, err = exprs[2].Eval(p.Ctx())
	if err == nil {
		t.Fatal("expected method receiver type error")
	}
	if !strings.Contains(err.Error(), "map[string]any") {
		t.Fatalf("expected actual receiver type in error, got %v", err)
	}
}

func TestMethodRejectsDifferentDeclaredReceiverType(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		type Admin struct { name: string }
		admin = Admin {name: "Root"}
		admin.name()
		fn (u: User) name() String { return u.name }
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[2].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	_, err = exprs[3].Eval(p.Ctx())
	if err == nil {
		t.Fatal("expected method receiver type error")
	}
	if !strings.Contains(err.Error(), "type Admin") {
		t.Fatalf("expected Admin receiver type in error, got %v", err)
	}
}

func TestMethodRequiresDeclaredReceiverType(t *testing.T) {
	p := newFuncDeclTestParser(
		`fn (u: Missing) name() String { return u.name }`,
	)

	if _, err := p.Parse(); err == nil {
		t.Fatal("expected undeclared receiver type error")
	}
}

func TestMethodReceiverTypeCanBeDeclaredLater(t *testing.T) {
	p := newFuncDeclTestParser(`
		fn (u: User) name() String { return u.name }
		type User struct { name: string }
	`)

	if _, err := p.Parse(); err != nil {
		t.Fatal(err)
	}
}
