package ast

import "github.com/fobus89/dsl/value"

type Expr interface {
	Eval() (value.Type, error)
	Type() string
}
