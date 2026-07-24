package select_parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/token"
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

func (StarExpr) PrintGO(ast.Ctx) (string, error) {
	return "Star()", nil
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

func (s *SelectExpr) PrintGO(ctx ast.Ctx) (string, error) {
	source, err := s.source.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	where := "true"
	if s.where != nil {
		where, err = selectRowExprGO(ctx, s.where)
		if err != nil {
			return "", err
		}
		where = "_truthy(" + where + ")"
	}

	limit := "-1"
	if s.limit != nil {
		printed, err := s.limit.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		limit = "int(_number(any(" + printed + ")))"
	}

	var projection strings.Builder
	for _, field := range s.fields {
		if _, ok := field[0].(StarExpr); ok {
			projection.WriteString(
				"\t\tfor _key, _value := range _row { _out[_key] = _value }\n",
			)
			continue
		}

		printed, err := selectRowExprGO(ctx, field[0])
		if err != nil {
			return "", err
		}
		fmt.Fprintf(
			&projection,
			"\t\t_out[%s] = %s\n",
			strconv.Quote(fieldName(ctx, field[1])),
			printed,
		)
	}

	code := `func() any {
	_source := any(` + source + `)
	_isNumber := func(_value any) bool {
		switch _value.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			return true
		}
		return false
	}
	_number := func(_value any) float64 {
		switch _value := _value.(type) {
		case int: return float64(_value)
		case int8: return float64(_value)
		case int16: return float64(_value)
		case int32: return float64(_value)
		case int64: return float64(_value)
		case uint: return float64(_value)
		case uint8: return float64(_value)
		case uint16: return float64(_value)
		case uint32: return float64(_value)
		case uint64: return float64(_value)
		case float32: return float64(_value)
		case float64: return _value
		}
		return math.NaN()
	}
	_truthy := func(_value any) bool {
		switch _value := _value.(type) {
		case nil: return false
		case bool: return _value
		case int: return _value != 0
		case int8: return _value != 0
		case int16: return _value != 0
		case int32: return _value != 0
		case int64: return _value != 0
		case uint: return _value != 0
		case uint8: return _value != 0
		case uint16: return _value != 0
		case uint32: return _value != 0
		case uint64: return _value != 0
		case float32: return _value != 0
		case float64: return _value != 0
		}
		return true
	}
	_string := func(_value any) string { return fmt.Sprint(_value) }
	_equal := func(_left, _right any) bool {
		if _isNumber(_left) && _isNumber(_right) {
			return _number(_left) == _number(_right)
		}
		return reflect.DeepEqual(_left, _right)
	}
	_get := func(_row map[string]any, _path ...string) any {
		var _value any = _row
		for _, _key := range _path {
			_object, _ok := _value.(map[string]any)
			if !_ok { return nil }
			_value, _ok = _object[_key]
			if !_ok { return nil }
		}
		return _value
	}
	_add := func(_left, _right any) any {
		if _, _ok := _left.(string); _ok {
			return _string(_left) + _string(_right)
		}
		if _, _ok := _right.(string); _ok {
			return _string(_left) + _string(_right)
		}
		return _number(_left) + _number(_right)
	}
	_ = _equal
	_ = _add
	_limit := ` + limit + `
	_match := func(_row map[string]any) bool {
		return ` + where + `
	}
	_project := func(_row map[string]any) (map[string]any, bool) {
		if !_match(_row) { return nil, false }
		_out := make(map[string]any)
` + projection.String() + `		return _out, true
	}

	switch _rows := _source.(type) {
	case map[string]any:
		_out, _ok := _project(_rows)
		if !_ok { return nil }
		return _out
	case []map[string]any:
		_out := make([]map[string]any, 0, len(_rows))
		for _, _row := range _rows {
			if _limit >= 0 && len(_out) >= _limit { break }
			if _projected, _ok := _project(_row); _ok {
				_out = append(_out, _projected)
			}
		}
		return _out
	case []any:
		_out := make([]map[string]any, 0, len(_rows))
		for _, _value := range _rows {
			if _limit >= 0 && len(_out) >= _limit { break }
			_row, _ok := _value.(map[string]any)
			if !_ok { continue }
			if _projected, _ok := _project(_row); _ok {
				_out = append(_out, _projected)
			}
		}
		return _out
	}

	_reflectRows := reflect.ValueOf(_source)
	if _reflectRows.IsValid() && _reflectRows.Kind() == reflect.Slice {
		_out := reflect.MakeSlice(_reflectRows.Type(), 0, _reflectRows.Len())
		for _index := 0; _index < _reflectRows.Len(); _index++ {
			if _limit >= 0 && _out.Len() >= _limit { break }
			_value := _reflectRows.Index(_index)
			if _match(map[string]any{"value": _value.Interface()}) {
				_out = reflect.Append(_out, _value)
			}
		}
		return _out.Interface()
	}
	return nil
}()`

	return code, nil
}

func selectRowExprGO(ctx ast.Ctx, expr ast.Expr) (string, error) {
	if ident, ok := expr.(Ident); ok {
		return "_get(_row, " + strconv.Quote(string(ident)) + ")", nil
	}

	if member, ok := expr.(interface{ Path() ([]string, bool) }); ok {
		if path, valid := member.Path(); valid {
			quoted := make([]string, 0, len(path))
			for _, part := range path {
				quoted = append(quoted, strconv.Quote(part))
			}
			return "_get(_row, " + strings.Join(quoted, ", ") + ")", nil
		}
	}

	if parts, ok := expr.(interface {
		Parts() (ast.Expr, token.TokenType, ast.Expr)
	}); ok {
		left, op, right := parts.Parts()
		leftGO, err := selectRowExprGO(ctx, left)
		if err != nil {
			return "", err
		}
		rightGO, err := selectRowExprGO(ctx, right)
		if err != nil {
			return "", err
		}

		switch expr.Type(ctx) {
		case "binary":
			switch op {
			case token.PLUS:
				return "_add(" + leftGO + ", " + rightGO + ")", nil
			case token.MINUS:
				return "(_number(" + leftGO + ") - _number(" + rightGO + "))", nil
			case token.STAR:
				return "(_number(" + leftGO + ") * _number(" + rightGO + "))", nil
			case token.SLASH:
				return "(_number(" + leftGO + ") / _number(" + rightGO + "))", nil
			case token.PERCENT:
				return "math.Mod(_number(" + leftGO + "), _number(" + rightGO + "))", nil
			}
		case "comparison":
			switch op {
			case token.EQ_EQ:
				return "_equal(" + leftGO + ", " + rightGO + ")", nil
			case token.BANG_EQ:
				return "!_equal(" + leftGO + ", " + rightGO + ")", nil
			case token.GT, token.LT, token.GT_EQ, token.LT_EQ:
				return "(_number(" + leftGO + ") " + op.String() +
					" _number(" + rightGO + "))", nil
			}
		case "logical":
			logicalOp := op.String()
			if op == token.AND {
				logicalOp = "&&"
			} else if op == token.OR {
				logicalOp = "||"
			}
			return "(_truthy(" + leftGO + ") " + logicalOp +
				" _truthy(" + rightGO + "))", nil
		}
	}

	if parts, ok := expr.(interface {
		Parts() (token.TokenType, int, ast.Expr)
	}); ok {
		op, count, operand := parts.Parts()
		printed, err := selectRowExprGO(ctx, operand)
		if err != nil {
			return "", err
		}

		switch op {
		case token.BANG:
			if count%2 == 0 {
				return "_truthy(" + printed + ")", nil
			}
			return "!_truthy(" + printed + ")", nil
		case token.MINUS:
			return "-_number(" + printed + ")", nil
		case token.PLUS:
			return "_number(" + printed + ")", nil
		}
	}

	if call, ok := expr.(interface {
		Parts() (ast.Expr, []ast.Expr)
	}); ok {
		callee, args := call.Parts()
		calleeGO, err := callee.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		printedArgs := make([]string, 0, len(args))
		for _, arg := range args {
			printed, err := selectRowExprGO(ctx, arg)
			if err != nil {
				return "", err
			}
			printedArgs = append(printedArgs, printed)
		}
		return calleeGO + "(" + strings.Join(printedArgs, ", ") + ")", nil
	}

	return expr.PrintGO(ctx)
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
