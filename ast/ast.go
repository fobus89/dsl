package ast

import (
	"strconv"
	"strings"

	"github.com/fobus89/dsl/value"
)

type Func = func(...value.Type) (value.Type, error)

type TypeKind string

const (
	AliasType  TypeKind = "alias"
	StructType TypeKind = "struct"
)

type TypeRefKind uint8

const (
	NamedTypeRef TypeRefKind = iota
	ArrayTypeRef
	SliceTypeRef
)

type TypeRef struct {
	Name  string
	IsPtr bool
	Kind  TypeRefKind
	Elem  *TypeRef
	Len   int
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
	}
	if def != nil {
		if def.Kind == StructType {
			return StructMetaType
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
	ref, rest := parseTypeRefString(name)
	if rest == "" {
		return ref
	}
	return TypeRef{Name: name}
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
	return ref, ""
}

func (t TypeRef) IsZero() bool {
	return t.Name == "" && t.Elem == nil
}

func (t TypeRef) IsMeta() bool {
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
	default:
		return prefix + t.Name
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
	}

	name := t.Name
	if ctx != nil {
		if _, declared := ctx.GetType(name); !declared {
			name = GoTypeName(name)
		}
	} else {
		name = GoTypeName(name)
	}

	return prefix + name
}

type FieldDef struct {
	Type    TypeRef
	Default Expr
}

type TypeDef struct {
	Name       string
	Kind       TypeKind
	Underlying TypeRef
	Fields     map[string]FieldDef
}

type Ctx interface {
	SetValue(key string, val value.Type)
	GetValue(key string) (value.Type, bool)
	SetFunc(key string, val Func)
	GetFunc(key string) (Func, bool)
	SetType(key string, def TypeDef)
	GetType(key string) (TypeDef, bool)
	SetMethod(typeName, methodName string, fn Func)
	GetMethod(typeName, methodName string) (Func, bool)
	GetLocalCtx() Ctx
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
