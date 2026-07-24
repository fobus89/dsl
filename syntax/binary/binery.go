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

func (b *BinaryExpr) Parts() (ast.Expr, token.TokenType, ast.Expr) {
	return b.Left, b.Op, b.Right
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
		left = valueExpr(ctx, b.Left, left)
		right = valueExpr(ctx, b.Right, right)

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

func (b *BinaryExpr) ValueType(ctx ast.Ctx) ast.TypeRef {
	if b.Op == token.PLUS &&
		(isStringExpr(ctx, b.Left) || isStringExpr(ctx, b.Right)) {
		return ast.TypeRef{Name: "string"}
	}

	return ast.TypeRef{Name: "float64"}
}

func isStringExpr(ctx ast.Ctx, expr ast.Expr) bool {
	ref, ok := exprType(ctx, expr)
	return ok && strings.EqualFold(ref.Name, "string")
}

func strconvExpr(ctx ast.Ctx, expr ast.Expr) (string, error) {
	code, err := expr.PrintGO(ctx)
	if err != nil {
		return "", err
	}
	code = valueExpr(ctx, expr, code)

	switch expr.(type) {
	case literal_parser.Int:
		return "strconv.FormatInt(int64(" + code + "), 10)", nil
	case literal_parser.Float64:
		return "strconv.FormatFloat(float64(" + code + "), 'f', -1, 64)", nil
	case literal_parser.Bool:
		return "strconv.FormatBool(" + code + ")", nil
	}

	if ref, ok := exprType(ctx, expr); ok {
		switch strings.ToLower(ref.Name) {
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

	return "fmt.Sprint(" + code + ")", nil
}

func valueExpr(ctx ast.Ctx, expr ast.Expr, code string) string {
	ref, ok := exprType(ctx, expr)
	if ok && ref.IsPtr {
		return "(*" + code + ")"
	}
	return code
}

func exprType(ctx ast.Ctx, expr ast.Expr) (ast.TypeRef, bool) {
	switch expr.(type) {
	case literal_parser.String:
		return ast.TypeRef{Name: "string"}, true
	case literal_parser.Int:
		return ast.TypeRef{Name: "int64"}, true
	case literal_parser.Float64:
		return ast.TypeRef{Name: "float64"}, true
	case literal_parser.Bool:
		return ast.TypeRef{Name: "bool"}, true
	}

	if typed, ok := expr.(interface {
		ValueType(ast.Ctx) ast.TypeRef
	}); ok {
		ref := typed.ValueType(ctx)
		return ref, !ref.IsZero()
	}

	if ident, ok := expr.(literal_parser.Ident); ok {
		if v, found := ctx.GetValue(string(ident)); found {
			return ast.ParseTypeRef(v.TypeName()), true
		}
	}

	return ast.TypeRef{}, false
}
