package binary_parser

import (
	"math"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/token"
	"github.com/fobus89/dsl/value"
)

type BinaryExpr struct {
	Left  ast.Expr
	Op    token.TokenType
	Right ast.Expr
}

func NewBinaryExpr(op token.TokenType, left, right ast.Expr) *BinaryExpr {
	return &BinaryExpr{
		Left:  left,
		Op:    op,
		Right: right,
	}
}

func (b *BinaryExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	leftVal, err := b.Left.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	rightVal, err := b.Right.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	if leftVal.IsString() || rightVal.IsString() {
		switch b.Op {
		case token.PLUS:
			return value.NewType(leftVal.UnsafeCastString() + rightVal.UnsafeCastString()), nil
		}
	}

	if leftVal.IsNumber() && rightVal.IsNumber() {
		left := leftVal.UnsafeCastFloat64()
		right := rightVal.UnsafeCastFloat64()

		switch b.Op {
		case token.PLUS:
			return value.NewType(left + right), nil
		case token.MINUS:
			return value.NewType(left - right), nil
		case token.STAR:
			return value.NewType(left * right), nil
		case token.SLASH:
			return value.NewType(left / right), nil
		case token.PERCENT:
			return value.NewType(math.Mod(left, right)), nil
		}
	}

	return value.NewType(math.NaN()), nil
}

func (b *BinaryExpr) Type(_ ast.Ctx) string {
	return "binary"
}

func (b *BinaryExpr) PrintGO(ctx ast.Ctx) (string, error) {
	left, err := b.Left.PrintGO(ctx)
	if err != nil {
		return "", err
	}
	right, err := b.Right.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	if b.Op == token.PLUS {
		leftString := isStringExpr(ctx, b.Left)
		rightString := isStringExpr(ctx, b.Right)

		switch {
		case leftString && rightString:
			return "(" + left + " + " + right + ")", nil
		case leftString:
			converted, err := strconvExpr(ctx, b.Right)
			if err != nil {
				return "", err
			}
			return "(" + left + " + " + converted + ")", nil
		case rightString:
			converted, err := strconvExpr(ctx, b.Left)
			if err != nil {
				return "", err
			}
			return "(" + converted + " + " + right + ")", nil
		}
	}

	if b.Op == token.PERCENT {
		return "math.Mod(float64(" + left + "), float64(" + right + "))", nil
	}

	return "(" + left + " " + b.Op.String() + " " + right + ")", nil
}

func (b *BinaryExpr) ValueType(ctx ast.Ctx) string {
	if b.Op == token.PLUS &&
		(isStringExpr(ctx, b.Left) || isStringExpr(ctx, b.Right)) {
		return "string"
	}

	return "float64"
}

func isStringExpr(ctx ast.Ctx, expr ast.Expr) bool {
	if _, ok := expr.(literal_parser.String); ok {
		return true
	}

	if typed, ok := expr.(interface{ ValueType(ast.Ctx) string }); ok {
		return strings.EqualFold(typed.ValueType(ctx), "string")
	}

	if ident, ok := expr.(literal_parser.Ident); ok {
		v, found := ctx.GetValue(string(ident))
		return found && (v.IsString() ||
			strings.EqualFold(v.TypeName(), "string"))
	}

	return false
}

func strconvExpr(ctx ast.Ctx, expr ast.Expr) (string, error) {
	code, err := expr.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	switch expr.(type) {
	case literal_parser.Int:
		return "strconv.FormatInt(int64(" + code + "), 10)", nil
	case literal_parser.Float64:
		return "strconv.FormatFloat(float64(" + code + "), 'f', -1, 64)", nil
	case literal_parser.Bool:
		return "strconv.FormatBool(" + code + ")", nil
	}

	if ident, ok := expr.(literal_parser.Ident); ok {
		if v, found := ctx.GetValue(string(ident)); found {
			switch v.Typeof() {
			case "int", "int8", "int16", "int32", "int64":
				return "strconv.FormatInt(int64(" + code + "), 10)", nil
			case "uint", "uint8", "uint16", "uint32", "uint64":
				return "strconv.FormatUint(uint64(" + code + "), 10)", nil
			case "float32":
				return "strconv.FormatFloat(float64(" + code + "), 'f', -1, 32)", nil
			case "float64":
				return "strconv.FormatFloat(float64(" + code + "), 'f', -1, 64)", nil
			case "bool":
				return "strconv.FormatBool(" + code + ")", nil
			case "string":
				return code, nil
			}
		}
	}

	return "fmt.Sprint(" + code + ")", nil
}
