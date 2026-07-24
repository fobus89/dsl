package generic_parser

import (
	"fmt"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type GenericExpr struct {
	Base ast.Expr
	Args []ast.TypeRef
}

func NewGenericExpr(
	base ast.Expr,
	args []ast.TypeRef,
) *GenericExpr {
	return &GenericExpr{Base: base, Args: args}
}

func (g *GenericExpr) Eval(ctx ast.Ctx) (value.Type, error) {
	if err := g.ValidateGeneric(ctx); err != nil {
		return value.NewTypeNil(), err
	}
	return g.Base.Eval(ctx)
}

func (*GenericExpr) Type(ast.Ctx) string {
	return "generic"
}

func (g *GenericExpr) PrintGO(ctx ast.Ctx) (string, error) {
	if err := g.ValidateGeneric(ctx); err != nil {
		return "", err
	}
	g.RegisterInstantiation(ctx)
	base, err := g.Base.PrintGO(ctx)
	if err != nil {
		return "", err
	}
	if ident, ok := g.Base.(literal_parser.Ident); ok {
		return ast.MangleGenericName(
			string(ident),
			g.Args,
		), nil
	}
	args := make([]string, 0, len(g.Args))
	for _, arg := range g.Args {
		args = append(args, arg.String())
	}
	return base + "_" + strings.Join(args, "_"), nil
}

func (g *GenericExpr) RegisterInstantiation(ctx ast.Ctx) {
	if ident, ok := g.Base.(literal_parser.Ident); ok {
		name := string(ident)
		if _, declared := ctx.GetType(name); declared {
			ctx.RegisterGenericInstance("type:"+name, g.Args)
			return
		}
		ctx.RegisterGenericInstance("func:"+name, g.Args)
		return
	}
	if keyed, ok := g.Base.(interface {
		GenericSignatureKey(ast.Ctx) (string, bool)
		GenericReceiverType(ast.Ctx) ast.TypeRef
	}); ok {
		key, exists := keyed.GenericSignatureKey(ctx)
		if !exists {
			return
		}
		receiver := keyed.GenericReceiverType(ctx)
		args := append([]ast.TypeRef{}, receiver.Args...)
		args = append(args, g.Args...)
		ctx.RegisterGenericInstance(key, args)
	}
}

func (g *GenericExpr) ValidateGeneric(ctx ast.Ctx) error {
	if ident, ok := g.Base.(literal_parser.Ident); ok {
		name := string(ident)
		if def, declared := ctx.GetType(name); declared {
			return ast.ValidateTypeArguments(
				ctx,
				"type "+name,
				def.TypeParams,
				g.Args,
			)
		}
		if params, declared := ctx.GetGeneric(
			"func:" + name,
		); declared {
			return ast.ValidateTypeArguments(
				ctx,
				"func "+name,
				params,
				g.Args,
			)
		}
		return fmt.Errorf("%s is not generic", name)
	}
	if keyed, ok := g.Base.(interface {
		GenericSignatureKey(ast.Ctx) (string, bool)
	}); ok {
		key, exists := keyed.GenericSignatureKey(ctx)
		if !exists {
			return fmt.Errorf("method is not generic")
		}
		params, declared := ctx.GetGeneric(key)
		if !declared {
			return fmt.Errorf("method is not generic")
		}
		return ast.ValidateTypeArguments(
			ctx,
			key,
			params,
			g.Args,
		)
	}
	return fmt.Errorf("expression is not generic")
}

func (g *GenericExpr) BaseExpr() ast.Expr {
	return g.Base
}

func (g *GenericExpr) TypeArgs() []ast.TypeRef {
	return g.Args
}

func (g *GenericExpr) StructTypeName() string {
	if ident, ok := g.Base.(literal_parser.Ident); ok {
		return string(ident)
	}
	return ""
}

func (g *GenericExpr) TypeReference() ast.TypeRef {
	return ast.TypeRef{
		Name: g.StructTypeName(),
		Args: g.Args,
	}
}

func (g *GenericExpr) ValueType(ast.Ctx) ast.TypeRef {
	return ast.TypeRef{
		Name: g.StructTypeName(),
		Args: g.Args,
	}
}
