package map_parser

import (
	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/value"
)

type Entry struct {
	key   string
	value ast.Expr
}

type MapExpr struct {
	entries []Entry
}

func NewMapExpr(entries []Entry) *MapExpr {
	return &MapExpr{
		entries: entries,
	}
}

func (m *MapExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	out := make(map[string]any, len(m.entries))

	for _, entry := range m.entries {
		val, err := entry.value.Eval(ctx)
		if err != nil {
			return value.NewTypeNil(), err
		}

		out[entry.key] = val.Any()
	}

	return value.NewType(out), nil
}

func (*MapExpr) Type(_ ast.Ctx) string {
	return "map"
}
