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
			name+" "+d.Def.Fields[name].Type.GoString(),
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

	result := make(map[string]value.Type, len(def.Fields))
	for _, field := range s.Fields {
		fieldDef, ok := def.Fields[field.Name]
		if !ok {
			return value.NewTypeNil(), fmt.Errorf(
				"field %s does not exist on type %s",
				field.Name,
				typeName,
			)
		}

		evaluated, err := evalStructField(ctx, fieldDef, field.Value)
		if err != nil {
			return value.NewTypeNil(), err
		}
		result[field.Name] = evaluated
	}

	names := make([]string, 0, len(def.Fields))
	for name := range def.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if _, initialized := result[name]; initialized {
			continue
		}

		fieldDef := def.Fields[name]
		if fieldDef.Default == nil {
			continue
		}

		evaluated, err := evalStructField(
			ctx,
			fieldDef,
			fieldDef.Default,
		)
		if err != nil {
			return value.NewTypeNil(), err
		}
		result[name] = evaluated
	}

	return value.NewStructType(result, typeName), nil
}

func evalStructField(
	ctx ast.Ctx,
	fieldDef ast.FieldDef,
	expr ast.Expr,
) (value.Type, error) {
	evaluated, err := expr.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	fieldType := fieldDef.Type
	if _, declared := ctx.GetType(fieldType.Name); declared &&
		evaluated.TypeName() != fieldType.Name {
		evaluated = value.NewTypeWithExplicit(
			evaluated.Any(),
			fieldType.Name,
		)
	}

	return evaluated, nil
}

func (*StructLiteral) Type(ast.Ctx) string {
	return "struct_literal"
}

func (s *StructLiteral) PrintGO(ctx ast.Ctx) (string, error) {
	typeName := string(s.TypeName)
	def, ok := ctx.GetType(typeName)
	if !ok {
		return "", fmt.Errorf("type %s not found", typeName)
	}

	fields := make([]string, 0, len(def.Fields))
	initialized := make(map[string]struct{}, len(s.Fields))
	for _, field := range s.Fields {
		fieldDef, ok := def.Fields[field.Name]
		if !ok {
			return "", fmt.Errorf(
				"field %s does not exist on type %s",
				field.Name,
				typeName,
			)
		}

		printed, err := field.Value.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		fields = append(
			fields,
			field.Name+": "+
				printFieldValue(fieldDef.Type, field.Value, printed),
		)
		initialized[field.Name] = struct{}{}
	}

	names := make([]string, 0, len(def.Fields))
	for name := range def.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if _, exists := initialized[name]; exists {
			continue
		}

		fieldDef := def.Fields[name]
		if fieldDef.Default == nil {
			continue
		}

		printed, err := fieldDef.Default.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		fields = append(
			fields,
			name+": "+
				printFieldValue(fieldDef.Type, fieldDef.Default, printed),
		)
	}

	return typeName + "{" + strings.Join(fields, ", ") + "}", nil
}

func printFieldValue(
	fieldType ast.TypeRef,
	expr ast.Expr,
	printed string,
) string {
	if !fieldType.IsPtr {
		return printed
	}
	if _, isNil := expr.(literal_parser.Nil); isNil {
		return "nil"
	}

	return "new(" + printed + ")"
}
