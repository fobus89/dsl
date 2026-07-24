package funcdecl_parser

import (
	"strings"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	member_parser "github.com/fobus89/dsl/syntax/member"
	"github.com/fobus89/dsl/token"
)

func TestPrintGORejectsMethodWithSameNameAsField(t *testing.T) {
	ctx := parser.NewCtx()
	ctx.SetType("User", ast.TypeDef{
		Name: "User",
		Kind: ast.StructType,
		Fields: map[string]ast.FieldDef{
			"name": {Type: ast.TypeRef{Name: "string"}},
		},
	})

	decl := NewFuncDecl(
		ctx,
		&Param{
			Name: literal_parser.NewIdentExpr("u"),
			Type: &ast.TypeRef{Name: "User"},
		},
		literal_parser.NewIdentExpr("name"),
		nil,
		&ast.TypeRef{Name: "string"},
		nil,
	)

	_, err := decl.PrintGO(ctx)
	if err == nil {
		t.Fatal("expected field/method name collision")
	}
	if !strings.Contains(err.Error(), "both field and method named name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrintGOUsesMemberTypeForStringConcatenation(t *testing.T) {
	ctx := parser.NewCtx()
	ctx.SetType("User", ast.TypeDef{
		Name: "User",
		Kind: ast.StructType,
		Fields: map[string]ast.FieldDef{
			"name": {Type: ast.TypeRef{Name: "string"}},
		},
	})

	receiver := literal_parser.NewIdentExpr("u")
	decl := NewFuncDecl(
		ctx,
		&Param{
			Name: receiver,
			Type: &ast.TypeRef{Name: "User"},
		},
		literal_parser.NewIdentExpr("Name"),
		nil,
		&ast.TypeRef{Name: "String"},
		[]ast.Expr{
			NewReturnStmt(
				binary_parser.NewBinaryExpr(
					token.PLUS,
					member_parser.NewMemberExpr(
						receiver,
						literal_parser.NewIdentExpr("name"),
					),
					literal_parser.NewIntExpr(2121),
				),
			),
		},
	)

	got, err := decl.PrintGO(ctx)
	if err != nil {
		t.Fatal(err)
	}

	want := "func (u User) Name() string {\n" +
		"\treturn (u.name + strconv.FormatInt(int64(2121), 10))\n" +
		"}"
	if got != want {
		t.Fatalf("PrintGO() = %q, want %q", got, want)
	}
}
