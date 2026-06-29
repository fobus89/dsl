package literal_parser

import (
	"fmt"
	"math"
	"strings"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/value"
)

type Float64 float64

func NewFloat64Expr(value float64) Float64 {
	return Float64(value)
}

func (f Float64) Eval(ctx ast.Ctx) (value.Type, error) {
	return value.NewType(float64(f)), nil
}

func (f Float64) Type(ctx ast.Ctx) string {
	return "float64"
}

type Int int

func NewIntExpr(value int) Int {
	return Int(value)
}

func (i Int) Eval(ctx ast.Ctx) (value.Type, error) {
	return value.NewType(int64(i)), nil
}

func (f Int) Type(ctx ast.Ctx) string {
	return "int"
}

type String string

func NewStringExpr(value string) String {
	return String(value)
}

func (s String) Eval(ctx ast.Ctx) (value.Type, error) {
	return value.NewType(string(s)), nil
}

func (s String) Type(ctx ast.Ctx) string {
	return "string"
}

type Bool bool

func NewBoolExpr(value bool) Bool {
	return Bool(value)
}

func (b Bool) Eval(ctx ast.Ctx) (value.Type, error) {
	return value.NewType(bool(b)), nil
}

func (s Bool) Type(ctx ast.Ctx) string {
	return "bool"
}

type Nil struct{}

func NewNilExpr() Nil {
	return Nil{}
}

func (Nil) Eval(ctx ast.Ctx) (value.Type, error) {
	return value.NewTypeNil(), nil
}

func (Nil) Type(ctx ast.Ctx) string {
	return "nil"
}

type Nan struct{}

func NewNanExpr() Nan {
	return Nan{}
}

func (Nan) Eval(ctx ast.Ctx) (value.Type, error) {
	return value.NewType(math.NaN()), nil
}

func (Nan) Type(ctx ast.Ctx) string {
	return "nan"
}

type Ident String

func NewIdentExpr(name string) Ident {
	return Ident(name)
}

func (i Ident) Eval(ctx ast.Ctx) (value.Type, error) {

	v, ok := ctx.GetValue(string(i))
	{
		if !ok {
			return value.NewTypeNil(), fmt.Errorf("ident %s not found eval", i)
		}
	}

	return v, nil
}

func (Ident) Type(ctx ast.Ctx) string {
	return "ident"
}

type FormatStringExpr struct {
	parts []ast.Expr
}

func NewFormatStringExpr(parts []ast.Expr) *FormatStringExpr {
	return &FormatStringExpr{
		parts: parts,
	}
}

func (f *FormatStringExpr) Eval(ctx ast.Ctx) (value.Type, error) {

	var builder strings.Builder

	for _, part := range f.parts {
		v, err := part.Eval(ctx)
		{
			if err != nil {
				return value.NewTypeNil(), err
			}
		}

		builder.WriteString(v.UnsafeCastString())
	}

	return value.NewType(builder.String()), nil
}

func (FormatStringExpr) Type(ctx ast.Ctx) string {
	return "formatString"
}
