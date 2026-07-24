package ifstmt_parser

import (
	"fmt"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type IfExpr struct {
	Condition ast.Expr
	Then      []ast.Expr
	Else      []ast.Expr
}

func NewIfExpr(
	condition ast.Expr,
	thenBranch []ast.Expr,
	elseBranch []ast.Expr,
) *IfExpr {
	return &IfExpr{
		Condition: condition,
		Then:      thenBranch,
		Else:      elseBranch,
	}
}

func (i *IfExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	condition, err := i.Condition.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	branch := i.Else
	if condition.UnsafeCastBool() {
		branch = i.Then
	}

	result := value.NewTypeNil()
	for _, expr := range branch {
		result, err = expr.Eval(ctx)
		if err != nil {
			return value.NewTypeNil(), err
		}
	}

	if ref := i.ValueType(ctx); !ref.IsZero() {
		if len(branch) == 0 {
			result = value.NewType(zeroIfType(ctx, ref))
		}
		return value.NewTypeWithExplicit(
			result.Any(),
			ref.String(),
		), nil
	}
	return result, nil
}

func (*IfExpr) Type(ast.Ctx) string {
	return "if"
}

func (i *IfExpr) ValueType(ctx ast.Ctx) ast.TypeRef {
	ref, known, err := i.resultType(ctx)
	if err != nil || !known {
		return ast.TypeRef{}
	}
	return ref
}

func (i *IfExpr) ChildExprs() []ast.Expr {
	children := make([]ast.Expr, 0, len(i.Then)+len(i.Else))
	children = append(children, i.Then...)
	children = append(children, i.Else...)
	return children
}

func (i *IfExpr) MapBranchValues(
	mapper func(ast.Expr) ast.Expr,
) {
	i.Then = mapLastBranchValue(i.Then, mapper)
	i.Else = mapLastBranchValue(i.Else, mapper)
}

func mapLastBranchValue(
	branch []ast.Expr,
	mapper func(ast.Expr) ast.Expr,
) []ast.Expr {
	if len(branch) == 0 {
		return branch
	}

	last := len(branch) - 1
	if nested, ok := branch[last].(interface {
		MapBranchValues(func(ast.Expr) ast.Expr)
	}); ok {
		nested.MapBranchValues(mapper)
		return branch
	}
	branch[last] = mapper(branch[last])
	return branch
}

func (i *IfExpr) PrintGO(ctx ast.Ctx) (string, error) {
	condition, err := i.Condition.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	thenBranch, err := printBlock(ctx, i.Then)
	if err != nil {
		return "", err
	}

	var out strings.Builder
	out.WriteString("if ")
	out.WriteString(condition)
	out.WriteString(" {\n")
	out.WriteString(thenBranch)
	out.WriteString("\n}")

	if len(i.Else) != 0 {
		if nested, ok := i.Else[0].(*IfExpr); ok && len(i.Else) == 1 {
			printed, err := nested.PrintGO(ctx)
			if err != nil {
				return "", err
			}
			out.WriteString(" else ")
			out.WriteString(printed)
		} else {
			elseBranch, err := printBlock(ctx, i.Else)
			if err != nil {
				return "", err
			}
			out.WriteString(" else {\n")
			out.WriteString(elseBranch)
			out.WriteString("\n}")
		}
	}

	return out.String(), nil
}

func (i *IfExpr) PrintGOAssign(
	ctx ast.Ctx,
	name string,
) (string, error) {
	printed, err := i.printGOValue(ctx)
	if err != nil {
		return "", err
	}
	return name + " := " + printed, nil
}

func (i *IfExpr) printGOValue(ctx ast.Ctx) (string, error) {
	resultType := "any"
	resultKnown := false
	if ref, known, err := i.resultType(ctx); err != nil {
		return "", err
	} else if known {
		resultType = ref.GoString(ctx)
		resultKnown = true
	}

	condition, err := i.Condition.PrintGO(ctx)
	if err != nil {
		return "", err
	}
	thenBranch, err := printValueBlock(
		ctx,
		i.Then,
		resultType,
		resultKnown,
	)
	if err != nil {
		return "", err
	}
	elseBranch, err := printValueBlock(
		ctx,
		i.Else,
		resultType,
		resultKnown,
	)
	if err != nil {
		return "", err
	}

	return "func() " + resultType + " {\n\tif " + condition + " {\n" +
		indent(thenBranch) + "\n\t}\n" +
		indent(elseBranch) + "\n}()", nil
}

func (i *IfExpr) resultType(
	ctx ast.Ctx,
) (ast.TypeRef, bool, error) {
	thenType, thenKnown := blockValueType(ctx, i.Then)
	elseType, elseKnown := blockValueType(ctx, i.Else)

	if len(i.Else) == 0 && thenKnown {
		return thenType, true, nil
	}
	if len(i.Then) == 0 && elseKnown {
		return elseType, true, nil
	}
	if !thenKnown || !elseKnown {
		return ast.TypeRef{}, false, nil
	}
	if thenType.GoString(ctx) != elseType.GoString(ctx) {
		return ast.TypeRef{}, false, fmt.Errorf(
			"if branches return incompatible types %s and %s",
			thenType.String(),
			elseType.String(),
		)
	}
	return thenType, true, nil
}

func blockValueType(
	ctx ast.Ctx,
	exprs []ast.Expr,
) (ast.TypeRef, bool) {
	if len(exprs) == 0 {
		return ast.TypeRef{}, false
	}
	return ifExprType(ctx, exprs[len(exprs)-1])
}

func ifExprType(
	ctx ast.Ctx,
	expr ast.Expr,
) (ast.TypeRef, bool) {
	if typed, ok := expr.(interface {
		ValueType(ast.Ctx) ast.TypeRef
	}); ok {
		ref := typed.ValueType(ctx)
		if !ref.IsZero() {
			return ref, true
		}
	}

	switch expr.(type) {
	case literal_parser.Int:
		return ast.TypeRef{Name: "int"}, true
	case literal_parser.Float64:
		return ast.TypeRef{Name: "float64"}, true
	case literal_parser.String:
		return ast.TypeRef{Name: "string"}, true
	case literal_parser.Bool:
		return ast.TypeRef{Name: "bool"}, true
	}
	return ast.TypeRef{}, false
}

func printBlock(ctx ast.Ctx, exprs []ast.Expr) (string, error) {
	lines := make([]string, 0, len(exprs))
	for _, expr := range exprs {
		printed, err := expr.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		lines = append(lines, indent(printed))
	}
	return strings.Join(lines, "\n"), nil
}

func indent(code string) string {
	return "\t" + strings.ReplaceAll(code, "\n", "\n\t")
}

func printValueBlock(
	ctx ast.Ctx,
	exprs []ast.Expr,
	resultType string,
	resultKnown bool,
) (string, error) {
	if len(exprs) == 0 {
		if resultKnown {
			return "var zero " + resultType + "\nreturn zero", nil
		}
		return "return nil", nil
	}

	lines := make([]string, 0, len(exprs))
	for index, expr := range exprs {
		if index == len(exprs)-1 {
			if nested, ok := expr.(*IfExpr); ok {
				printed, err := nested.printGOValue(ctx)
				if err != nil {
					return "", err
				}
				lines = append(lines, "return "+printed)
				continue
			}

			printed, err := expr.PrintGO(ctx)
			if err != nil {
				return "", err
			}
			lines = append(lines, "return "+printed)
			continue
		}

		printed, err := expr.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		lines = append(lines, printed)
	}

	return strings.Join(lines, "\n"), nil
}

func zeroIfType(ctx ast.Ctx, ref ast.TypeRef) any {
	if ref.IsPtr || ref.Kind == ast.SliceTypeRef {
		return nil
	}
	if ref.Kind == ast.ArrayTypeRef {
		result := make([]any, ref.Len)
		if ref.Elem != nil {
			for index := range result {
				result[index] = zeroIfType(ctx, *ref.Elem)
			}
		}
		return result
	}
	if def, declared := ctx.GetType(ref.Name); declared {
		if def.Kind == ast.AliasType {
			return zeroIfType(ctx, def.Underlying)
		}
		fields := make(map[string]any, len(def.Fields))
		for name, field := range def.Fields {
			fields[name] = zeroIfType(ctx, field.Type)
		}
		return fields
	}
	switch strings.ToLower(ref.Name) {
	case "string":
		return ""
	case "bool":
		return false
	case "float32", "float64":
		return float64(0)
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"rune", "byte":
		return int64(0)
	}
	return nil
}
