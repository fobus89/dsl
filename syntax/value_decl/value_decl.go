// Package valuedecl_parser exposes value declarations. Both let and const
// declarations are parsed into ValueDecl nodes.
package valuedecl_parser

import (
	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	let_parser "github.com/fobus89/dsl/syntax/let"
)

type Ident = let_parser.Ident
type ValueDecl = let_parser.ValueDecl

func RegisterParser(p parser.Parser) {
	let_parser.RegisterParser(p)
}

func NewValueDecl(
	name Ident,
	value ast.Expr,
	constant bool,
) *ValueDecl {
	return let_parser.NewValueDecl(name, value, constant)
}
