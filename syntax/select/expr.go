package select_parser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type Ident = literal_parser.Ident

type StarExpr struct{}

func NewStarExpr() StarExpr {
	return StarExpr{}
}

func (StarExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	return value.NewTypeNil(), nil
}

func (StarExpr) Type(ctx ast.Ctx) string {
	return "star"
}

type SelectExpr struct {
	fields [][2]ast.Expr
	source ast.Expr
	where  ast.Expr
	limit  ast.Expr
}

func NewSelectExpr(fields [][2]ast.Expr, source ast.Expr, where, limit ast.Expr) *SelectExpr {
	return &SelectExpr{
		fields: fields,
		source: source,
		where:  where,
		limit:  limit,
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

	case []int:
		outChild, ok, err := projectPrimitiveRow(ctx, o, s.where, s.limit)
		{
			if err != nil {
				return value.NewTypeNil(), err
			}
		}

		if !ok {
			return value.NewTypeNil(), nil
		}

		return value.NewType(outChild), nil

	case []string:
		outChild, ok, err := projectPrimitiveRow(ctx, o, s.where, s.limit)
		{
			if err != nil {
				return value.NewTypeNil(), err
			}
		}

		if !ok {
			return value.NewTypeNil(), nil
		}

		return value.NewType(outChild), nil

	case []int64:
		outChild, ok, err := projectPrimitiveRow(ctx, o, s.where, s.limit)
		{
			if err != nil {
				return value.NewTypeNil(), err
			}
		}

		if !ok {
			return value.NewTypeNil(), nil
		}

		return value.NewType(outChild), nil

	case []float64:
		outChild, ok, err := projectPrimitiveRow(ctx, o, s.where, s.limit)
		{
			if err != nil {
				return value.NewTypeNil(), err
			}
		}

		if !ok {
			return value.NewTypeNil(), nil
		}

		return value.NewType(outChild), nil

	case map[string]any:
		outChild, ok, err := s.projectRow(ctx, o)
		{
			if err != nil {
				return value.NewTypeNil(), err
			}
		}

		if !ok {
			return value.NewTypeNil(), nil
		}

		return value.NewType(outChild), nil

	case []any:
		out := []map[string]any{}

		_limit := -1

		if s.limit != nil {
			v, err := s.limit.Eval(ctx)
			if err != nil {
				return value.NewTypeNil(), err
			}

			if !v.IsNumber() {
				return value.NewTypeNil(), fmt.Errorf("limit invalid type %s", v.Typeof())
			}

			_limit = v.UnsafeCastInt()
		}

		for _, row := range o {

			if _limit != -1 && len(out) >= _limit {
				break
			}

			tmp, ok := row.(map[string]any)
			{
				if !ok {
					return value.NewTypeNil(), fmt.Errorf(
						"select: expected map[string]any, got %T",
						row,
					)
				}
			}

			outChild, ok, err := s.projectRow(ctx, tmp)
			{
				if err != nil {
					return value.NewTypeNil(), err
				}
			}
			if ok {
				out = append(out, outChild)
			}
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

func (s *SelectExpr) projectRow(
	ctx ast.Ctx,
	row map[string]any,
) (map[string]any, bool, error) {
	localCtx := ctx.GetLocalCtx()

	for k, v := range row {
		localCtx.SetValue(k, value.NewType(v))
	}

	if s.where != nil {
		cond, err := s.where.Eval(localCtx)
		if err != nil {
			return nil, false, err
		}

		if !cond.UnsafeCastBool() {
			return nil, false, nil
		}
	}

	out := make(map[string]any, len(s.fields))

	for _, field := range s.fields {
		if _, ok := field[0].(StarExpr); ok {
			for k, v := range row {
				out[k] = v
			}
			continue
		}

		val, err := field[0].Eval(localCtx)
		if err != nil {
			out[fieldName(ctx, field[1])] = nil
			continue
		}

		out[fieldName(ctx, field[1])] = val.Any()
	}

	return out, true, nil
}

func fieldName(ctx ast.Ctx, expr ast.Expr) string {
	if ident, ok := expr.(Ident); ok {
		return string(ident)
	}

	if name, ok := expr.(fmt.Stringer); ok {
		parts := strings.Split(name.String(), ".")
		return parts[len(parts)-1]
	}

	return expr.Type(ctx)
}

func projectPrimitiveRow[T any](
	ctx ast.Ctx,
	rows []T,
	where ast.Expr,
	limit ast.Expr,
) ([]T, bool, error) {
	_limit := -1

	if limit != nil {
		v, err := limit.Eval(ctx)
		if err != nil {
			return nil, false, err
		}

		if !v.IsNumber() {
			return nil, false, fmt.Errorf("limit invalid type %s", v.Typeof())
		}

		_limit = v.UnsafeCastInt()
	}

	localCtx := ctx.GetLocalCtx()

	var out []T

	for _, row := range rows {
		localCtx.SetValue("value", value.NewType(row))

		if where != nil {

			cond, err := where.Eval(localCtx)
			{
				if err != nil {
					return nil, false, err
				}
			}

			if !cond.UnsafeCastBool() {
				continue
			}
		}

		if _limit != -1 && len(out) >= _limit {
			break
		}

		out = append(out, row)
	}

	return out, true, nil
}
