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

	if def.Kind == ast.EnumType {
		registerEnumVariants(ctx, def)
		return decl
	}

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

func registerEnumVariants(ctx ast.Ctx, def ast.TypeDef) {
	for _, variant := range def.Variants {
		variant := variant
		if variant.FieldNames != nil {
			continue
		}
		ctx.SetMethod(
			def.Name,
			variant.Name,
			func(args ...value.Type) (value.Type, error) {
				payload := args
				if len(payload) > 0 {
					payload = payload[1:] // enum type receiver
				}
				if len(payload) != len(variant.Fields) {
					return value.NewTypeNil(), fmt.Errorf(
						"enum variant %s.%s expects %d values, got %d",
						def.Name,
						variant.Name,
						len(variant.Fields),
						len(payload),
					)
				}
				for index, fieldType := range variant.Fields {
					if !enumValueMatchesType(
						ctx,
						payload[index],
						fieldType,
					) {
						return value.NewTypeNil(), fmt.Errorf(
							"enum variant %s.%s field %d expects %s, got %s",
							def.Name,
							variant.Name,
							index,
							fieldType.String(),
							payload[index].TypeName(),
						)
					}
				}
				return value.NewTypeWithExplicit(
					ast.EnumValue{
						TypeName: def.Name,
						Variant:  variant.Name,
						Payload:  payload,
					},
					def.Name,
				), nil
			},
		)
	}
}

func enumValueMatchesType(
	ctx ast.Ctx,
	got value.Type,
	expected ast.TypeRef,
) bool {
	gotRef := ast.ParseTypeRef(got.TypeName())
	if gotRef.GoString(ctx) == expected.GoString(ctx) {
		return true
	}
	if enumNumericType(gotRef) && enumNumericType(expected) {
		return true
	}
	if def, declared := ctx.GetType(expected.Name); declared &&
		def.Kind == ast.AliasType {
		return gotRef.GoString(ctx) ==
			def.Underlying.GoString(ctx)
	}
	return false
}

func enumNumericType(ref ast.TypeRef) bool {
	if ref.IsPtr || ref.Kind != ast.NamedTypeRef {
		return false
	}
	switch strings.ToLower(ref.Name) {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "rune", "byte":
		return true
	}
	return false
}

func (*TypeDecl) Eval(ast.Ctx) (value.Type, error) {
	return value.NewTypeNil(), nil
}

func (*TypeDecl) Type(ast.Ctx) string {
	return "type_decl"
}

func (d *TypeDecl) PrintGO(ctx ast.Ctx) (string, error) {
	if len(d.Def.TypeParams) != 0 {
		instances := ctx.GetGenericInstances(
			"type:" + d.Def.Name,
		)
		declarations := make([]string, 0, len(instances))
		for _, args := range instances {
			if err := ast.ValidateTypeArguments(
				ctx,
				"type "+d.Def.Name,
				d.Def.TypeParams,
				args,
			); err != nil {
				return "", err
			}
			instantiated := instantiateTypeDef(d.Def, args)
			printed, err := (&TypeDecl{
				Def: instantiated,
			}).PrintGO(ctx)
			if err != nil {
				return "", err
			}
			declarations = append(declarations, printed)
		}
		return strings.Join(declarations, "\n\n"), nil
	}

	typeName := d.Def.Name + printTypeParams(
		ctx,
		d.Def.TypeParams,
	)
	if d.Def.Kind == ast.AliasType {
		if d.Def.Underlying.IsMeta() {
			return "", fmt.Errorf(
				"type %s contains compile-time type values",
				d.Def.Name,
			)
		}
		return fmt.Sprintf(
			"type %s %s",
			typeName,
			d.Def.Underlying.GoString(ctx),
		), nil
	}
	if d.Def.Kind == ast.EnumType {
		return printGOEnum(ctx, d.Def)
	}

	names := make([]string, 0, len(d.Def.Fields))
	for name := range d.Def.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	fields := make([]string, 0, len(names))
	for _, name := range names {
		if d.Def.Fields[name].Type.IsMeta() {
			return "", fmt.Errorf(
				"field %s.%s contains compile-time type values",
				d.Def.Name,
				name,
			)
		}
		fields = append(
			fields,
			name+" "+d.Def.Fields[name].Type.GoString(ctx),
		)
	}

	return fmt.Sprintf(
		"type %s struct {\n\t%s\n}",
		typeName,
		strings.Join(fields, "\n\t"),
	), nil
}

func instantiateTypeDef(
	def ast.TypeDef,
	args []ast.TypeRef,
) ast.TypeDef {
	instantiated := def
	instantiated.Name = ast.MangleGenericName(def.Name, args)
	instantiated.TypeParams = nil
	instantiated.Underlying = ast.SubstituteType(
		def.Underlying,
		def.TypeParams,
		args,
	)
	instantiated.Fields = make(
		map[string]ast.FieldDef,
		len(def.Fields),
	)
	for name, field := range def.Fields {
		field.Type = ast.SubstituteType(
			field.Type,
			def.TypeParams,
			args,
		)
		instantiated.Fields[name] = field
	}
	instantiated.Variants = append(
		[]ast.EnumVariantDef(nil),
		def.Variants...,
	)
	for index := range instantiated.Variants {
		for field := range instantiated.Variants[index].Fields {
			instantiated.Variants[index].Fields[field] =
				ast.SubstituteType(
					instantiated.Variants[index].Fields[field],
					def.TypeParams,
					args,
				)
		}
	}
	return instantiated
}

func printTypeParams(
	ctx ast.Ctx,
	params []ast.TypeParam,
) string {
	if len(params) == 0 {
		return ""
	}
	printed := make([]string, 0, len(params))
	for _, param := range params {
		printed = append(
			printed,
			param.Name+" "+param.Constraint.GoString(ctx),
		)
	}
	return "[" + strings.Join(printed, ", ") + "]"
}

func printGOEnum(ctx ast.Ctx, def ast.TypeDef) (string, error) {
	var declarations []string
	declarations = append(
		declarations,
		"type "+def.Name+printTypeParams(ctx, def.TypeParams)+
			" interface {\n\tis"+def.Name+"()\n}",
	)

	for _, variant := range def.Variants {
		goName := ast.EnumVariantGoName(def.Name, variant.Name)
		fields := make([]string, 0, len(variant.Fields))
		for index, fieldType := range variant.Fields {
			if fieldType.IsMeta() {
				return "", fmt.Errorf(
					"enum variant %s.%s contains compile-time type values",
					def.Name,
					variant.Name,
				)
			}
			fieldName := fmt.Sprintf("V%d", index)
			if variant.FieldNames != nil {
				fieldName = variant.FieldNames[index]
			}
			fields = append(
				fields,
				fmt.Sprintf(
					"%s %s",
					fieldName,
					fieldType.GoString(ctx),
				),
			)
		}
		body := ""
		if len(fields) != 0 {
			body = "\n\t" + strings.Join(fields, "\n\t") + "\n"
		}
		declarations = append(
			declarations,
			"type "+goName+" struct {"+body+"}",
			"func ("+goName+") is"+def.Name+"() {}",
		)
	}
	return strings.Join(declarations, "\n\n"), nil
}

type FieldValue struct {
	Name  string
	Value ast.Expr
}

type EnumStructLiteral struct {
	EnumName    string
	VariantName string
	Fields      []FieldValue
}

func NewEnumStructLiteral(
	enumName string,
	variantName string,
	fields []FieldValue,
) *EnumStructLiteral {
	return &EnumStructLiteral{
		EnumName:    enumName,
		VariantName: variantName,
		Fields:      fields,
	}
}

func (e *EnumStructLiteral) Eval(ctx ast.Ctx) (value.Type, error) {
	def, declared := ctx.GetType(e.EnumName)
	if !declared || def.Kind != ast.EnumType {
		return value.NewTypeNil(), fmt.Errorf(
			"type %s is not an enum",
			e.EnumName,
		)
	}
	variant, exists := enumVariantDef(def, e.VariantName)
	if !exists || variant.FieldNames == nil {
		return value.NewTypeNil(), fmt.Errorf(
			"%s.%s is not a struct enum variant",
			e.EnumName,
			e.VariantName,
		)
	}
	ordered, err := orderEnumStructFields(
		variant,
		e.Fields,
	)
	if err != nil {
		return value.NewTypeNil(), err
	}
	payload := make([]value.Type, len(ordered))
	for index, field := range ordered {
		evaluated, err := field.Value.Eval(ctx)
		if err != nil {
			return value.NewTypeNil(), err
		}
		if !enumValueMatchesType(
			ctx,
			evaluated,
			variant.Fields[index],
		) {
			return value.NewTypeNil(), fmt.Errorf(
				"enum variant %s.%s field %s expects %s, got %s",
				e.EnumName,
				e.VariantName,
				field.Name,
				variant.Fields[index].String(),
				evaluated.TypeName(),
			)
		}
		payload[index] = evaluated
	}
	return value.NewTypeWithExplicit(
		ast.EnumValue{
			TypeName: e.EnumName,
			Variant:  e.VariantName,
			Payload:  payload,
		},
		e.EnumName,
	), nil
}

func (*EnumStructLiteral) Type(ast.Ctx) string {
	return "enum_struct_literal"
}

func (e *EnumStructLiteral) ValueType(ast.Ctx) ast.TypeRef {
	return ast.TypeRef{Name: e.EnumName}
}

func (e *EnumStructLiteral) PrintGO(
	ctx ast.Ctx,
) (string, error) {
	def, declared := ctx.GetType(e.EnumName)
	if !declared || def.Kind != ast.EnumType {
		return "", fmt.Errorf("type %s is not an enum", e.EnumName)
	}
	variant, exists := enumVariantDef(def, e.VariantName)
	if !exists || variant.FieldNames == nil {
		return "", fmt.Errorf(
			"%s.%s is not a struct enum variant",
			e.EnumName,
			e.VariantName,
		)
	}
	ordered, err := orderEnumStructFields(variant, e.Fields)
	if err != nil {
		return "", err
	}
	fields := make([]string, 0, len(ordered))
	for index, field := range ordered {
		printed, err := field.Value.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		fields = append(
			fields,
			field.Name+": "+printed,
		)
		_ = index
	}
	return e.EnumName + "(" +
		ast.EnumVariantGoName(
			e.EnumName,
			e.VariantName,
		) + "{" + strings.Join(fields, ", ") + "})", nil
}

func enumVariantDef(
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

func orderEnumStructFields(
	variant ast.EnumVariantDef,
	fields []FieldValue,
) ([]FieldValue, error) {
	byName := make(map[string]FieldValue, len(fields))
	for _, field := range fields {
		byName[field.Name] = field
	}
	ordered := make([]FieldValue, len(variant.FieldNames))
	for index, name := range variant.FieldNames {
		field, exists := byName[name]
		if !exists {
			return nil, fmt.Errorf(
				"missing enum variant field %s",
				name,
			)
		}
		ordered[index] = field
		delete(byName, name)
	}
	for name := range byName {
		return nil, fmt.Errorf("unknown enum variant field %s", name)
	}
	return ordered, nil
}

type StructLiteral struct {
	TypeName Ident
	TypeRef  ast.TypeRef
	Fields   []FieldValue
}

func NewStructLiteral(typeName Ident, fields []FieldValue) *StructLiteral {
	return NewStructLiteralRef(
		typeName,
		ast.TypeRef{Name: string(typeName)},
		fields,
	)
}

func NewStructLiteralRef(
	typeName Ident,
	typeRef ast.TypeRef,
	fields []FieldValue,
) *StructLiteral {
	return &StructLiteral{
		TypeName: typeName,
		TypeRef:  typeRef,
		Fields:   fields,
	}
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
	if err := validateTypeArgumentCount(
		ctx,
		s.TypeRef,
		def,
	); err != nil {
		return value.NewTypeNil(), err
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
		fieldDef.Type = ast.SubstituteType(
			fieldDef.Type,
			def.TypeParams,
			s.TypeRef.Args,
		)

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
		fieldDef.Type = ast.SubstituteType(
			fieldDef.Type,
			def.TypeParams,
			s.TypeRef.Args,
		)

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

	return value.NewStructType(result, s.TypeRef.String()), nil
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
	if !structValueMatches(ctx, fieldType, evaluated) {
		return value.NewTypeNil(), fmt.Errorf(
			"cannot use %s as struct field type %s",
			evaluated.TypeName(),
			fieldType.String(),
		)
	}
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

func (s *StructLiteral) ValueType(ast.Ctx) ast.TypeRef {
	if !s.TypeRef.IsZero() {
		return s.TypeRef
	}
	return ast.TypeRef{Name: string(s.TypeName)}
}

func (s *StructLiteral) PrintGO(ctx ast.Ctx) (string, error) {
	typeName := string(s.TypeName)
	def, ok := ctx.GetType(typeName)
	if !ok {
		return "", fmt.Errorf("type %s not found", typeName)
	}
	if err := validateTypeArgumentCount(
		ctx,
		s.TypeRef,
		def,
	); err != nil {
		return "", err
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
		fieldDef.Type = ast.SubstituteType(
			fieldDef.Type,
			def.TypeParams,
			s.TypeRef.Args,
		)
		if actual := structExprType(ctx, field.Value); !actual.IsZero() &&
			!structTypesCompatible(ctx, fieldDef.Type, actual) {
			return "", fmt.Errorf(
				"field %s.%s expects %s, got %s",
				typeName,
				field.Name,
				fieldDef.Type.String(),
				actual.String(),
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
		fieldDef.Type = ast.SubstituteType(
			fieldDef.Type,
			def.TypeParams,
			s.TypeRef.Args,
		)

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

	printedType := s.TypeRef.GoString(ctx)
	return printedType + "{" + strings.Join(fields, ", ") + "}", nil
}

func structExprType(ctx ast.Ctx, expr ast.Expr) ast.TypeRef {
	if typed, ok := expr.(interface {
		ValueType(ast.Ctx) ast.TypeRef
	}); ok {
		return typed.ValueType(ctx)
	}
	return ast.TypeRef{}
}

func structTypesCompatible(
	ctx ast.Ctx,
	expected ast.TypeRef,
	actual ast.TypeRef,
) bool {
	if expected.GoString(ctx) == actual.GoString(ctx) {
		return true
	}
	if def, declared := ctx.GetType(expected.Name); declared &&
		def.Kind == ast.AliasType {
		expected = def.Underlying
	}
	if def, declared := ctx.GetType(actual.Name); declared &&
		def.Kind == ast.AliasType {
		actual = def.Underlying
	}
	return enumNumericType(expected) && enumNumericType(actual)
}

func structValueMatches(
	ctx ast.Ctx,
	expected ast.TypeRef,
	actual value.Type,
) bool {
	if expected.IsPtr {
		expected.IsPtr = false
	}
	actualRef := ast.ParseTypeRef(actual.TypeName())
	if structTypesCompatible(ctx, expected, actualRef) {
		return true
	}
	switch strings.ToLower(expected.Name) {
	case "string":
		return actual.IsString()
	case "bool":
		return actual.IsBool()
	}
	return false
}

func validateTypeArgumentCount(
	ctx ast.Ctx,
	ref ast.TypeRef,
	def ast.TypeDef,
) error {
	return ast.ValidateTypeArguments(
		ctx,
		"type "+def.Name,
		def.TypeParams,
		ref.Args,
	)
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
