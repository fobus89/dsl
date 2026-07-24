package matchstmt_parser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type MatchArm struct {
	Pattern      ast.Expr
	Binding      string
	Wildcard     bool
	EnumType     string
	EnumVariant  string
	EnumBindings []string
	Guard        ast.Expr
	Body         []ast.Expr
}

type MatchExpr struct {
	Subject ast.Expr
	Arms    []MatchArm
}

func NewMatchExpr(subject ast.Expr, arms []MatchArm) *MatchExpr {
	return &MatchExpr{Subject: subject, Arms: arms}
}

func (m *MatchExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	subject, err := m.Subject.Eval(ctx)
	if err != nil {
		return value.NewTypeNil(), err
	}

	for _, arm := range m.Arms {
		matched, err := arm.matches(ctx, subject)
		if err != nil {
			return value.NewTypeNil(), err
		}
		if !matched {
			continue
		}

		result := value.NewTypeNil()
		for _, expr := range arm.Body {
			result, err = expr.Eval(ctx)
			if err != nil {
				return value.NewTypeNil(), err
			}
		}
		if ref := m.ValueType(ctx); !ref.IsZero() {
			return value.NewTypeWithExplicit(
				result.Any(),
				ref.String(),
			), nil
		}
		return result, nil
	}

	return value.NewTypeNil(), fmt.Errorf("non-exhaustive match")
}

func (a MatchArm) matches(
	ctx ast.Ctx,
	subject value.Type,
) (bool, error) {
	if a.Binding != "" {
		ctx.SetValue(a.Binding, subject)
	}

	matched := a.Wildcard || a.Binding != ""
	if a.EnumType != "" {
		enumValue, ok := subject.Any().(ast.EnumValue)
		matched = ok &&
			enumValue.TypeName == a.EnumType &&
			enumValue.Variant == a.EnumVariant
		if matched {
			for index, name := range a.EnumBindings {
				if name == "_" {
					continue
				}
				ctx.SetValue(name, enumValue.Payload[index])
			}
		}
	}
	if a.Pattern != nil {
		pattern, err := a.Pattern.Eval(ctx)
		if err != nil {
			return false, err
		}
		matched = reflect.DeepEqual(subject.Any(), pattern.Any())
	}
	if !matched || a.Guard == nil {
		return matched, nil
	}

	guard, err := a.Guard.Eval(ctx)
	if err != nil {
		return false, err
	}
	guardValue, ok := guard.Any().(bool)
	if !ok {
		return false, fmt.Errorf(
			"match guard must be bool, got %s",
			guard.TypeName(),
		)
	}
	return guardValue, nil
}

func (*MatchExpr) Type(ast.Ctx) string {
	return "match"
}

func (m *MatchExpr) ChildExprs() []ast.Expr {
	children := []ast.Expr{m.Subject}
	for _, arm := range m.Arms {
		if arm.Pattern != nil {
			children = append(children, arm.Pattern)
		}
		if arm.Guard != nil {
			children = append(children, arm.Guard)
		}
		children = append(children, arm.Body...)
	}
	return children
}

func (m *MatchExpr) ValueType(ctx ast.Ctx) ast.TypeRef {
	ref, known, err := m.resultType(ctx)
	if err != nil || !known {
		return ast.TypeRef{}
	}
	return ref
}

func (m *MatchExpr) PrintGO(ctx ast.Ctx) (string, error) {
	return m.printGOValue(ctx)
}

func (m *MatchExpr) PrintGOValue(ctx ast.Ctx) (string, error) {
	return m.printGOValue(ctx)
}

func (m *MatchExpr) PrintGOAssign(
	ctx ast.Ctx,
	name string,
) (string, error) {
	printed, err := m.printGOValue(ctx)
	if err != nil {
		return "", err
	}
	return name + " := " + printed, nil
}

func (m *MatchExpr) printGOValue(ctx ast.Ctx) (string, error) {
	if err := m.validatePatterns(ctx); err != nil {
		return "", err
	}
	ref, known, err := m.resultType(ctx)
	if err != nil {
		return "", err
	}
	if !known {
		return "", fmt.Errorf("cannot infer match result type")
	}
	if exhaustive, reason := m.isExhaustive(ctx); !exhaustive {
		return "", fmt.Errorf("non-exhaustive match: %s", reason)
	}

	subject, err := m.Subject.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	resultType := ref.GoString(ctx)
	var out strings.Builder
	out.WriteString("func() ")
	out.WriteString(resultType)
	out.WriteString(" {\n\t__matchValue := ")
	out.WriteString(subject)
	out.WriteString("\n")

	for _, arm := range m.Arms {
		printed, err := arm.printGO(ctx, resultType)
		if err != nil {
			return "", err
		}
		out.WriteString(indentMatch(printed))
		out.WriteByte('\n')
	}

	out.WriteString("\tvar zero ")
	out.WriteString(resultType)
	out.WriteString("\n\treturn zero\n}()")
	return out.String(), nil
}

func (m *MatchExpr) validatePatterns(ctx ast.Ctx) error {
	subjectType := matchExprValueType(ctx, m.Subject)
	for _, arm := range m.Arms {
		if arm.Binding != "" && !subjectType.IsZero() {
			ctx.SetValue(
				arm.Binding,
				value.NewTypeWithExplicit(
					nil,
					subjectType.String(),
				),
			)
		}
		if arm.EnumType != "" {
			if subjectType.IsZero() {
				return fmt.Errorf(
					"cannot infer match subject type for enum pattern %s.%s",
					arm.EnumType,
					arm.EnumVariant,
				)
			}
			if subjectType.Name != arm.EnumType {
				return fmt.Errorf(
					"enum pattern %s.%s is incompatible with %s",
					arm.EnumType,
					arm.EnumVariant,
					subjectType.String(),
				)
			}
			def, variant, err := matchEnumVariant(
				ctx,
				arm.EnumType,
				arm.EnumVariant,
			)
			if err != nil {
				return err
			}
			_ = def
			if len(arm.EnumBindings) != len(variant.Fields) {
				return fmt.Errorf(
					"enum pattern %s.%s expects %d bindings, got %d",
					arm.EnumType,
					arm.EnumVariant,
					len(variant.Fields),
					len(arm.EnumBindings),
				)
			}
			for index, name := range arm.EnumBindings {
				if name == "_" {
					continue
				}
				ctx.SetValue(
					name,
					value.NewTypeWithExplicit(
						nil,
						variant.Fields[index].String(),
					),
				)
			}
		}
		if arm.Pattern != nil {
			patternType := matchExprValueType(ctx, arm.Pattern)
			if !subjectType.IsZero() &&
				!patternType.IsZero() &&
				subjectType.GoString(ctx) !=
					patternType.GoString(ctx) {
				return fmt.Errorf(
					"match pattern type %s is incompatible with %s",
					patternType.String(),
					subjectType.String(),
				)
			}
		}
		if arm.Guard != nil {
			guardType := matchExprValueType(ctx, arm.Guard)
			if !guardType.IsZero() &&
				!strings.EqualFold(guardType.Name, "bool") {
				return fmt.Errorf(
					"match guard must be bool, got %s",
					guardType.String(),
				)
			}
		}
	}
	return nil
}

func (a MatchArm) printGO(
	ctx ast.Ctx,
	resultType string,
) (string, error) {
	body, err := printMatchValueBlock(ctx, a.Body)
	if err != nil {
		return "", err
	}

	var condition string
	if a.EnumType != "" {
		_, variant, err := matchEnumVariant(
			ctx,
			a.EnumType,
			a.EnumVariant,
		)
		if err != nil {
			return "", err
		}
		lines := make([]string, 0, len(a.EnumBindings)+1)
		variantVariable := "_"
		for _, name := range a.EnumBindings {
			if name != "_" {
				variantVariable = "__matchVariant"
				break
			}
		}
		for index, name := range a.EnumBindings {
			if name == "_" {
				continue
			}
			lines = append(
				lines,
				fmt.Sprintf(
					"%s := __matchVariant.V%d",
					name,
					index,
				),
			)
		}
		if a.Guard != nil {
			guard, err := a.Guard.PrintGO(ctx)
			if err != nil {
				return "", err
			}
			lines = append(
				lines,
				"if "+guard+" {\n"+
					indentMatch(body)+"\n}",
			)
		} else {
			lines = append(lines, body)
		}
		_ = variant
		return "if " + variantVariable +
			", ok := __matchValue.(" +
			ast.EnumVariantGoName(
				a.EnumType,
				a.EnumVariant,
			) +
			"); ok {\n" +
			indentMatch(strings.Join(lines, "\n")) +
			"\n}", nil
	}
	switch {
	case a.Pattern != nil:
		pattern, err := a.Pattern.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		condition = "__matchValue == " + pattern
	case a.Binding != "":
		condition = "true"
	case a.Wildcard:
		condition = "true"
	}

	if a.Binding != "" {
		if a.Guard == nil {
			body = a.Binding + " := __matchValue\n" + body
		}
	}
	if a.Guard != nil {
		guard, err := a.Guard.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		if a.Binding != "" {
			return "if " + a.Binding + " := __matchValue; " +
				guard + " {\n" +
				indentMatch(body) + "\n}", nil
		}
		condition = "(" + condition + ") && (" + guard + ")"
	}

	return "if " + condition + " {\n" +
		indentMatch(body) + "\n}", nil
}

func (m *MatchExpr) resultType(
	ctx ast.Ctx,
) (ast.TypeRef, bool, error) {
	var result ast.TypeRef
	known := false
	subjectType := matchExprValueType(ctx, m.Subject)

	for _, arm := range m.Arms {
		if arm.EnumType != "" {
			_, variant, err := matchEnumVariant(
				ctx,
				arm.EnumType,
				arm.EnumVariant,
			)
			if err != nil {
				return ast.TypeRef{}, false, err
			}
			for index, name := range arm.EnumBindings {
				if name == "_" || index >= len(variant.Fields) {
					continue
				}
				ctx.SetValue(
					name,
					value.NewTypeWithExplicit(
						nil,
						variant.Fields[index].String(),
					),
				)
			}
		}
		if arm.Binding != "" && !subjectType.IsZero() {
			ctx.SetValue(
				arm.Binding,
				value.NewTypeWithExplicit(nil, subjectType.String()),
			)
		}
		ref, armKnown := matchBlockValueType(ctx, arm.Body)
		if !armKnown {
			return ast.TypeRef{}, false, nil
		}
		if !known {
			result = ref
			known = true
			continue
		}
		if result.GoString(ctx) != ref.GoString(ctx) {
			return ast.TypeRef{}, false, fmt.Errorf(
				"match arms return incompatible types %s and %s",
				result.String(),
				ref.String(),
			)
		}
	}
	return result, known, nil
}

func matchExprValueType(ctx ast.Ctx, expr ast.Expr) ast.TypeRef {
	if typed, ok := expr.(interface {
		ValueType(ast.Ctx) ast.TypeRef
	}); ok {
		return typed.ValueType(ctx)
	}
	switch expr.(type) {
	case literal_parser.Int:
		return ast.TypeRef{Name: "int"}
	case literal_parser.Float64:
		return ast.TypeRef{Name: "float64"}
	case literal_parser.String:
		return ast.TypeRef{Name: "string"}
	case literal_parser.Bool:
		return ast.TypeRef{Name: "bool"}
	}
	return ast.TypeRef{}
}

func (m *MatchExpr) isExhaustive(
	ctx ast.Ctx,
) (bool, string) {
	for _, arm := range m.Arms {
		if arm.Guard == nil &&
			(arm.Wildcard || arm.Binding != "") {
			return true, ""
		}
	}

	subjectType := matchExprValueType(ctx, m.Subject)
	def, declared := ctx.GetType(subjectType.Name)
	if !declared || def.Kind != ast.EnumType {
		return false, "add an unguarded _ or binding arm"
	}

	covered := map[string]struct{}{}
	for _, arm := range m.Arms {
		if arm.EnumType == def.Name && arm.Guard == nil {
			covered[arm.EnumVariant] = struct{}{}
		}
	}
	var missing []string
	for _, variant := range def.Variants {
		if _, ok := covered[variant.Name]; !ok {
			missing = append(missing, def.Name+"."+variant.Name)
		}
	}
	if len(missing) != 0 {
		return false, "missing " + strings.Join(missing, ", ")
	}
	return true, ""
}

func matchEnumVariant(
	ctx ast.Ctx,
	typeName string,
	variantName string,
) (ast.TypeDef, ast.EnumVariantDef, error) {
	def, declared := ctx.GetType(typeName)
	if !declared || def.Kind != ast.EnumType {
		return ast.TypeDef{}, ast.EnumVariantDef{}, fmt.Errorf(
			"type %s is not an enum",
			typeName,
		)
	}
	for _, variant := range def.Variants {
		if variant.Name == variantName {
			return def, variant, nil
		}
	}
	return ast.TypeDef{}, ast.EnumVariantDef{}, fmt.Errorf(
		"variant %s does not exist on enum %s",
		variantName,
		typeName,
	)
}

func matchBlockValueType(
	ctx ast.Ctx,
	body []ast.Expr,
) (ast.TypeRef, bool) {
	if len(body) == 0 {
		return ast.TypeRef{}, false
	}
	expr := body[len(body)-1]
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

func printMatchValueBlock(
	ctx ast.Ctx,
	body []ast.Expr,
) (string, error) {
	lines := make([]string, 0, len(body))
	for index, expr := range body {
		var (
			printed string
			err     error
		)
		if valueExpr, ok := expr.(interface {
			PrintGOValue(ast.Ctx) (string, error)
		}); ok {
			printed, err = valueExpr.PrintGOValue(ctx)
		} else {
			printed, err = expr.PrintGO(ctx)
		}
		if err != nil {
			return "", err
		}
		if index == len(body)-1 {
			printed = "return " + printed
		}
		lines = append(lines, printed)
	}
	return strings.Join(lines, "\n"), nil
}

func indentMatch(code string) string {
	return "\t" + strings.ReplaceAll(code, "\n", "\n\t")
}
