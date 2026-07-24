package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fobus89/dsl/value"
)

type Func = func(...value.Type) (value.Type, error)

type TypeKind string

const (
	AliasType  TypeKind = "alias"
	StructType TypeKind = "struct"
	EnumType   TypeKind = "enum"
)

type TypeRefKind uint8

const (
	NamedTypeRef TypeRefKind = iota
	ArrayTypeRef
	SliceTypeRef
	TupleTypeRef
)

type TypeRef struct {
	Name  string
	IsPtr bool
	Kind  TypeRefKind
	Elem  *TypeRef
	Elems []TypeRef
	Args  []TypeRef
	Len   int
}

type TypeParam struct {
	Name       string
	Constraint TypeRef
}

const (
	MetaTypeName      = "type"
	PrimitiveTypeName = "primitivetype"
	NumberTypeName    = "numbertype"
	StringTypeName    = "stringtype"
	BoolTypeName      = "booltype"
	SliceTypeName     = "slicetype"
	ArrayTypeName     = "arraytype"
	StructTypeName    = "structtype"
	AliasTypeName     = "aliastype"
	NamedTypeName     = "namedtype"
	MapTypeName       = "maptype"
	TupleTypeName     = "tupletype"
	EnumTypeName      = "enumtype"
	InterfaceTypeName = "interfacetype"
	FuncTypeName      = "functype"
	AnyTypeName       = "anytype"
	VoidTypeName      = "voidtype"
	NeverTypeName     = "nevertype"
)

type MetaTypeKind string

const (
	PrimitiveMetaType MetaTypeKind = PrimitiveTypeName
	SliceMetaType     MetaTypeKind = SliceTypeName
	ArrayMetaType     MetaTypeKind = ArrayTypeName
	StructMetaType    MetaTypeKind = StructTypeName
	AliasMetaType     MetaTypeKind = AliasTypeName
	MapMetaType       MetaTypeKind = MapTypeName
	TupleMetaType     MetaTypeKind = TupleTypeName
	EnumMetaType      MetaTypeKind = EnumTypeName
	InterfaceMetaType MetaTypeKind = InterfaceTypeName
	FuncMetaType      MetaTypeKind = FuncTypeName
	AnyMetaType       MetaTypeKind = AnyTypeName
	VoidMetaType      MetaTypeKind = VoidTypeName
	NeverMetaType     MetaTypeKind = NeverTypeName
	UnknownMetaType   MetaTypeKind = "unknowntype"
)

// MetaTypeValue is the compile-time representation of a type.
// Def is nil for built-in types and set for declared aliases/structs.
type MetaTypeValue struct {
	Ref   TypeRef
	Def   *TypeDef
	Kind  MetaTypeKind
	IsPtr bool
}

func NewMetaTypeValue(ref TypeRef, def *TypeDef) MetaTypeValue {
	return MetaTypeValue{
		Ref:   ref,
		Def:   def,
		Kind:  classifyMetaType(ref, def),
		IsPtr: ref.IsPtr,
	}
}

func NewFuncMetaTypeValue(name string) MetaTypeValue {
	return MetaTypeValue{
		Ref:  TypeRef{Name: name},
		Kind: FuncMetaType,
	}
}

func IsMetaTypeName(name string) bool {
	switch strings.ToLower(name) {
	case MetaTypeName,
		PrimitiveTypeName,
		NumberTypeName,
		StringTypeName,
		BoolTypeName,
		SliceTypeName,
		ArrayTypeName,
		StructTypeName,
		AliasTypeName,
		NamedTypeName,
		MapTypeName,
		TupleTypeName,
		EnumTypeName,
		InterfaceTypeName,
		FuncTypeName,
		AnyTypeName,
		VoidTypeName,
		NeverTypeName:
		return true
	}
	return false
}

func (m MetaTypeValue) Matches(name string) bool {
	switch strings.ToLower(name) {
	case MetaTypeName:
		return true
	case NamedTypeName:
		return m.Ref.Kind == NamedTypeRef
	case PrimitiveTypeName:
		return m.isPrimitive()
	case NumberTypeName:
		return m.matchesUnderlying(isMetaNumber)
	case StringTypeName:
		return m.matchesUnderlying(func(ref TypeRef) bool {
			return strings.EqualFold(ref.Name, "string")
		})
	case BoolTypeName:
		return m.matchesUnderlying(func(ref TypeRef) bool {
			return strings.EqualFold(ref.Name, "bool")
		})
	default:
		return strings.EqualFold(string(m.Kind), name)
	}
}

func (m MetaTypeValue) MethodReceiverTypes() []string {
	types := []string{string(m.Kind)}

	for _, candidate := range []string{
		NumberTypeName,
		StringTypeName,
		BoolTypeName,
		PrimitiveTypeName,
		NamedTypeName,
		MetaTypeName,
	} {
		if candidate != string(m.Kind) && m.Matches(candidate) {
			types = append(types, candidate)
		}
	}
	return types
}

func (m MetaTypeValue) isPrimitive() bool {
	return m.matchesUnderlying(func(ref TypeRef) bool {
		return isMetaNumber(ref) ||
			strings.EqualFold(ref.Name, "string") ||
			strings.EqualFold(ref.Name, "bool") ||
			strings.EqualFold(ref.Name, "char")
	})
}

func (m MetaTypeValue) matchesUnderlying(
	match func(TypeRef) bool,
) bool {
	if match(m.Ref) {
		return true
	}
	return m.Def != nil &&
		m.Def.Kind == AliasType &&
		match(m.Def.Underlying)
}

func classifyMetaType(ref TypeRef, def *TypeDef) MetaTypeKind {
	switch ref.Kind {
	case SliceTypeRef:
		return SliceMetaType
	case ArrayTypeRef:
		return ArrayMetaType
	case TupleTypeRef:
		return TupleMetaType
	}
	if def != nil {
		if def.Kind == StructType {
			return StructMetaType
		}
		if def.Kind == EnumType {
			return EnumMetaType
		}
		return AliasMetaType
	}

	switch strings.ToLower(ref.Name) {
	case "map":
		return MapMetaType
	case "tuple":
		return TupleMetaType
	case "struct":
		return StructMetaType
	case "enum":
		return EnumMetaType
	case "interface":
		return InterfaceMetaType
	case "func":
		return FuncMetaType
	case "any":
		return AnyMetaType
	case "void":
		return VoidMetaType
	case "never":
		return NeverMetaType
	}
	meta := MetaTypeValue{Ref: ref}
	if meta.isPrimitive() {
		return PrimitiveMetaType
	}
	return UnknownMetaType
}

func isMetaNumber(ref TypeRef) bool {
	switch strings.ToLower(ref.Name) {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "rune", "byte":
		return true
	}
	return false
}

func ParseTypeRef(name string) TypeRef {
	if strings.HasPrefix(name, "(") &&
		strings.HasSuffix(name, ")") {
		inner := name[1 : len(name)-1]
		ref := TypeRef{Kind: TupleTypeRef}
		if inner == "" {
			return ref
		}
		for _, part := range splitTupleType(inner) {
			ref.Elems = append(
				ref.Elems,
				ParseTypeRef(strings.TrimSpace(part)),
			)
		}
		return ref
	}
	ref, rest := parseTypeRefString(name)
	if rest == "" {
		return ref
	}
	return TypeRef{Name: name}
}

func splitTupleType(input string) []string {
	var (
		parts []string
		start int
		depth int
	)
	for index, char := range input {
		switch char {
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, input[start:index])
				start = index + 1
			}
		}
	}
	return append(parts, input[start:])
}

func parseTypeRefString(name string) (TypeRef, string) {
	ref := TypeRef{}
	if strings.HasPrefix(name, "*") {
		ref.IsPtr = true
		name = strings.TrimPrefix(name, "*")
	}

	if strings.HasPrefix(name, "[]") {
		elem, rest := parseTypeRefString(strings.TrimPrefix(name, "[]"))
		ref.Kind = SliceTypeRef
		ref.Elem = &elem
		return ref, rest
	}

	if strings.HasPrefix(name, "[") {
		end := strings.IndexByte(name, ']')
		if end > 1 {
			length, err := strconv.Atoi(name[1:end])
			if err == nil {
				elem, rest := parseTypeRefString(name[end+1:])
				ref.Kind = ArrayTypeRef
				ref.Len = length
				ref.Elem = &elem
				return ref, rest
			}
		}
	}

	ref.Name = name
	if start := strings.IndexByte(name, '['); start > 0 &&
		strings.HasSuffix(name, "]") {
		ref.Name = name[:start]
		inner := name[start+1 : len(name)-1]
		for _, part := range splitTupleType(inner) {
			ref.Args = append(
				ref.Args,
				ParseTypeRef(strings.TrimSpace(part)),
			)
		}
	}
	return ref, ""
}

func (t TypeRef) IsZero() bool {
	return t.Name == "" &&
		t.Elem == nil &&
		t.Elems == nil &&
		t.Kind == NamedTypeRef
}

func (t TypeRef) IsMeta() bool {
	if t.Kind == TupleTypeRef {
		for _, elem := range t.Elems {
			if elem.IsMeta() {
				return true
			}
		}
		return false
	}
	for _, arg := range t.Args {
		if arg.IsMeta() {
			return true
		}
	}
	if t.Kind == NamedTypeRef {
		return IsMetaTypeName(t.Name)
	}
	return t.Elem != nil && t.Elem.IsMeta()
}

func (t TypeRef) IsDirectMeta() bool {
	return t.Kind == NamedTypeRef && IsMetaTypeName(t.Name)
}

func (t TypeRef) String() string {
	prefix := ""
	if t.IsPtr {
		prefix = "*"
	}

	switch t.Kind {
	case ArrayTypeRef:
		if t.Elem == nil {
			return prefix + "[" + strconv.Itoa(t.Len) + "]any"
		}
		return prefix + "[" + strconv.Itoa(t.Len) + "]" + t.Elem.String()
	case SliceTypeRef:
		if t.Elem == nil {
			return prefix + "[]any"
		}
		return prefix + "[]" + t.Elem.String()
	case TupleTypeRef:
		elems := make([]string, 0, len(t.Elems))
		for _, elem := range t.Elems {
			elems = append(elems, elem.String())
		}
		return prefix + "(" + strings.Join(elems, ",") + ")"
	default:
		name := prefix + t.Name
		if len(t.Args) == 0 {
			return name
		}
		args := make([]string, 0, len(t.Args))
		for _, arg := range t.Args {
			args = append(args, arg.String())
		}
		return name + "[" + strings.Join(args, ",") + "]"
	}
}

func (t TypeRef) GoString(ctx Ctx) string {
	prefix := ""
	if t.IsPtr {
		prefix = "*"
	}

	switch t.Kind {
	case ArrayTypeRef:
		if t.Elem == nil {
			return prefix + "[" + strconv.Itoa(t.Len) + "]any"
		}
		return prefix + "[" + strconv.Itoa(t.Len) + "]" + t.Elem.GoString(ctx)
	case SliceTypeRef:
		if t.Elem == nil {
			return prefix + "[]any"
		}
		return prefix + "[]" + t.Elem.GoString(ctx)
	case TupleTypeRef:
		fields := make([]string, 0, len(t.Elems))
		for index, elem := range t.Elems {
			fields = append(
				fields,
				"V"+strconv.Itoa(index)+" "+elem.GoString(ctx),
			)
		}
		return prefix + "struct { " +
			strings.Join(fields, "; ") + " }"
	}

	name := t.Name
	if ctx != nil {
		if _, declared := ctx.GetType(name); !declared {
			name = GoTypeName(name)
		}
	} else {
		name = GoTypeName(name)
	}

	if len(t.Args) != 0 {
		if ctx != nil {
			if def, declared := ctx.GetType(t.Name); declared &&
				len(def.TypeParams) != 0 {
				return prefix + MangleGenericName(
					t.Name,
					t.Args,
				)
			}
		}
		args := make([]string, 0, len(t.Args))
		for _, arg := range t.Args {
			args = append(args, arg.GoString(ctx))
		}
		name += "[" + strings.Join(args, ", ") + "]"
	}
	return prefix + name
}

type FieldDef struct {
	Type    TypeRef
	Default Expr
}

type EnumVariantDef struct {
	Name       string
	Fields     []TypeRef
	FieldNames []string
}

type TypeDef struct {
	Name       string
	Kind       TypeKind
	Underlying TypeRef
	Fields     map[string]FieldDef
	Variants   []EnumVariantDef
	TypeParams []TypeParam
}

type EnumValue struct {
	TypeName string
	Variant  string
	Payload  []value.Type
}

type TupleValue struct {
	Elements []value.Type
}

func EnumVariantGoName(typeName, variant string) string {
	return typeName + variant
}

func SubstituteType(
	ref TypeRef,
	params []TypeParam,
	args []TypeRef,
) TypeRef {
	substitutions := make(map[string]TypeRef, len(params))
	for index, param := range params {
		if index < len(args) {
			substitutions[param.Name] = args[index]
		}
	}
	return substituteType(ref, substitutions)
}

func substituteType(
	ref TypeRef,
	substitutions map[string]TypeRef,
) TypeRef {
	if ref.Kind == NamedTypeRef && len(ref.Args) == 0 {
		if replacement, ok := substitutions[ref.Name]; ok {
			replacement.IsPtr = ref.IsPtr || replacement.IsPtr
			return replacement
		}
	}
	if ref.Elem != nil {
		elem := substituteType(*ref.Elem, substitutions)
		ref.Elem = &elem
	}
	if len(ref.Elems) != 0 {
		elems := make([]TypeRef, len(ref.Elems))
		for index := range ref.Elems {
			elems[index] = substituteType(
				ref.Elems[index],
				substitutions,
			)
		}
		ref.Elems = elems
	}
	if len(ref.Args) != 0 {
		args := make([]TypeRef, len(ref.Args))
		for index := range ref.Args {
			args[index] = substituteType(
				ref.Args[index],
				substitutions,
			)
		}
		ref.Args = args
	}
	return ref
}

type Ctx interface {
	SetValue(key string, val value.Type)
	SetValueTypeHint(key string, val value.Type)
	GetValue(key string) (value.Type, bool)
	DeclareValue(key string, val value.Type, constant bool)
	AssignValue(key string, val value.Type) error
	SetFunc(key string, val Func)
	GetFunc(key string) (Func, bool)
	SetType(key string, def TypeDef)
	GetType(key string) (TypeDef, bool)
	SetMethod(typeName, methodName string, fn Func)
	GetMethod(typeName, methodName string) (Func, bool)
	SetGeneric(key string, params []TypeParam)
	GetGeneric(key string) ([]TypeParam, bool)
	RegisterGenericInstance(key string, args []TypeRef)
	GetGenericInstances(key string) [][]TypeRef
	GetLocalCtx() Ctx
}

func MangleGenericName(name string, args []TypeRef) string {
	if len(args) == 0 {
		return name
	}
	parts := []string{name}
	for _, arg := range args {
		parts = append(parts, mangleTypeRef(arg))
	}
	return strings.Join(parts, "_")
}

func mangleTypeRef(ref TypeRef) string {
	var name string
	switch ref.Kind {
	case ArrayTypeRef:
		name = "array" + strconv.Itoa(ref.Len)
		if ref.Elem != nil {
			name += "_" + mangleTypeRef(*ref.Elem)
		}
	case SliceTypeRef:
		name = "slice"
		if ref.Elem != nil {
			name += "_" + mangleTypeRef(*ref.Elem)
		}
	case TupleTypeRef:
		parts := []string{"tuple"}
		for _, elem := range ref.Elems {
			parts = append(parts, mangleTypeRef(elem))
		}
		name = strings.Join(parts, "_")
	default:
		name = ref.Name
		if len(ref.Args) != 0 {
			name = MangleGenericName(name, ref.Args)
		}
	}
	if ref.IsPtr {
		name = "ptr_" + name
	}
	replacer := strings.NewReplacer(
		"*", "ptr_",
		"[", "_",
		"]", "",
		",", "_",
		" ", "",
		".", "_",
	)
	return replacer.Replace(name)
}

func ValidateTypeArguments(
	ctx Ctx,
	owner string,
	params []TypeParam,
	args []TypeRef,
) error {
	if len(args) != len(params) {
		return fmt.Errorf(
			"%s expects %d type arguments, got %d",
			owner,
			len(params),
			len(args),
		)
	}
	for index, param := range params {
		if !typeSatisfiesConstraint(
			ctx,
			args[index],
			param.Constraint,
		) {
			return fmt.Errorf(
				"type argument %s does not satisfy %s for %s",
				args[index].String(),
				param.Constraint.String(),
				param.Name,
			)
		}
	}
	return nil
}

func typeSatisfiesConstraint(
	ctx Ctx,
	arg TypeRef,
	constraint TypeRef,
) bool {
	switch strings.ToLower(constraint.Name) {
	case "any":
		return true
	case "comparable":
		return isComparableTypeRef(ctx, arg)
	}
	return arg.GoString(ctx) == constraint.GoString(ctx)
}

func isComparableTypeRef(ctx Ctx, ref TypeRef) bool {
	if ref.IsPtr {
		return true
	}
	switch ref.Kind {
	case SliceTypeRef:
		return false
	case ArrayTypeRef:
		return ref.Elem != nil &&
			isComparableTypeRef(ctx, *ref.Elem)
	case TupleTypeRef:
		for _, elem := range ref.Elems {
			if !isComparableTypeRef(ctx, elem) {
				return false
			}
		}
		return true
	}
	if def, declared := ctx.GetType(ref.Name); declared {
		if def.Kind == AliasType {
			return isComparableTypeRef(ctx, def.Underlying)
		}
		for _, field := range def.Fields {
			if !isComparableTypeRef(ctx, field.Type) {
				return false
			}
		}
	}
	switch strings.ToLower(ref.Name) {
	case "map", "slice", "func":
		return false
	}
	return true
}

type Expr interface {
	Eval(Ctx) (value.Type, error)
	Type(Ctx) string
	PrintGO(Ctx) (string, error)
}

type Validatable interface {
	Validate(Ctx) error
}

type FlowKind uint8

const (
	BreakFlow FlowKind = iota + 1
	ContinueFlow
	YieldFlow
)

type FlowSignal struct {
	Kind     FlowKind
	Value    value.Type
	HasValue bool
}

func (s FlowSignal) Error() string {
	switch s.Kind {
	case BreakFlow:
		return "break"
	case ContinueFlow:
		return "continue"
	case YieldFlow:
		return "yield"
	default:
		return "control flow"
	}
}

func GoTypeName(name string) string {
	pointer := ""
	for strings.HasPrefix(name, "*") {
		pointer += "*"
		name = strings.TrimPrefix(name, "*")
	}

	switch strings.ToLower(name) {
	case "bool":
		return pointer + "bool"
	case "int":
		return pointer + "int"
	case "int8":
		return pointer + "int8"
	case "int16":
		return pointer + "int16"
	case "int32":
		return pointer + "int32"
	case "int64":
		return pointer + "int64"
	case "uint":
		return pointer + "uint"
	case "uint8":
		return pointer + "uint8"
	case "uint16":
		return pointer + "uint16"
	case "uint32":
		return pointer + "uint32"
	case "uint64":
		return pointer + "uint64"
	case "float32":
		return pointer + "float32"
	case "float64":
		return pointer + "float64"
	case "string":
		return pointer + "string"
	case "char", "rune":
		return pointer + "rune"
	case "byte":
		return pointer + "byte"
	case "any":
		return pointer + "any"
	}

	return pointer + name
}
