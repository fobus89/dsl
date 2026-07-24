package funcdecl_parser_test

import (
	"strings"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	call_parser "github.com/fobus89/dsl/syntax/call"
	collection_parser "github.com/fobus89/dsl/syntax/collection"
	funcdecl_parser "github.com/fobus89/dsl/syntax/func_decl"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	map_parser "github.com/fobus89/dsl/syntax/map"
	member_parser "github.com/fobus89/dsl/syntax/member"
	typedecl_parser "github.com/fobus89/dsl/syntax/type_decl"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newFuncDeclTestParser(input string) testParser {
	p := parser.NewParser(input)
	literal_parser.RegisterParser(p)
	assignment_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	call_parser.RegisterParser(p)
	map_parser.RegisterParser(p)
	member_parser.RegisterParser(p)
	collection_parser.RegisterParser(p)
	typedecl_parser.RegisterParser(p)
	funcdecl_parser.RegisterParser(p)
	return p
}

func TestCollectionReturnTypeIsChecked(t *testing.T) {
	p := newFuncDeclTestParser(`
		fn items() []int {
			return []string{"wrong"}
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	_, err = exprs[0].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(err.Error(), "cannot return []string as []int") {
		t.Fatalf("expected collection return type error, got %v", err)
	}
}

func TestMetaTypeParameterAndReturn(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		comptime fn identity(t: type) type {
			return t
		}
		userType = identity(User)
		intType = identity(int)
		sliceType = identity([][]int)
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	decl := exprs[1].(*funcdecl_parser.FuncDecl)
	if got := decl.Params[0].Type.String(); got != "type" {
		t.Fatalf("parameter type = %q, want type", got)
	}
	if got := decl.ReturnType.String(); got != "type" {
		t.Fatalf("return type = %q, want type", got)
	}

	for _, expr := range exprs {
		if _, err := expr.Eval(p.Ctx()); err != nil {
			t.Fatal(err)
		}
	}

	userType, _ := p.Ctx().GetValue("userType")
	userMeta, ok := userType.Any().(ast.MetaTypeValue)
	if !ok || userMeta.Ref.String() != "User" ||
		userMeta.Def == nil ||
		userMeta.Def.Kind != ast.StructType {
		t.Fatalf("unexpected User meta type: %#v", userType.Any())
	}

	intType, _ := p.Ctx().GetValue("intType")
	intMeta, ok := intType.Any().(ast.MetaTypeValue)
	if !ok || intMeta.Ref.String() != "int" || intMeta.Def != nil {
		t.Fatalf("unexpected int meta type: %#v", intType.Any())
	}

	sliceType, _ := p.Ctx().GetValue("sliceType")
	sliceMeta, ok := sliceType.Any().(ast.MetaTypeValue)
	if !ok || sliceMeta.Ref.String() != "[][]int" ||
		sliceMeta.Ref.Kind != ast.SliceTypeRef {
		t.Fatalf("unexpected slice meta type: %#v", sliceType.Any())
	}

	generated, err := decl.PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if generated != "" {
		t.Fatalf("comptime function generated runtime code: %q", generated)
	}
}

func TestMetaTypeParameterRejectsRuntimeValue(t *testing.T) {
	p := newFuncDeclTestParser(`
		comptime fn inspect(t: type) type { return t }
		result = inspect(42)
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	_, err = exprs[1].Eval(p.Ctx())
	if err == nil || !strings.Contains(err.Error(), "expects type") {
		t.Fatalf("expected meta type argument error, got %v", err)
	}
}

func TestRuntimeFunctionRejectsMetaTypes(t *testing.T) {
	p := newFuncDeclTestParser(`
		fn inspect(t: structtype) type { return t }
	`)

	_, err := p.Parse()
	if err == nil ||
		!strings.Contains(err.Error(), "must be declared comptime") {
		t.Fatalf("expected runtime meta type declaration error, got %v", err)
	}
}

func TestComptimeMetaReceiverMethod(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		comptime fn (st: structtype) pointer() bool {
			return st.isptr
		}
		valueResult = User.pointer()
		pointerResult = (*User).pointer()
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	for _, expr := range exprs {
		if _, err := expr.Eval(p.Ctx()); err != nil {
			t.Fatal(err)
		}
	}

	valueResult, _ := p.Ctx().GetValue("valueResult")
	if valueResult.UnsafeCastBool() {
		t.Fatal("User.pointer() = true, want false")
	}
	pointerResult, _ := p.Ctx().GetValue("pointerResult")
	if !pointerResult.UnsafeCastBool() {
		t.Fatal("(*User).pointer() = false, want true")
	}
}

func TestSpecializedMetaTypeParameters(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		type ID int

		fn sample() int { return 1 }
		comptime fn requireStruct(t: structtype) type { return t }
		comptime fn requireSlice(t: slicetype) type { return t }
		comptime fn requireArray(t: arraytype) type { return t }
		comptime fn requireAlias(t: aliastype) type { return t }
		comptime fn requirePrimitive(t: primitivetype) type { return t }
		comptime fn requireNumber(t: numbertype) type { return t }
		comptime fn requireFunc(t: functype) type { return t }
		comptime fn isPointer(t: type) bool { return t.isptr }

		structResult = requireStruct(User)
		sliceResult = requireSlice([]int)
		arrayResult = requireArray([2]int)
		pointerStructResult = requireStruct(*User)
		pointerNumberResult = requireNumber(*int)
		aliasResult = requireAlias(ID)
		primitiveResult = requirePrimitive(string)
		numberResult = requireNumber(ID)
		funcResult = requireFunc(sample)
		structIsPointer = isPointer(*User)
		structIsValue = isPointer(User)
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	for _, expr := range exprs {
		if _, err := expr.Eval(p.Ctx()); err != nil {
			t.Fatal(err)
		}
	}

	expectedKinds := map[string]ast.MetaTypeKind{
		"structResult":        ast.StructMetaType,
		"sliceResult":         ast.SliceMetaType,
		"arrayResult":         ast.ArrayMetaType,
		"pointerStructResult": ast.StructMetaType,
		"pointerNumberResult": ast.PrimitiveMetaType,
		"aliasResult":         ast.AliasMetaType,
		"primitiveResult":     ast.PrimitiveMetaType,
		"numberResult":        ast.AliasMetaType,
		"funcResult":          ast.FuncMetaType,
	}
	for name, expected := range expectedKinds {
		result, ok := p.Ctx().GetValue(name)
		if !ok {
			t.Fatalf("%s not found", name)
		}
		meta, ok := result.Any().(ast.MetaTypeValue)
		if !ok || meta.Kind != expected {
			t.Errorf(
				"%s = %#v, want kind %s",
				name,
				result.Any(),
				expected,
			)
		}
	}

	for _, name := range []string{
		"pointerStructResult",
		"pointerNumberResult",
	} {
		result, _ := p.Ctx().GetValue(name)
		meta := result.Any().(ast.MetaTypeValue)
		if !meta.IsPtr {
			t.Errorf("%s.isptr = false, want true", name)
		}
	}

	structIsPointer, _ := p.Ctx().GetValue("structIsPointer")
	if !structIsPointer.UnsafeCastBool() {
		t.Fatal("(*User).isptr = false, want true")
	}
	structIsValue, _ := p.Ctx().GetValue("structIsValue")
	if structIsValue.UnsafeCastBool() {
		t.Fatal("User.isptr = true, want false")
	}
}

func TestSpecializedMetaTypeRejectsDifferentKind(t *testing.T) {
	p := newFuncDeclTestParser(`
		comptime fn requireStruct(t: structtype) type { return t }
		result = requireStruct(int)
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	_, err = exprs[1].Eval(p.Ctx())
	if err == nil ||
		!strings.Contains(
			err.Error(),
			"expects structtype, got primitivetype",
		) {
		t.Fatalf("expected specialized meta type error, got %v", err)
	}
}

func TestParseFunctionDeclaration(t *testing.T) {
	p := newFuncDeclTestParser(`fn add(a: Int, b: Int) Int { return a + b }`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	decl, ok := exprs[0].(*funcdecl_parser.FuncDecl)
	if !ok {
		t.Fatalf("expected *FuncDecl, got %T", exprs[0])
	}
	if decl.IsMethod() {
		t.Fatal("ordinary function must not be a method")
	}
	if decl.Name != "add" {
		t.Fatalf("expected add, got %s", decl.Name)
	}
	if len(decl.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(decl.Params))
	}
	if decl.ReturnType == nil {
		t.Fatal("expected return type")
	}
	if len(decl.Body) != 1 {
		t.Fatalf("expected one body statement, got %d", len(decl.Body))
	}
}

func TestParseMethodDeclaration(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		fn (u: User) name() String { return u.name }
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	decl, ok := exprs[1].(*funcdecl_parser.FuncDecl)
	if !ok {
		t.Fatalf("expected *FuncDecl, got %T", exprs[1])
	}
	if !decl.IsMethod() {
		t.Fatal("declaration with receiver must be a method")
	}
	if decl.Recv.Name != "u" {
		t.Fatalf("expected receiver u, got %s", decl.Recv.Name)
	}
	if decl.Recv.Type.IsPtr {
		t.Fatal("expected value receiver")
	}
	if decl.Name != "name" {
		t.Fatalf("expected name, got %s", decl.Name)
	}
}

func TestParseAndPrintPointerMethodReceiver(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		fn (u: *User) Name() String { return u.name }
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	decl, ok := exprs[1].(*funcdecl_parser.FuncDecl)
	if !ok {
		t.Fatalf("expected *FuncDecl, got %T", exprs[1])
	}
	if !decl.Recv.Type.IsPtr {
		t.Fatal("expected pointer receiver")
	}

	got, err := decl.PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	want := "func (u *User) Name() string {\n\treturn u.name\n}"
	if got != want {
		t.Fatalf("PrintGO() = %q, want %q", got, want)
	}
}

func TestPointerParameterAndReturnType(t *testing.T) {
	p := newFuncDeclTestParser(`
		fn copy(value: *String) *String {
			return value
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	decl := exprs[0].(*funcdecl_parser.FuncDecl)
	if decl.Params[0].Type.Name != "String" ||
		!decl.Params[0].Type.IsPtr {
		t.Fatalf("unexpected parameter type: %#v", decl.Params[0].Type)
	}
	if decl.ReturnType.Name != "String" || !decl.ReturnType.IsPtr {
		t.Fatalf("unexpected return type: %#v", decl.ReturnType)
	}

	got, err := decl.PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	want := "func copy(value *string) *string {\n\treturn value\n}"
	if got != want {
		t.Fatalf("PrintGO() = %q, want %q", got, want)
	}
}

func TestReturnTypeRejectsNumberAsString(t *testing.T) {
	p := newFuncDeclTestParser(`
		fn bad() String {
			return 123
		}
		bad()
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := exprs[0].PrintGO(p.Ctx()); err == nil {
		t.Fatal("expected PrintGO return type error")
	} else if !strings.Contains(err.Error(), "cannot return int as String") {
		t.Fatalf("unexpected PrintGO error: %v", err)
	}

	if _, err := exprs[1].Eval(p.Ctx()); err == nil {
		t.Fatal("expected interpreter return type error")
	} else if !strings.Contains(err.Error(), "cannot return int64 as String") {
		t.Fatalf("unexpected interpreter error: %v", err)
	}
}

func TestCustomIntReturnTypeIsPreserved(t *testing.T) {
	p := newFuncDeclTestParser(`
		type Int int
		fn numb() Int {
			return 122
		}
		numb()
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	typeCode, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if typeCode != "type Int int" {
		t.Fatalf("generated type = %q", typeCode)
	}

	funcCode, err := exprs[1].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	want := "func numb() Int {\n\treturn 122\n}"
	if funcCode != want {
		t.Fatalf("PrintGO() = %q, want %q", funcCode, want)
	}

	result, err := exprs[2].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if result.TypeName() != "Int" {
		t.Fatalf("return type = %q, want Int", result.TypeName())
	}
	if result.UnsafeCastInt() != 122 {
		t.Fatalf("return value = %v, want 122", result.Any())
	}
}

func TestFunctionIsAvailableBeforeItsDeclaration(t *testing.T) {
	p := newFuncDeclTestParser(`
		add(2, 3)
		fn add(a: Int, b: Int) Int { return a + b }
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	result, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if got := result.UnsafeCastInt(); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestCallMethodThroughReceiver(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		user = User {name: "Bob"}
		user.name()
		fn (u: User) name() String { return u.name }
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := exprs[1].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	result, err := exprs[2].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if got := result.UnsafeCastString(); got != "Bob" {
		t.Fatalf("expected Bob, got %q", got)
	}
}

func TestMethodRejectsReceiverWithoutDeclaredType(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		plain = {name: "Bob"}
		plain.name()
		fn (u: User) name() String { return u.name }
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[1].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	_, err = exprs[2].Eval(p.Ctx())
	if err == nil {
		t.Fatal("expected method receiver type error")
	}
	if !strings.Contains(err.Error(), "map[string]any") {
		t.Fatalf("expected actual receiver type in error, got %v", err)
	}
}

func TestMethodRejectsDifferentDeclaredReceiverType(t *testing.T) {
	p := newFuncDeclTestParser(`
		type User struct { name: string }
		type Admin struct { name: string }
		admin = Admin {name: "Root"}
		admin.name()
		fn (u: User) name() String { return u.name }
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[2].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}

	_, err = exprs[3].Eval(p.Ctx())
	if err == nil {
		t.Fatal("expected method receiver type error")
	}
	if !strings.Contains(err.Error(), "type Admin") {
		t.Fatalf("expected Admin receiver type in error, got %v", err)
	}
}

func TestMethodRequiresDeclaredReceiverType(t *testing.T) {
	p := newFuncDeclTestParser(
		`fn (u: Missing) name() String { return u.name }`,
	)

	if _, err := p.Parse(); err == nil {
		t.Fatal("expected undeclared receiver type error")
	}
}

func TestMethodReceiverTypeCanBeDeclaredLater(t *testing.T) {
	p := newFuncDeclTestParser(`
		fn (u: User) name() String { return u.name }
		type User struct { name: string }
	`)

	if _, err := p.Parse(); err != nil {
		t.Fatal(err)
	}
}
