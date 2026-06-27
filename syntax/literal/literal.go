package literal_parser

import (
	"fmt"

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
