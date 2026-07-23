package typedecl_parser_test

import (
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
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

func newTypeDeclTestParser(input string) testParser {
	p := parser.NewParser(input)
	literal_parser.RegisterParser(p)
	assignment_parser.RegisterParser(p)
	call_parser.RegisterParser(p)
	map_parser.RegisterParser(p)
	member_parser.RegisterParser(p)
	typedecl_parser.RegisterParser(p)
	funcdecl_parser.RegisterParser(p)
	return p
}

func TestPointerStructFieldPrintGO(t *testing.T) {
	p := newTypeDeclTestParser(`
		type User struct {
			name: *String
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	decl, ok := exprs[0].(*typedecl_parser.TypeDecl)
	if !ok {
		t.Fatalf("expected *TypeDecl, got %T", exprs[0])
	}
	if got := decl.Def.Fields["name"]; got != (ast.TypeRef{
		Name:  "String",
		IsPtr: true,
	}) {
		t.Fatalf("field type = %#v, want pointer to String", got)
	}

	got, err := decl.PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	want := "type User struct {\n\tname *string\n}"
	if got != want {
		t.Fatalf("PrintGO() = %q, want %q", got, want)
	}
}

func TestAliasConstructorPreservesDeclaredType(t *testing.T) {
	p := newTypeDeclTestParser(`
		type UserID int
		id = UserID(42)
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[1].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	id, ok := p.Ctx().GetValue("id")
	if !ok {
		t.Fatal("expected id in context")
	}
	if id.TypeName() != "UserID" {
		t.Fatalf("expected UserID, got %s", id.TypeName())
	}
}

func TestStructLiteralPreservesDeclaredType(t *testing.T) {
	p := newTypeDeclTestParser(`
		type User struct {
			name: string,
			age: int
		}
		user = User {name: "Bob", age: 40}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[1].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	user, ok := p.Ctx().GetValue("user")
	if !ok {
		t.Fatal("expected user in context")
	}
	if user.TypeName() != "User" {
		t.Fatalf("expected User, got %s", user.TypeName())
	}
	t.Log(user)
}

func TestStructLiteralRejectsUnknownField(t *testing.T) {
	p := newTypeDeclTestParser(`
		type User struct { name: string }
		User {missing: true}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[1].Eval(p.Ctx()); err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestStructConstructorRejectsUnknownField(t *testing.T) {
	p := newTypeDeclTestParser(`
		type User struct { name: string }
		User({missing: true})
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[1].Eval(p.Ctx()); err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestStructFieldKeepsDeclaredTypeForMethodCall(t *testing.T) {
	p := newTypeDeclTestParser(`
		type UserID int

		fn (u: UserID) name() int {
			return u
		}

		type User struct {
			name: string
			id: UserID
			u: User
		}

		user = User {name: "Bob", id: 10,u:User{id:22} }

		fn (u: User) name() String {
			return u.name
		}

		user.name()
		user.id.name()
		user.u.id.name()
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	for _, expr := range exprs {
		t.Log(expr.Eval(p.Ctx()))
	}

	if _, err := exprs[3].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	name, err := exprs[5].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if got := name.UnsafeCastString(); got != "Bob" {
		t.Fatalf("expected Bob, got %q", got)
	}

	id, err := exprs[6].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if id.TypeName() != "int" {
		t.Fatalf("expected int, got %s", id.TypeName())
	}
	if got := id.UnsafeCastInt(); got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}
}
