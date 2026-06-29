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

	switch v := obj.Any().(type) {

	case map[string]any:
		val, ok := v[string(m.property)]
		if !ok {
			return value.NewTypeNil(), nil
		}

		return value.NewType(val), nil
	}

	return value.NewTypeNil(), nil
}

func (MemberExpr) Type(ctx ast.Ctx) string {
	return "member"
}
