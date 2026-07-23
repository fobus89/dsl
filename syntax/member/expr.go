package member_parser

import (
	"fmt"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type Ident = literal_parser.Ident

type MemberExpr struct {
	object   ast.Expr
	property Ident
}

func NewMemberExpr(object ast.Expr, filed Ident) *MemberExpr {
	return &MemberExpr{
		object:   object,
		property: filed,
	}
}

func (m *MemberExpr) Receiver() ast.Expr {
	return m.object
}

func (m *MemberExpr) MethodName() string {
	return string(m.property)
}

func (m *MemberExpr) String() string {
	return fmt.Sprintf("%s.%s", m.object, m.property)
}

func (m *MemberExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	obj, err := m.object.Eval(ctx)
	{
		if err != nil {
			return value.NewTypeNil(), err
		}
	}

	if field, ok := obj.Field(string(m.property)); ok {
		return field, nil
	}

	switch v := obj.Any().(type) {

	case map[string]any:
		val, ok := v[string(m.property)]
		if !ok {
			return value.NewTypeUndefined(), nil
		}

		return value.NewType(val), nil
	}

	return value.NewTypeNil(), nil
}

func (MemberExpr) Type(ctx ast.Ctx) string {
	return "member"
}

func (m MemberExpr) ValueType(ctx ast.Ctx) string {
	var objectType string

	switch object := m.object.(type) {
	case Ident:
		if val, ok := ctx.GetValue(string(object)); ok {
			objectType = val.TypeName()
		}
	case interface{ ValueType(ast.Ctx) string }:
		objectType = object.ValueType(ctx)
	}

	def, ok := ctx.GetType(objectType)
	if !ok {
		return ""
	}

	return def.Fields[string(m.property)]
}

func (m MemberExpr) PrintGO(ctx ast.Ctx) (string, error) {
	object, err := m.object.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	return object + "." + string(m.property), nil
}
