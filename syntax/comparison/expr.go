package comparison_parser

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/token"
	"github.com/fobus89/dsl/value"
)

type ComparisonExpr struct {
	left  ast.Expr
	op    token.TokenType
	right ast.Expr
}

func NewComparisonExpr(op token.TokenType, left, right ast.Expr) *ComparisonExpr {
	return &ComparisonExpr{
		left:  left,
		op:    op,
		right: right,
	}
}

func (c *ComparisonExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	if err := c.validateTypes(ctx); err != nil {
		return value.NewTypeNil(), err
	}

	leftVal, err := c.left.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	rightVal, err := c.right.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	switch c.op {
	case token.EQ_EQ:
		return value.NewType(equalValues(leftVal.Any(), rightVal.Any())), nil
	case token.BANG_EQ:
		return value.NewType(!equalValues(leftVal.Any(), rightVal.Any())), nil
	}

	if leftVal.IsNumber() && rightVal.IsNumber() {
		left := leftVal.UnsafeCastFloat64()
		right := rightVal.UnsafeCastFloat64()

		switch c.op {
		case token.GT:
			return value.NewType(left > right), nil
		case token.LT:
			return value.NewType(left < right), nil
		case token.GT_EQ:
			return value.NewType(left >= right), nil
		case token.LT_EQ:
			return value.NewType(left <= right), nil
		}
	}

	return value.NewTypeNil(), fmt.Errorf("operator %q is not supported for %s and %s", c.op, leftVal.Typeof(), rightVal.Typeof())
}

func (*ComparisonExpr) Type(_ ast.Ctx) string {
	return "comparison"
}

func (*ComparisonExpr) ValueType(ast.Ctx) ast.TypeRef {
	return ast.TypeRef{Name: "bool"}
}

func (c *ComparisonExpr) Parts() (ast.Expr, token.TokenType, ast.Expr) {
	return c.left, c.op, c.right
}

func (c *ComparisonExpr) PrintGO(ctx ast.Ctx) (string, error) {
	if err := c.validateTypes(ctx); err != nil {
		return "", err
	}

	left, err := c.left.PrintGO(ctx)
	if err != nil {
		return "", err
	}
	right, err := c.right.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	return "(" + left + " " + c.op.String() + " " + right + ")", nil
}

type comparisonOperandType struct {
	ref     ast.TypeRef
	known   bool
	untyped bool
}

func (c *ComparisonExpr) validateTypes(ctx ast.Ctx) error {
	left := comparisonExprType(ctx, c.left)
	right := comparisonExprType(ctx, c.right)
	if !left.known || !right.known {
		return nil
	}

	if isNilType(left.ref) || isNilType(right.ref) {
		other := right.ref
		if isNilType(right.ref) {
			other = left.ref
		}
		if c.op != token.EQ_EQ && c.op != token.BANG_EQ {
			return comparisonTypeError(c.op, left.ref, right.ref)
		}
		if isNilType(left.ref) && isNilType(right.ref) {
			return nil
		}
		if !isNilableType(ctx, other) {
			return comparisonTypeError(c.op, left.ref, right.ref)
		}
		return nil
	}

	if !comparisonTypesCompatible(ctx, left, right) {
		return comparisonTypeError(c.op, left.ref, right.ref)
	}

	switch c.op {
	case token.EQ_EQ, token.BANG_EQ:
		if !isComparableType(ctx, left.ref) ||
			!isComparableType(ctx, right.ref) {
			return fmt.Errorf(
				"operator %s requires comparable types, got %s and %s",
				c.op,
				left.ref.String(),
				right.ref.String(),
			)
		}
	default:
		if !isOrderedType(ctx, left.ref) ||
			!isOrderedType(ctx, right.ref) {
			return fmt.Errorf(
				"operator %s requires numbers or strings, got %s and %s",
				c.op,
				left.ref.String(),
				right.ref.String(),
			)
		}
	}
	return nil
}

func comparisonExprType(
	ctx ast.Ctx,
	expr ast.Expr,
) comparisonOperandType {
	switch expr.(type) {
	case literal_parser.Int:
		return comparisonOperandType{
			ref:     ast.TypeRef{Name: "int"},
			known:   true,
			untyped: true,
		}
	case literal_parser.Float64:
		return comparisonOperandType{
			ref:     ast.TypeRef{Name: "float64"},
			known:   true,
			untyped: true,
		}
	case literal_parser.String:
		return comparisonOperandType{
			ref:     ast.TypeRef{Name: "string"},
			known:   true,
			untyped: true,
		}
	case literal_parser.Bool:
		return comparisonOperandType{
			ref:     ast.TypeRef{Name: "bool"},
			known:   true,
			untyped: true,
		}
	case literal_parser.Nil:
		return comparisonOperandType{
			ref:     ast.TypeRef{Name: "nil"},
			known:   true,
			untyped: true,
		}
	}

	if typed, ok := expr.(interface {
		ValueType(ast.Ctx) ast.TypeRef
	}); ok {
		ref := typed.ValueType(ctx)
		if !ref.IsZero() {
			return comparisonOperandType{ref: ref, known: true}
		}
	}
	return comparisonOperandType{}
}

func comparisonTypesCompatible(
	ctx ast.Ctx,
	left comparisonOperandType,
	right comparisonOperandType,
) bool {
	if left.ref.GoString(ctx) == right.ref.GoString(ctx) {
		return true
	}
	if !left.untyped && !right.untyped {
		return false
	}

	leftUnderlying := comparisonUnderlying(ctx, left.ref)
	rightUnderlying := comparisonUnderlying(ctx, right.ref)
	return isNumericRef(leftUnderlying) &&
		isNumericRef(rightUnderlying) ||
		strings.EqualFold(leftUnderlying.Name, "string") &&
			strings.EqualFold(rightUnderlying.Name, "string") ||
		strings.EqualFold(leftUnderlying.Name, "bool") &&
			strings.EqualFold(rightUnderlying.Name, "bool")
}

func comparisonUnderlying(
	ctx ast.Ctx,
	ref ast.TypeRef,
) ast.TypeRef {
	if ref.Kind != ast.NamedTypeRef {
		return ref
	}
	if def, declared := ctx.GetType(ref.Name); declared &&
		def.Kind == ast.AliasType {
		underlying := comparisonUnderlying(ctx, def.Underlying)
		underlying.IsPtr = ref.IsPtr || underlying.IsPtr
		return underlying
	}
	return ref
}

func isNilType(ref ast.TypeRef) bool {
	return ref.Kind == ast.NamedTypeRef &&
		strings.EqualFold(ref.Name, "nil")
}

func isNilableType(ctx ast.Ctx, ref ast.TypeRef) bool {
	if ref.IsPtr || ref.Kind == ast.SliceTypeRef {
		return true
	}
	ref = comparisonUnderlying(ctx, ref)
	switch strings.ToLower(ref.Name) {
	case "any", "interface", "map", "func":
		return true
	}
	return false
}

func isComparableType(ctx ast.Ctx, ref ast.TypeRef) bool {
	if ref.IsPtr {
		return true
	}
	switch ref.Kind {
	case ast.SliceTypeRef:
		return false
	case ast.ArrayTypeRef:
		return ref.Elem != nil && isComparableType(ctx, *ref.Elem)
	case ast.TupleTypeRef:
		for _, elem := range ref.Elems {
			if !isComparableType(ctx, elem) {
				return false
			}
		}
		return true
	}

	if def, declared := ctx.GetType(ref.Name); declared {
		if def.Kind == ast.AliasType {
			return isComparableType(ctx, def.Underlying)
		}
		for _, field := range def.Fields {
			if !isComparableType(ctx, field.Type) {
				return false
			}
		}
		return true
	}

	switch strings.ToLower(ref.Name) {
	case "slice", "map", "func":
		return false
	}
	return true
}

func isOrderedType(ctx ast.Ctx, ref ast.TypeRef) bool {
	if ref.IsPtr || ref.Kind != ast.NamedTypeRef {
		return false
	}
	ref = comparisonUnderlying(ctx, ref)
	return isNumericRef(ref) ||
		strings.EqualFold(ref.Name, "string")
}

func isNumericRef(ref ast.TypeRef) bool {
	if ref.IsPtr || ref.Kind != ast.NamedTypeRef {
		return false
	}
	switch strings.ToLower(ref.Name) {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "rune", "byte":
		return true
	}
	return false
}

func comparisonTypeError(
	op token.TokenType,
	left ast.TypeRef,
	right ast.TypeRef,
) error {
	return fmt.Errorf(
		"operator %s cannot compare %s with %s",
		op,
		left.String(),
		right.String(),
	)
}

func equalValues(left, right any) bool {
	leftNumber, leftNumberOK := castNumber(left)
	rightNumber, rightNumberOK := castNumber(right)
	if leftNumberOK && rightNumberOK {
		if math.IsNaN(leftNumber) && math.IsNaN(rightNumber) {
			return true
		}

		return leftNumber == rightNumber
	}

	leftFloat, leftIsFloat := left.(float64)
	rightFloat, rightIsFloat := right.(float64)
	if leftIsFloat && rightIsFloat && math.IsNaN(leftFloat) && math.IsNaN(rightFloat) {
		return true
	}

	return reflect.DeepEqual(left, right)
}

func castNumber(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	}

	return 0, false
}
