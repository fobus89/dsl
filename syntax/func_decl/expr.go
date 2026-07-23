package funcdecl_parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fobus89/dsl/ast"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	"github.com/fobus89/dsl/value"
)

type (
	Expr  = ast.Expr
	Ident = literal_parser.Ident
)

type Param struct {
	Name  Ident
	Type  Expr
	IsPtr bool
}

// FuncDecl represents both a function and a method declaration.
// A declaration is a method when Recv is not nil.
type FuncDecl struct {
	Recv       *Param
	Name       Ident
	Params     []Param
	ReturnType Expr
	Body       []Expr
}

func NewFuncDecl(
	ctx ast.Ctx,
	recv *Param,
	name Ident,
	params []Param,
	returnType Expr,
	body []Expr,
) *FuncDecl {
	decl := &FuncDecl{
		Recv:       recv,
		Name:       name,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}

	decl.bind(ctx)
	return decl
}

func (d *FuncDecl) IsMethod() bool {
	return d.Recv != nil
}

func (d *FuncDecl) Validate(ctx ast.Ctx) error {
	if !d.IsMethod() {
		return nil
	}

	receiverType := d.Recv.Type.(Ident)
	if _, ok := ctx.GetType(string(receiverType)); !ok {
		return fmt.Errorf("receiver type %s is not declared", receiverType)
	}

	return nil
}

func (d *FuncDecl) bind(ctx ast.Ctx) {
	fn := func(args ...value.Type) (value.Type, error) {
		receiverCount := 0
		if d.IsMethod() {
			receiverCount = 1
		}

		expectedArgs := receiverCount + len(d.Params)
		if len(args) != expectedArgs {
			return value.NewTypeNil(), fmt.Errorf(
				"func %s expects %d arguments, got %d",
				d.Name,
				expectedArgs,
				len(args),
			)
		}

		localCtx := ctx.GetLocalCtx()
		if d.IsMethod() {
			localCtx.SetValue(string(d.Recv.Name), args[0])
		}
		for i, param := range d.Params {
			localCtx.SetValue(string(param.Name), args[i+receiverCount])
		}

		for _, expr := range d.Body {
			_, err := expr.Eval(localCtx)
			if err == nil {
				continue
			}

			if returned, ok := errors.AsType[returnSignal](err); ok {
				return d.coerceReturn(ctx, returned.value)
			}

			return value.NewTypeNil(), err
		}

		return value.NewTypeNil(), nil
	}

	if d.IsMethod() {
		receiverType := d.Recv.Type.(Ident)
		ctx.SetMethod(string(receiverType), string(d.Name), fn)
		return
	}

	ctx.SetFunc(string(d.Name), fn)
}

func (d *FuncDecl) coerceReturn(
	ctx ast.Ctx,
	result value.Type,
) (value.Type, error) {
	if d.ReturnType == nil {
		return result, nil
	}

	returnType, ok := d.ReturnType.(Ident)
	if !ok {
		return value.NewTypeNil(), fmt.Errorf(
			"invalid return type %T for func %s",
			d.ReturnType,
			d.Name,
		)
	}

	name := string(returnType)
	if _, declared := ctx.GetType(name); declared {
		if result.TypeName() == name {
			return result, nil
		}
		return value.NewTypeWithExplicit(result.Any(), name), nil
	}

	switch strings.ToLower(name) {
	case "any":
		return result, nil
	case "void":
		return value.NewTypeNil(), nil
	case "int":
		if v, ok := result.CastInt(); ok {
			return value.NewType(v), nil
		}
	case "int8":
		if v, ok := result.CastInt8(); ok {
			return value.NewType(v), nil
		}
	case "int16":
		if v, ok := result.CastInt16(); ok {
			return value.NewType(v), nil
		}
	case "int32", "rune":
		if v, ok := result.CastInt32(); ok {
			return value.NewType(v), nil
		}
	case "int64":
		if v, ok := result.CastInt64(); ok {
			return value.NewType(v), nil
		}
	case "uint":
		if v, ok := result.CastUint(); ok {
			return value.NewType(v), nil
		}
	case "uint8", "byte":
		if v, ok := result.CastUint8(); ok {
			return value.NewType(v), nil
		}
	case "uint16":
		if v, ok := result.CastUint16(); ok {
			return value.NewType(v), nil
		}
	case "uint32":
		if v, ok := result.CastUint32(); ok {
			return value.NewType(v), nil
		}
	case "uint64":
		if v, ok := result.CastUint64(); ok {
			return value.NewType(v), nil
		}
	case "float32":
		if v, ok := result.CastFloat32(); ok {
			return value.NewType(v), nil
		}
	case "float64":
		if v, ok := result.CastFloat64(); ok {
			return value.NewType(v), nil
		}
	case "string":
		if v, ok := result.CastString(); ok {
			return value.NewType(v), nil
		}
	case "bool":
		if v, ok := result.CastBool(); ok {
			return value.NewType(v), nil
		}
	}

	return value.NewTypeNil(), fmt.Errorf(
		"func %s cannot return %s as %s",
		d.Name,
		result.TypeName(),
		name,
	)
}

func (*FuncDecl) Eval(ast.Ctx) (value.Type, error) {
	return value.NewTypeNil(), nil
}

func (*FuncDecl) Type(_ ast.Ctx) string {
	return "func_decl"
}

func (d *FuncDecl) PrintGO(ctx ast.Ctx) (string, error) {
	printCtx := ctx.GetLocalCtx()

	if d.Recv != nil {
		receiverType, err := d.Recv.Type.PrintGO(ctx)
		if err != nil {
			return "", err
		}

		if def, ok := ctx.GetType(receiverType); ok {
			if _, exists := def.Fields[string(d.Name)]; exists {
				return "", fmt.Errorf(
					"type %s has both field and method named %s",
					receiverType,
					d.Name,
				)
			}
		}

		printCtx.SetValue(
			string(d.Recv.Name),
			value.NewTypeWithExplicit(nil, receiverType),
		)
	}

	params := make([]string, 0, len(d.Params))
	for _, param := range d.Params {
		paramType := "any"
		if param.Type != nil {
			printed, err := param.Type.PrintGO(ctx)
			if err != nil {
				return "", err
			}
			paramType = ast.GoTypeName(printed)
			printCtx.SetValue(
				string(param.Name),
				value.NewTypeWithExplicit(nil, printed),
			)
		}
		params = append(params, string(param.Name)+" "+paramType)
	}

	receiver := ""
	if d.Recv != nil {
		receiverName, err := d.Recv.Name.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		receiverType, err := d.Recv.Type.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		receiverPointer := ""
		if d.Recv.IsPtr {
			receiverPointer = "*"
		}
		receiver = fmt.Sprintf(
			"(%s %s%s) ",
			receiverName,
			receiverPointer,
			receiverType,
		)
	}

	returnType := ""
	if d.ReturnType != nil {
		printed, err := d.ReturnType.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		returnType = " " + ast.GoTypeName(printed)
	}

	body := make([]string, 0, len(d.Body))
	for _, expr := range d.Body {
		printed, err := expr.PrintGO(printCtx)
		if err != nil {
			return "", err
		}
		body = append(body, "\t"+printed)
	}

	return fmt.Sprintf(
		"func %s%s(%s)%s {\n%s\n}",
		receiver,
		d.Name,
		strings.Join(params, ", "),
		returnType,
		strings.Join(body, "\n"),
	), nil
}

type ReturnStmt struct {
	Value Expr
}

func NewReturnStmt(value Expr) *ReturnStmt {
	return &ReturnStmt{Value: value}
}

func (r *ReturnStmt) Eval(ctx ast.Ctx) (value.Type, error) {
	result := value.NewTypeNil()

	if r.Value != nil {
		evaluated, err := r.Value.Eval(ctx)
		if err != nil {
			return value.NewTypeNil(), err
		}
		result = evaluated
	}

	return value.NewTypeNil(), returnSignal{value: result}
}

func (*ReturnStmt) Type(_ ast.Ctx) string {
	return "return_stmt"
}

func (r *ReturnStmt) PrintGO(ctx ast.Ctx) (string, error) {
	if r.Value == nil {
		return "return", nil
	}

	value, err := r.Value.PrintGO(ctx)
	if err != nil {
		return "", err
	}

	return "return " + value, nil
}

type returnSignal struct {
	value value.Type
}

func (returnSignal) Error() string {
	return "return"
}
