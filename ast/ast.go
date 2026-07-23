package ast

import (
	"strings"

	"github.com/fobus89/dsl/value"
)

type Func = func(...value.Type) (value.Type, error)

type TypeKind string

const (
	AliasType  TypeKind = "alias"
	StructType TypeKind = "struct"
)

type TypeRef struct {
	Name  string
	IsPtr bool
}

func ParseTypeRef(name string) TypeRef {
	if strings.HasPrefix(name, "*") {
		return TypeRef{
			Name:  strings.TrimPrefix(name, "*"),
			IsPtr: true,
		}
	}
	return TypeRef{Name: name}
}

func (t TypeRef) String() string {
	if t.IsPtr {
		return "*" + t.Name
	}
	return t.Name
}

func (t TypeRef) GoString() string {
	if t.IsPtr {
		return "*" + GoTypeName(t.Name)
	}
	return GoTypeName(t.Name)
}

type TypeDef struct {
	Name       string
	Kind       TypeKind
	Underlying TypeRef
	Fields     map[string]TypeRef
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
