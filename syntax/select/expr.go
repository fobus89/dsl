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
	fields [][2]Ident
	source ast.Expr
	where  ast.Expr
	limit  ast.Expr
}

func NewSelectExpr(fields [][2]Ident, source ast.Expr, where, limit ast.Expr) *SelectExpr {

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

		var _limit = -1

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

	if s.where != nil {
		localCtx := ctx.GetLocalCtx()

		for k, v := range row {
			localCtx.SetValue(k, value.NewType(v))
		}

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

		val, ok := row[string(field[0])]
		if !ok {
			return nil, false,
				fmt.Errorf("property %q not found", field)
		}

		out[string(field[1])] = val
	}

	return out, true, nil
}

func projectPrimitiveRow[T any](
	ctx ast.Ctx,
	rows []T,
	where ast.Expr,
	limit ast.Expr,
) ([]T, bool, error) {

	var _limit = -1

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
