package literal_parser

import (
	"fmt"
	"math"
	"strconv"
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

func (f Float64) PrintGO(ast.Ctx) (string, error) {
	return strconv.FormatFloat(float64(f), 'f', -1, 64), nil
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

func (i Int) PrintGO(ast.Ctx) (string, error) {
	return strconv.FormatInt(int64(i), 10), nil
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

func (s String) PrintGO(ast.Ctx) (string, error) {
	return strconv.Quote(string(s)), nil
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

func (b Bool) PrintGO(ast.Ctx) (string, error) {
	return strconv.FormatBool(bool(b)), nil
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

func (Nil) PrintGO(ast.Ctx) (string, error) {
	return "nil", nil
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

func (Nan) PrintGO(ast.Ctx) (string, error) {
	return "math.NaN()", nil
}

type Undefined struct{}

func NewUndefinedExpr() Undefined {
	return Undefined{}
}

func (Undefined) Eval(ctx ast.Ctx) (value.Type, error) {
	return value.NewTypeUndefined(), nil
}

func (Undefined) Type(ctx ast.Ctx) string {
	return "undefined"
}

func (Undefined) PrintGO(ast.Ctx) (string, error) {
	return "nil", nil
}

type Ident String

func NewIdentExpr(name string) Ident {
	return Ident(name)
}

func (i Ident) Eval(ctx ast.Ctx) (value.Type, error) {

	v, ok := ctx.GetValue(string(i))
	{
		if !ok {
			if def, declared := ctx.GetType(string(i)); declared {
				return value.NewTypeWithExplicit(
					ast.NewMetaTypeValue(
						ast.TypeRef{Name: string(i)},
						&def,
					),
					ast.MetaTypeName,
				), nil
			}
			if _, declared := ctx.GetFunc(string(i)); declared {
				return value.NewTypeWithExplicit(
					ast.NewFuncMetaTypeValue(string(i)),
					ast.MetaTypeName,
				), nil
			}
			return value.NewTypeNil(), fmt.Errorf(
				"ident %s not found eval",
				i,
			)
		}
	}

	return v, nil
}

func (Ident) Type(ctx ast.Ctx) string {
	return "ident"
}

func (i Ident) ValueType(ctx ast.Ctx) ast.TypeRef {
	if val, ok := ctx.GetValue(string(i)); ok {
		return ast.ParseTypeRef(val.TypeName())
	}
	if _, declared := ctx.GetType(string(i)); declared {
		return ast.TypeRef{Name: ast.MetaTypeName}
	}
	if _, declared := ctx.GetFunc(string(i)); declared {
		return ast.TypeRef{Name: ast.FuncTypeName}
	}
	return ast.TypeRef{}
}

func (i Ident) PrintGO(ast.Ctx) (string, error) {
	return string(i), nil
}

func (i Ident) StructTypeName() string {
	return string(i)
}

type TypeLiteral struct {
	Ref ast.TypeRef
}

func NewTypeLiteral(ref ast.TypeRef) TypeLiteral {
	return TypeLiteral{Ref: ref}
}

func (t TypeLiteral) Eval(ctx ast.Ctx) (value.Type, error) {
	if def, declared := ctx.GetType(t.Ref.Name); declared {
		return value.NewTypeWithExplicit(
			ast.NewMetaTypeValue(t.Ref, &def),
			ast.MetaTypeName,
		), nil
	}
	return value.NewTypeWithExplicit(
		ast.NewMetaTypeValue(t.Ref, nil),
		ast.MetaTypeName,
	), nil
}

func (TypeLiteral) Type(ast.Ctx) string {
	return ast.MetaTypeName
}

func (TypeLiteral) ValueType(ast.Ctx) ast.TypeRef {
	return ast.TypeRef{Name: ast.MetaTypeName}
}

func (t TypeLiteral) PrintGO(ast.Ctx) (string, error) {
	return "", fmt.Errorf(
		"type value %s exists only at compile time",
		t.Ref.String(),
	)
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

func (f FormatStringExpr) PrintGO(ctx ast.Ctx) (string, error) {
	parts := make([]string, 0, len(f.parts))
	for _, part := range f.parts {
		printed, err := part.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		parts = append(parts, printed)
	}

	return "fmt.Sprint(" + strings.Join(parts, ", ") + ")", nil
}
