package ifstmt_parser

import (
	"strings"

	"github.com/fobus89/dsl/ast"
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

	return result, nil
}

func (*IfExpr) Type(ast.Ctx) string {
	return "if"
}

func (i *IfExpr) ChildExprs() []ast.Expr {
	children := make([]ast.Expr, 0, len(i.Then)+len(i.Else))
	children = append(children, i.Then...)
	children = append(children, i.Else...)
	return children
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
	condition, err := i.Condition.PrintGO(ctx)
	if err != nil {
		return "", err
	}
	thenBranch, err := printValueBlock(ctx, i.Then)
	if err != nil {
		return "", err
	}
	elseBranch, err := printValueBlock(ctx, i.Else)
	if err != nil {
		return "", err
	}

	return "func() any {\n\tif " + condition + " {\n" +
		indent(thenBranch) + "\n\t}\n" +
		indent(elseBranch) + "\n}()", nil
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

func printValueBlock(ctx ast.Ctx, exprs []ast.Expr) (string, error) {
	if len(exprs) == 0 {
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
