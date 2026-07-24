package collection_parser

import (
	"fmt"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type CollectionExpr struct {
	TypeRef  ast.TypeRef
	Elements []ast.Expr
}

func NewCollectionExpr(
	typeRef ast.TypeRef,
	elements []ast.Expr,
) *CollectionExpr {
	return &CollectionExpr{TypeRef: typeRef, Elements: elements}
}

func (c *CollectionExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	length := len(c.Elements)
	if c.TypeRef.Kind == ast.ArrayTypeRef {
		length = c.TypeRef.Len
	}
	result := make([]any, length)
	if c.TypeRef.Kind == ast.ArrayTypeRef &&
		c.TypeRef.Elem != nil {
		for i := range result {
			result[i] = collectionZeroValue(ctx, *c.TypeRef.Elem)
		}
	}

	for i, element := range c.Elements {
		evaluated, err := element.Eval(ctx)
		if err != nil {
			return value.NewTypeNil(), err
		}
		if c.TypeRef.Elem != nil &&
			!collectionValueMatches(ctx, *c.TypeRef.Elem, evaluated) {
			return value.NewTypeNil(), fmt.Errorf(
				"element %d of %s has type %s, expected %s",
				i,
				c.TypeRef.String(),
				evaluated.TypeName(),
				c.TypeRef.Elem.String(),
			)
		}
		result[i] = evaluated.Any()
	}

	return value.NewTypeWithExplicit(result, c.TypeRef.String()), nil
}

func (c *CollectionExpr) Type(ast.Ctx) string {
	return c.TypeRef.String()
}

func (c *CollectionExpr) ValueType(ast.Ctx) ast.TypeRef {
	return c.TypeRef
}

func (c *CollectionExpr) PrintGO(ctx ast.Ctx) (string, error) {
	elements := make([]string, 0, len(c.Elements))
	for i, element := range c.Elements {
		if c.TypeRef.Elem != nil {
			if actual, known := collectionExprType(ctx, element); known &&
				!collectionTypeMatches(ctx, *c.TypeRef.Elem, actual) {
				return "", fmt.Errorf(
					"element %d of %s has type %s, expected %s",
					i,
					c.TypeRef.String(),
					actual.String(),
					c.TypeRef.Elem.String(),
				)
			}
		}

		printed, err := element.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		elements = append(elements, printed)
	}

	return c.TypeRef.GoString(ctx) +
		"{" + strings.Join(elements, ", ") + "}", nil
}

func collectionValueMatches(
	ctx ast.Ctx,
	target ast.TypeRef,
	actual value.Type,
) bool {
	if actual.IsNil() {
		return target.IsPtr ||
			target.Kind == ast.SliceTypeRef ||
			strings.EqualFold(target.Name, "any")
	}

	if target.Kind == ast.ArrayTypeRef ||
		target.Kind == ast.SliceTypeRef {
		return actual.TypeName() == target.String()
	}
	if target.Kind == ast.TupleTypeRef {
		return actual.TypeName() == target.String()
	}
	if target.IsPtr {
		return actual.TypeName() == target.String()
	}
	if def, declared := ctx.GetType(target.Name); declared {
		if actual.TypeName() == target.Name {
			return true
		}
		return def.Kind == ast.AliasType &&
			collectionValueMatches(ctx, def.Underlying, actual)
	}

	switch strings.ToLower(target.Name) {
	case "any":
		return true
	case "string":
		return actual.IsString()
	case "bool":
		return actual.IsBool()
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "rune", "byte":
		return actual.IsNumber()
	}

	return actual.TypeName() == target.Name
}

func collectionExprType(
	ctx ast.Ctx,
	expr ast.Expr,
) (ast.TypeRef, bool) {
	switch expr.(type) {
	case literal_parser.Int:
		return ast.TypeRef{Name: "int"}, true
	case literal_parser.Float64:
		return ast.TypeRef{Name: "float64"}, true
	case literal_parser.String:
		return ast.TypeRef{Name: "string"}, true
	case literal_parser.Bool:
		return ast.TypeRef{Name: "bool"}, true
	case literal_parser.Nil:
		return ast.TypeRef{Name: "nil"}, true
	}

	if typed, ok := expr.(interface {
		ValueType(ast.Ctx) ast.TypeRef
	}); ok {
		ref := typed.ValueType(ctx)
		return ref, !ref.IsZero()
	}
	if ident, ok := expr.(literal_parser.Ident); ok {
		if actual, found := ctx.GetValue(string(ident)); found {
			return ast.ParseTypeRef(actual.TypeName()), true
		}
	}

	return ast.TypeRef{}, false
}

func collectionTypeMatches(
	ctx ast.Ctx,
	target ast.TypeRef,
	actual ast.TypeRef,
) bool {
	if actual.Kind == ast.NamedTypeRef && actual.Name == "nil" {
		return target.IsPtr || target.Kind == ast.SliceTypeRef
	}
	if target.IsPtr != actual.IsPtr {
		return false
	}
	if target.Kind != actual.Kind {
		return false
	}
	if target.Kind == ast.ArrayTypeRef {
		if target.Len != actual.Len ||
			target.Elem == nil ||
			actual.Elem == nil {
			return false
		}
		return collectionTypeMatches(ctx, *target.Elem, *actual.Elem)
	}
	if target.Kind == ast.SliceTypeRef {
		if target.Elem == nil || actual.Elem == nil {
			return false
		}
		return collectionTypeMatches(ctx, *target.Elem, *actual.Elem)
	}

	if target.Name == actual.Name {
		return true
	}
	if def, declared := ctx.GetType(target.Name); declared &&
		def.Kind == ast.AliasType {
		return collectionTypeMatches(ctx, def.Underlying, actual)
	}

	return isCollectionNumeric(target.Name) &&
		isCollectionNumeric(actual.Name)
}

func isCollectionNumeric(name string) bool {
	switch strings.ToLower(name) {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "rune", "byte":
		return true
	}
	return false
}

func collectionZeroValue(ctx ast.Ctx, ref ast.TypeRef) any {
	if ref.IsPtr || ref.Kind == ast.SliceTypeRef {
		return nil
	}
	if ref.Kind == ast.ArrayTypeRef {
		result := make([]any, ref.Len)
		if ref.Elem != nil {
			for i := range result {
				result[i] = collectionZeroValue(ctx, *ref.Elem)
			}
		}
		return result
	}

	if def, declared := ctx.GetType(ref.Name); declared {
		if def.Kind == ast.AliasType {
			return collectionZeroValue(ctx, def.Underlying)
		}
		fields := make(map[string]any, len(def.Fields))
		for name, field := range def.Fields {
			fields[name] = collectionZeroValue(ctx, field.Type)
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
