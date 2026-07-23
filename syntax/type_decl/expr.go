package typedecl_parser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type Ident = literal_parser.Ident

type TypeDecl struct {
	Def ast.TypeDef
}

func NewTypeDecl(ctx ast.Ctx, def ast.TypeDef) *TypeDecl {
	decl := &TypeDecl{Def: def}
	ctx.SetType(def.Name, def)

	ctx.SetFunc(def.Name, func(args ...value.Type) (value.Type, error) {
		if len(args) != 1 {
			return value.NewTypeNil(), fmt.Errorf(
				"type %s expects exactly one value, got %d",
				def.Name,
				len(args),
			)
		}

		if def.Kind == ast.StructType {
			fields, ok := args[0].Map()
			if !ok {
				return value.NewTypeNil(), fmt.Errorf(
					"type %s expects a struct value, got %s",
					def.Name,
					args[0].TypeName(),
				)
			}
			for field := range fields {
				if _, ok := def.Fields[field]; !ok {
					return value.NewTypeNil(), fmt.Errorf(
						"field %s does not exist on type %s",
						field,
						def.Name,
					)
				}
			}
		}

		return value.NewTypeWithExplicit(args[0].Any(), def.Name), nil
	})

	return decl
}

func (*TypeDecl) Eval(ast.Ctx) (value.Type, error) {
	return value.NewTypeNil(), nil
}

func (*TypeDecl) Type(ast.Ctx) string {
	return "type_decl"
}

func (d *TypeDecl) PrintGO(ast.Ctx) (string, error) {
	if d.Def.Kind == ast.AliasType {
		return fmt.Sprintf(
			"type %s %s",
			d.Def.Name,
			d.Def.Underlying.GoString(),
		), nil
	}

	names := make([]string, 0, len(d.Def.Fields))
	for name := range d.Def.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	fields := make([]string, 0, len(names))
	for _, name := range names {
		fields = append(
			fields,
			name+" "+d.Def.Fields[name].GoString(),
		)
	}

	return fmt.Sprintf(
		"type %s struct {\n\t%s\n}",
		d.Def.Name,
		strings.Join(fields, "\n\t"),
	), nil
}

type FieldValue struct {
	Name  string
	Value ast.Expr
}

type StructLiteral struct {
	TypeName Ident
	Fields   []FieldValue
}

func NewStructLiteral(typeName Ident, fields []FieldValue) *StructLiteral {
	return &StructLiteral{TypeName: typeName, Fields: fields}
}

func (s *StructLiteral) Eval(ctx ast.Ctx) (value.Type, error) {
	typeName := string(s.TypeName)
	def, ok := ctx.GetType(typeName)
	if !ok {
		return value.NewTypeNil(), fmt.Errorf("type %s not found", typeName)
	}
	if def.Kind != ast.StructType {
		return value.NewTypeNil(), fmt.Errorf("type %s is not a struct", typeName)
	}

	result := make(map[string]value.Type, len(s.Fields))
	for _, field := range s.Fields {
		fieldType, ok := def.Fields[field.Name]
		if !ok {
			return value.NewTypeNil(), fmt.Errorf(
				"field %s does not exist on type %s",
				field.Name,
				typeName,
			)
		}

		evaluated, err := field.Value.Eval(ctx)
		if err != nil {
			return value.NewTypeNil(), err
		}
		if _, declared := ctx.GetType(fieldType.Name); declared &&
			evaluated.TypeName() != fieldType.Name {
			evaluated = value.NewTypeWithExplicit(
				evaluated.Any(),
				fieldType.Name,
			)
		}
		result[field.Name] = evaluated
	}

	return value.NewStructType(result, typeName), nil
}

func (*StructLiteral) Type(ast.Ctx) string {
	return "struct_literal"
}

func (s *StructLiteral) PrintGO(ctx ast.Ctx) (string, error) {
	fields := make([]string, 0, len(s.Fields))
	for _, field := range s.Fields {
		printed, err := field.Value.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		fields = append(
			fields,
			field.Name+": "+printed,
		)
	}

	return string(s.TypeName) + "{" + strings.Join(fields, ", ") + "}", nil
}
