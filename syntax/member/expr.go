package member_parser

import (
	"fmt"
	"strings"

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

func (m *MemberExpr) Path() ([]string, bool) {
	switch object := m.object.(type) {
	case Ident:
		return []string{string(object), string(m.property)}, true
	case interface{ Path() ([]string, bool) }:
		path, ok := object.Path()
		if !ok {
			return nil, false
		}
		return append(path, string(m.property)), true
	default:
		return nil, false
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
	case ast.MetaTypeValue:
		switch string(m.property) {
		case "isptr":
			return value.NewType(v.IsPtr), nil
		}
		if v.Def != nil && v.Def.Kind == ast.EnumType {
			variant, ok := enumVariant(
				*v.Def,
				string(m.property),
			)
			if !ok {
				return value.NewTypeNil(), fmt.Errorf(
					"variant %s does not exist on enum %s",
					m.property,
					v.Def.Name,
				)
			}
			if len(variant.Fields) != 0 {
				return value.NewTypeNil(), fmt.Errorf(
					"enum variant %s.%s expects %d values",
					v.Def.Name,
					variant.Name,
					len(variant.Fields),
				)
			}
			return value.NewTypeWithExplicit(
				ast.EnumValue{
					TypeName: v.Def.Name,
					Variant:  variant.Name,
				},
				v.Def.Name,
			), nil
		}
	}

	return value.NewTypeNil(), nil
}

func (MemberExpr) Type(ctx ast.Ctx) string {
	return "member"
}

func (m MemberExpr) ValueType(ctx ast.Ctx) ast.TypeRef {
	if string(m.property) == "isptr" {
		return ast.TypeRef{Name: "bool"}
	}

	if enumType, _, ok := m.enumVariant(ctx); ok {
		return ast.TypeRef{Name: enumType.Name}
	}

	var objectType ast.TypeRef

	switch object := m.object.(type) {
	case Ident:
		if val, ok := ctx.GetValue(string(object)); ok {
			objectType = ast.ParseTypeRef(val.TypeName())
		}
	case interface {
		ValueType(ast.Ctx) ast.TypeRef
	}:
		objectType = object.ValueType(ctx)
	}

	def, ok := ctx.GetType(objectType.Name)
	if !ok {
		return ast.TypeRef{}
	}

	return def.Fields[string(m.property)].Type
}

func (m MemberExpr) PrintGO(ctx ast.Ctx) (string, error) {
	if enumType, variant, ok := m.enumVariant(ctx); ok {
		if len(variant.Fields) != 0 {
			return "", fmt.Errorf(
				"enum variant %s.%s expects %d values",
				enumType.Name,
				variant.Name,
				len(variant.Fields),
			)
		}
		return enumType.Name + "(" +
			ast.EnumVariantGoName(
				enumType.Name,
				variant.Name,
			) + "{})", nil
	}

	object, err := m.object.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	return object + "." + string(m.property), nil
}

func (m MemberExpr) CallValueType(ctx ast.Ctx) ast.TypeRef {
	if enumType, _, ok := m.enumVariant(ctx); ok {
		return ast.TypeRef{Name: enumType.Name}
	}
	return ast.TypeRef{}
}

func (m MemberExpr) PrintGOEnumCall(
	ctx ast.Ctx,
	args []ast.Expr,
) (string, bool, error) {
	enumType, variant, ok := m.enumVariant(ctx)
	if !ok {
		return "", false, nil
	}
	if len(args) != len(variant.Fields) {
		return "", true, fmt.Errorf(
			"enum variant %s.%s expects %d values, got %d",
			enumType.Name,
			variant.Name,
			len(variant.Fields),
			len(args),
		)
	}

	fields := make([]string, 0, len(args))
	for index, arg := range args {
		actual := memberExprValueType(ctx, arg)
		expected := variant.Fields[index]
		if !actual.IsZero() &&
			actual.GoString(ctx) != expected.GoString(ctx) {
			return "", true, fmt.Errorf(
				"enum variant %s.%s field %d expects %s, got %s",
				enumType.Name,
				variant.Name,
				index,
				expected.String(),
				actual.String(),
			)
		}
		printed, err := arg.PrintGO(ctx)
		if err != nil {
			return "", true, err
		}
		fields = append(
			fields,
			fmt.Sprintf("V%d: %s", index, printed),
		)
	}

	return enumType.Name + "(" +
		ast.EnumVariantGoName(
			enumType.Name,
			variant.Name,
		) + "{" + strings.Join(fields, ", ") + "})", true, nil
}

func (m MemberExpr) enumVariant(
	ctx ast.Ctx,
) (ast.TypeDef, ast.EnumVariantDef, bool) {
	object, ok := m.object.(Ident)
	if !ok {
		return ast.TypeDef{}, ast.EnumVariantDef{}, false
	}
	def, declared := ctx.GetType(string(object))
	if !declared || def.Kind != ast.EnumType {
		return ast.TypeDef{}, ast.EnumVariantDef{}, false
	}
	variant, exists := enumVariant(def, string(m.property))
	return def, variant, exists
}

func enumVariant(
	def ast.TypeDef,
	name string,
) (ast.EnumVariantDef, bool) {
	for _, variant := range def.Variants {
		if variant.Name == name {
			return variant, true
		}
	}
	return ast.EnumVariantDef{}, false
}

func memberExprValueType(ctx ast.Ctx, expr ast.Expr) ast.TypeRef {
	if typed, ok := expr.(interface {
		ValueType(ast.Ctx) ast.TypeRef
	}); ok {
		ref := typed.ValueType(ctx)
		if !ref.IsZero() {
			return ref
		}
	}
	switch expr.(type) {
	case literal_parser.Int:
		return ast.TypeRef{Name: "int"}
	case literal_parser.Float64:
		return ast.TypeRef{Name: "float64"}
	case literal_parser.String:
		return ast.TypeRef{Name: "string"}
	case literal_parser.Bool:
		return ast.TypeRef{Name: "bool"}
	}
	return ast.TypeRef{}
}
