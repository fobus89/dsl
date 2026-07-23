package ast

import "github.com/fobus89/dsl/value"

type Func = func(...value.Type) (value.Type, error)

type TypeKind string

const (
	AliasType  TypeKind = "alias"
	StructType TypeKind = "struct"
)

type TypeDef struct {
	Name       string
	Kind       TypeKind
	Underlying string
	Fields     map[string]string
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
}

type Validatable interface {
	Validate(Ctx) error
}
