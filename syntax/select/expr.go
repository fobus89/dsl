package select_parser

import (
	"fmt"
	"reflect"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type Ident = literal_parser.Ident

type SelectExpr struct {
	fields []Ident
	source ast.Expr
	where  ast.Expr
	limit  ast.Expr
}

func NewSelectExpr(fields []Ident, source ast.Expr) *SelectExpr {
	return &SelectExpr{
		fields: fields,
		source: source,
	}
}

func (s *SelectExpr) Eval(ctx ast.Ctx) (value.Type, error) {

	obj, err := s.source.Eval(ctx)
	{
		if err != nil {
			return value.NewTypeNil(), err
		}
	}

	switch o := obj.Any().(type) {
	case map[string]any:
		out := map[string]any{}

		for _, v := range s.fields {

			objValue, ok := o[string(v)]
			{
				if !ok {
					return value.NewTypeNil(), fmt.Errorf("property %q not found", v)
				}
			}

			out[string(v)] = objValue
		}

		return value.NewType(out), nil
	case []any:
		out := []map[string]any{}

		for _, v := range o {
			tmp := v.(map[string]any)
			outChild := map[string]any{}

			for _, v := range s.fields {
				objValue, ok := tmp[string(v)]
				{
					if !ok {
						return value.NewTypeNil(), fmt.Errorf("property %q not found", v)
					}
				}
				outChild[string(v)] = objValue
			}

			out = append(out, outChild)
		}

		return value.NewType(out), nil
	default:
		fmt.Println(reflect.TypeOf(o))
	}

	return value.NewTypeNil(), nil
}

func (s *SelectExpr) Type(ctx ast.Ctx) string {
	return ""
}
