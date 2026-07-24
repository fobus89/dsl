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
	Name Ident
	Type *ast.TypeRef
}

// FuncDecl represents both a function and a method declaration.
// A declaration is a method when Recv is not nil.
type FuncDecl struct {
	Recv       *Param
	Name       Ident
	Params     []Param
	ReturnType *ast.TypeRef
	Body       []Expr
	IsComptime bool
}

func NewFuncDecl(
	ctx ast.Ctx,
	recv *Param,
	name Ident,
	params []Param,
	returnType *ast.TypeRef,
	body []Expr,
) *FuncDecl {
	return newFuncDecl(
		ctx,
		recv,
		name,
		params,
		returnType,
		body,
		false,
	)
}

func NewComptimeFuncDecl(
	ctx ast.Ctx,
	recv *Param,
	name Ident,
	params []Param,
	returnType *ast.TypeRef,
	body []Expr,
) *FuncDecl {
	return newFuncDecl(
		ctx,
		recv,
		name,
		params,
		returnType,
		body,
		true,
	)
}

func newFuncDecl(
	ctx ast.Ctx,
	recv *Param,
	name Ident,
	params []Param,
	returnType *ast.TypeRef,
	body []Expr,
	isComptime bool,
) *FuncDecl {
	decl := &FuncDecl{
		Recv:       recv,
		Name:       name,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
		IsComptime: isComptime,
	}

	decl.bind(ctx)
	return decl
}

func (d *FuncDecl) IsMethod() bool {
	return d.Recv != nil
}

func (d *FuncDecl) Validate(ctx ast.Ctx) error {
	if d.usesMetaTypes() && !d.IsComptime {
		return fmt.Errorf(
			"func %s uses meta types and must be declared comptime",
			d.Name,
		)
	}

	if !d.IsMethod() {
		return nil
	}

	if d.Recv.Type.IsDirectMeta() {
		return nil
	}

	if d.Recv.Type.Kind != ast.NamedTypeRef ||
		d.Recv.Type.Name == "" {
		return fmt.Errorf(
			"receiver type %s must be a declared named type",
			d.Recv.Type.String(),
		)
	}

	receiverType := d.Recv.Type.Name
	if _, ok := ctx.GetType(receiverType); !ok {
		return fmt.Errorf("receiver type %s is not declared", receiverType)
	}

	return nil
}

func (d *FuncDecl) usesMetaTypes() bool {
	if d.Recv != nil &&
		d.Recv.Type != nil &&
		d.Recv.Type.IsMeta() {
		return true
	}
	if d.ReturnType != nil && d.ReturnType.IsMeta() {
		return true
	}
	for _, param := range d.Params {
		if param.Type != nil && param.Type.IsMeta() {
			return true
		}
	}
	return false
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
			if d.Recv.Type.IsDirectMeta() {
				meta, ok := args[0].Any().(ast.MetaTypeValue)
				if !ok || !meta.Matches(d.Recv.Type.Name) {
					actual := args[0].TypeName()
					if ok {
						actual = string(meta.Kind)
					}
					return value.NewTypeNil(), fmt.Errorf(
						"method %s receiver expects %s, got %s",
						d.Name,
						d.Recv.Type.Name,
						actual,
					)
				}
			}
			localCtx.SetValue(string(d.Recv.Name), args[0])
		}
		for i, param := range d.Params {
			arg := args[i+receiverCount]
			if param.Type != nil && param.Type.IsDirectMeta() {
				meta, ok := arg.Any().(ast.MetaTypeValue)
				if !ok || !meta.Matches(param.Type.Name) {
					actual := arg.TypeName()
					if ok {
						actual = string(meta.Kind)
					}
					return value.NewTypeNil(), fmt.Errorf(
						"func %s parameter %s expects %s, got %s",
						d.Name,
						param.Name,
						param.Type.Name,
						actual,
					)
				}
			}
			localCtx.SetValue(string(param.Name), arg)
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
		ctx.SetMethod(d.Recv.Type.Name, string(d.Name), fn)
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

	if d.ReturnType.Kind == ast.ArrayTypeRef ||
		d.ReturnType.Kind == ast.SliceTypeRef ||
		d.ReturnType.Kind == ast.TupleTypeRef ||
		d.ReturnType.IsPtr {
		if result.TypeName() == d.ReturnType.String() ||
			(result.IsNil() &&
				(d.ReturnType.IsPtr ||
					d.ReturnType.Kind == ast.SliceTypeRef)) {
			return result, nil
		}
		return value.NewTypeNil(), fmt.Errorf(
			"func %s cannot return %s as %s",
			d.Name,
			result.TypeName(),
			d.ReturnType.String(),
		)
	}

	name := d.ReturnType.Name
	if ast.IsMetaTypeName(name) {
		meta, ok := result.Any().(ast.MetaTypeValue)
		if ok && meta.Matches(name) {
			return result, nil
		}
		actual := result.TypeName()
		if ok {
			actual = string(meta.Kind)
		}
		return value.NewTypeNil(), fmt.Errorf(
			"func %s cannot return %s as %s",
			d.Name,
			actual,
			name,
		)
	}

	if def, declared := ctx.GetType(name); declared {
		if result.TypeName() == name {
			return result, nil
		}
		if def.Kind == ast.AliasType &&
			returnValueMatches(ctx, def.Underlying, result) {
			return value.NewTypeWithExplicit(result.Any(), name), nil
		}
		return value.NewTypeNil(), fmt.Errorf(
			"func %s cannot return %s as %s",
			d.Name,
			result.TypeName(),
			name,
		)
	}

	switch strings.ToLower(name) {
	case "any":
		return result, nil
	case "void":
		if result.IsNil() {
			return value.NewTypeNil(), nil
		}
	case "int":
		if v, ok := result.CastInt(); ok && result.IsNumber() {
			return value.NewType(v), nil
		}
	case "int8":
		if v, ok := result.CastInt8(); ok && result.IsNumber() {
			return value.NewType(v), nil
		}
	case "int16":
		if v, ok := result.CastInt16(); ok && result.IsNumber() {
			return value.NewType(v), nil
		}
	case "int32", "rune":
		if v, ok := result.CastInt32(); ok && result.IsNumber() {
			return value.NewType(v), nil
		}
	case "int64":
		if v, ok := result.CastInt64(); ok && result.IsNumber() {
			return value.NewType(v), nil
		}
	case "uint":
		if v, ok := result.CastUint(); ok && result.IsNumber() {
			return value.NewType(v), nil
		}
	case "uint8", "byte":
		if v, ok := result.CastUint8(); ok && result.IsNumber() {
			return value.NewType(v), nil
		}
	case "uint16":
		if v, ok := result.CastUint16(); ok && result.IsNumber() {
			return value.NewType(v), nil
		}
	case "uint32":
		if v, ok := result.CastUint32(); ok && result.IsNumber() {
			return value.NewType(v), nil
		}
	case "uint64":
		if v, ok := result.CastUint64(); ok && result.IsNumber() {
			return value.NewType(v), nil
		}
	case "float32":
		if v, ok := result.CastFloat32(); ok && result.IsNumber() {
			return value.NewType(v), nil
		}
	case "float64":
		if v, ok := result.CastFloat64(); ok && result.IsNumber() {
			return value.NewType(v), nil
		}
	case "string":
		if result.IsString() {
			return result, nil
		}
	case "bool":
		if result.IsBool() {
			return result, nil
		}
	}

	return value.NewTypeNil(), fmt.Errorf(
		"func %s cannot return %s as %s",
		d.Name,
		result.TypeName(),
		name,
	)
}

func returnValueMatches(
	ctx ast.Ctx,
	target ast.TypeRef,
	result value.Type,
) bool {
	if target.Kind == ast.ArrayTypeRef ||
		target.Kind == ast.SliceTypeRef ||
		target.Kind == ast.TupleTypeRef {
		return result.TypeName() == target.String() ||
			(result.IsNil() && target.Kind == ast.SliceTypeRef)
	}
	if target.IsPtr {
		return result.TypeName() == target.String()
	}

	if def, declared := ctx.GetType(target.Name); declared {
		if result.TypeName() == target.Name {
			return true
		}
		return def.Kind == ast.AliasType &&
			returnValueMatches(ctx, def.Underlying, result)
	}

	switch strings.ToLower(target.Name) {
	case "any":
		return true
	case "void":
		return result.IsNil()
	case "string":
		return result.IsString()
	case "bool":
		return result.IsBool()
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "rune", "byte":
		return result.IsNumber()
	}
	if ast.IsMetaTypeName(target.Name) {
		meta, ok := result.Any().(ast.MetaTypeValue)
		return ok && meta.Matches(target.Name)
	}

	return false
}

func (*FuncDecl) Eval(ast.Ctx) (value.Type, error) {
	return value.NewTypeNil(), nil
}

func (*FuncDecl) Type(_ ast.Ctx) string {
	return "func_decl"
}

func (d *FuncDecl) PrintGO(ctx ast.Ctx) (string, error) {
	if d.IsComptime {
		return "", nil
	}
	if d.ReturnType != nil && d.ReturnType.IsMeta() {
		return "", metaRuntimeError(d.Name)
	}
	for _, param := range d.Params {
		if param.Type != nil && param.Type.IsMeta() {
			return "", metaRuntimeError(d.Name)
		}
	}

	printCtx := ctx.GetLocalCtx()

	if d.Recv != nil {
		receiverType := d.Recv.Type.Name

		def, ok := ctx.GetType(receiverType)
		if !ok {
			return "", fmt.Errorf(
				"receiver type %s is not declared",
				receiverType,
			)
		}
		if _, exists := def.Fields[string(d.Name)]; exists {
			return "", fmt.Errorf(
				"type %s has both field and method named %s",
				receiverType,
				d.Name,
			)
		}
		if def.Kind == ast.EnumType {
			for _, variant := range def.Variants {
				if variant.Name == string(d.Name) {
					return "", fmt.Errorf(
						"enum %s has both variant and method named %s",
						receiverType,
						d.Name,
					)
				}
			}
		}

		printCtx.SetValue(
			string(d.Recv.Name),
			value.NewTypeWithExplicit(nil, d.Recv.Type.String()),
		)
	}

	params := make([]string, 0, len(d.Params))
	for _, param := range d.Params {
		paramType := "any"
		if param.Type != nil {
			paramType = param.Type.GoString(ctx)
			printCtx.SetValue(
				string(param.Name),
				value.NewTypeWithExplicit(nil, param.Type.String()),
			)
		}
		params = append(params, string(param.Name)+" "+paramType)
	}

	receiver := ""
	goName := string(d.Name)
	if d.Recv != nil {
		receiverName, err := d.Recv.Name.PrintGO(ctx)
		if err != nil {
			return "", err
		}
		receiverType := d.Recv.Type.GoString(ctx)
		def, _ := ctx.GetType(d.Recv.Type.Name)
		if def.Kind == ast.EnumType {
			params = append(
				[]string{receiverName + " " + receiverType},
				params...,
			)
			goName = d.Recv.Type.Name + "_" + string(d.Name)
		} else {
			receiver = fmt.Sprintf(
				"(%s %s) ",
				receiverName,
				receiverType,
			)
		}
	}

	returnType := ""
	if d.ReturnType != nil {
		returnType = " " + d.ReturnType.GoString(ctx)
	}

	body := make([]string, 0, len(d.Body))
	for _, expr := range d.Body {
		if err := validateReturnTree(
			printCtx,
			string(d.Name),
			d.ReturnType,
			expr,
		); err != nil {
			return "", err
		}

		printed, err := expr.PrintGO(printCtx)
		if err != nil {
			return "", err
		}
		body = append(body, "\t"+printed)
	}

	return fmt.Sprintf(
		"func %s%s(%s)%s {\n%s\n}",
		receiver,
		goName,
		strings.Join(params, ", "),
		returnType,
		strings.Join(body, "\n"),
	), nil
}

func metaRuntimeError(name Ident) error {
	return fmt.Errorf(
		"func %s uses type values and must be declared comptime",
		name,
	)
}

func validateReturnTree(
	ctx ast.Ctx,
	funcName string,
	target *ast.TypeRef,
	expr ast.Expr,
) error {
	if returned, ok := expr.(*ReturnStmt); ok {
		return validateReturnGO(ctx, funcName, target, returned)
	}

	if container, ok := expr.(interface{ ChildExprs() []ast.Expr }); ok {
		for _, child := range container.ChildExprs() {
			if err := validateReturnTree(
				ctx,
				funcName,
				target,
				child,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateReturnGO(
	ctx ast.Ctx,
	funcName string,
	target *ast.TypeRef,
	returned *ReturnStmt,
) error {
	if target == nil ||
		(target.Kind == ast.NamedTypeRef &&
			strings.EqualFold(target.Name, "any")) {
		return nil
	}

	if returned.Value == nil {
		if strings.EqualFold(target.Name, "void") {
			return nil
		}
		return fmt.Errorf(
			"func %s must return %s",
			funcName,
			target.String(),
		)
	}

	if target.Kind == ast.NamedTypeRef &&
		strings.EqualFold(target.Name, "void") {
		return fmt.Errorf("func %s cannot return a value as void", funcName)
	}

	source, known := goExprType(ctx, returned.Value)
	if !known {
		return nil
	}

	if source.Name == "nil" {
		if target.IsPtr {
			return nil
		}
		return returnTypeError(funcName, source, *target)
	}

	if source.IsPtr != target.IsPtr {
		return returnTypeError(funcName, source, *target)
	}

	sourceName := strings.TrimPrefix(source.GoString(ctx), "*")
	targetName := strings.TrimPrefix(target.GoString(ctx), "*")
	if sourceName == targetName {
		return nil
	}
	sourceUnderlying := underlyingGoType(ctx, source)
	targetUnderlying := underlyingGoType(ctx, *target)
	if sourceUnderlying == targetUnderlying ||
		(isNumericType(sourceUnderlying) &&
			isNumericType(targetUnderlying)) {
		return nil
	}

	return returnTypeError(funcName, source, *target)
}

func underlyingGoType(ctx ast.Ctx, ref ast.TypeRef) string {
	if ref.Kind != ast.NamedTypeRef {
		return ref.GoString(ctx)
	}
	if def, declared := ctx.GetType(ref.Name); declared &&
		def.Kind == ast.AliasType {
		return underlyingGoType(ctx, def.Underlying)
	}
	return ast.GoTypeName(ref.Name)
}

func goExprType(ctx ast.Ctx, expr ast.Expr) (ast.TypeRef, bool) {
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
		valueType := typed.ValueType(ctx)
		return valueType, !valueType.IsZero()
	}

	if ident, ok := expr.(Ident); ok {
		if val, found := ctx.GetValue(string(ident)); found {
			return ast.ParseTypeRef(val.TypeName()), true
		}
	}

	return ast.TypeRef{}, false
}

func isNumericType(name string) bool {
	switch name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "rune", "byte":
		return true
	}
	return false
}

func returnTypeError(
	funcName string,
	source ast.TypeRef,
	target ast.TypeRef,
) error {
	return fmt.Errorf(
		"func %s cannot return %s as %s",
		funcName,
		source.String(),
		target.String(),
	)
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
