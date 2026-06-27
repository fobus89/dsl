package ast

import "github.com/fobus89/dsl/value"

type functype = func(...value.Type) (value.Type, error)

type Ctx interface {
	SetValue(key string, val value.Type)
	GetValue(key string) (value.Type, bool)
	SetFunc(key string, val functype)
	GetFunc(key string) (functype, bool)
	GetLocalCtx() Ctx
}

type Expr interface {
	Eval(Ctx) (value.Type, error)
	Type(Ctx) string
}
