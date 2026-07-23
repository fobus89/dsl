package typedecl_parser

import (
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
)

func TestTypeDeclPrintGO(t *testing.T) {
	ctx := parser.NewCtx()
	decl := NewTypeDecl(ctx, ast.TypeDef{
		Name: "User",
		Kind: ast.StructType,
		Fields: map[string]ast.FieldDef{
			"name": {Type: ast.TypeRef{Name: "string"}},
			"id":   {Type: ast.TypeRef{Name: "int64"}},
		},
	})

	want := "type User struct {\n\tid int64\n\tname string\n}"
	got, err := decl.PrintGO(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("PrintGO() = %q, want %q", got, want)
	}
}

func TestStructLiteralPrintGO(t *testing.T) {
	ctx := parser.NewCtx()
	ctx.SetType("User", ast.TypeDef{
		Name: "User",
		Kind: ast.StructType,
		Fields: map[string]ast.FieldDef{
			"id": {
				Type: ast.TypeRef{Name: "int64"},
			},
			"name": {
				Type: ast.TypeRef{Name: "string"},
			},
		},
	})
	literal := NewStructLiteral(
		literal_parser.NewIdentExpr("User"),
		[]FieldValue{
			{
				Name:  "id",
				Value: literal_parser.NewIntExpr(7),
			},
			{
				Name:  "name",
				Value: literal_parser.NewStringExpr("Ali"),
			},
		},
	)

	want := `User{id: 7, name: "Ali"}`
	got, err := literal.PrintGO(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("PrintGO() = %q, want %q", got, want)
	}
}
