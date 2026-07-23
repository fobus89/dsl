package binary_parser

import (
	"testing"

	"github.com/fobus89/dsl/parser"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/token"
	"github.com/fobus89/dsl/value"
)

func TestPrintGOStringifiesNumericIdentifier(t *testing.T) {
	ctx := parser.NewCtx()
	ctx.SetValue("value", value.NewType(int64(42)))

	expr := NewBinaryExpr(
		token.PLUS,
		literal_parser.NewIdentExpr("value"),
		literal_parser.NewStringExpr(""),
	)

	want := `(strconv.FormatInt(int64(value), 10) + "")`
	got, err := expr.PrintGO(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("PrintGO() = %q, want %q", got, want)
	}
}

func TestPrintGORecursesThroughBinaryExpressions(t *testing.T) {
	ctx := parser.NewCtx()

	expr := NewBinaryExpr(
		token.STAR,
		NewBinaryExpr(
			token.PLUS,
			literal_parser.NewIntExpr(1),
			literal_parser.NewIntExpr(2),
		),
		literal_parser.NewIntExpr(3),
	)

	want := `((1 + 2) * 3)`
	got, err := expr.PrintGO(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("PrintGO() = %q, want %q", got, want)
	}
}
