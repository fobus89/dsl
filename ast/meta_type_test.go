package ast

import "testing"

func TestMetaTypeKindsAndRelations(t *testing.T) {
	userDef := TypeDef{Name: "User", Kind: StructType}
	idDef := TypeDef{
		Name:       "ID",
		Kind:       AliasType,
		Underlying: TypeRef{Name: "int"},
	}

	tests := []struct {
		name    string
		meta    MetaTypeValue
		kind    MetaTypeKind
		isPtr   bool
		matches []string
	}{
		{
			name: "primitive number",
			meta: NewMetaTypeValue(
				TypeRef{Name: "int"},
				nil,
			),
			kind: PrimitiveMetaType,
			matches: []string{
				MetaTypeName,
				PrimitiveTypeName,
				NumberTypeName,
				NamedTypeName,
			},
		},
		{
			name: "struct",
			meta: NewMetaTypeValue(
				TypeRef{Name: "User"},
				&userDef,
			),
			kind: StructMetaType,
			matches: []string{
				MetaTypeName,
				StructTypeName,
				NamedTypeName,
			},
		},
		{
			name: "numeric alias",
			meta: NewMetaTypeValue(
				TypeRef{Name: "ID"},
				&idDef,
			),
			kind: AliasMetaType,
			matches: []string{
				MetaTypeName,
				AliasTypeName,
				PrimitiveTypeName,
				NumberTypeName,
			},
		},
		{
			name: "slice",
			meta: NewMetaTypeValue(
				TypeRef{
					Kind: SliceTypeRef,
					Elem: &TypeRef{Name: "int"},
				},
				nil,
			),
			kind:    SliceMetaType,
			matches: []string{MetaTypeName, SliceTypeName},
		},
		{
			name: "array",
			meta: NewMetaTypeValue(
				TypeRef{
					Kind: ArrayTypeRef,
					Len:  3,
					Elem: &TypeRef{Name: "int"},
				},
				nil,
			),
			kind:    ArrayMetaType,
			matches: []string{MetaTypeName, ArrayTypeName},
		},
		{
			name: "pointer",
			meta: NewMetaTypeValue(
				TypeRef{Name: "User", IsPtr: true},
				&userDef,
			),
			kind:  StructMetaType,
			isPtr: true,
			matches: []string{
				MetaTypeName,
				StructTypeName,
				NamedTypeName,
			},
		},
		{
			name: "function",
			meta: NewFuncMetaTypeValue("build"),
			kind: FuncMetaType,
			matches: []string{
				MetaTypeName,
				FuncTypeName,
				NamedTypeName,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.meta.Kind != test.kind {
				t.Fatalf(
					"kind = %q, want %q",
					test.meta.Kind,
					test.kind,
				)
			}
			if test.meta.IsPtr != test.isPtr {
				t.Errorf(
					"isptr = %t, want %t",
					test.meta.IsPtr,
					test.isPtr,
				)
			}
			for _, expected := range test.matches {
				if !test.meta.Matches(expected) {
					t.Errorf("does not match %s", expected)
				}
			}
		})
	}
}

func TestPointerKeepsUnderlyingMetaType(t *testing.T) {
	meta := NewMetaTypeValue(
		TypeRef{Name: "int", IsPtr: true},
		nil,
	)

	if meta.Kind != PrimitiveMetaType {
		t.Fatalf("kind = %q, want primitivetype", meta.Kind)
	}
	if !meta.IsPtr {
		t.Fatal("isptr = false, want true")
	}
	if !meta.Matches(NumberTypeName) {
		t.Fatal("*int must match numbertype")
	}
}

func TestAllMetaTypeNamesAreRecognized(t *testing.T) {
	for _, name := range []string{
		MetaTypeName,
		PrimitiveTypeName,
		NumberTypeName,
		StringTypeName,
		BoolTypeName,
		SliceTypeName,
		ArrayTypeName,
		StructTypeName,
		AliasTypeName,
		NamedTypeName,
		MapTypeName,
		TupleTypeName,
		EnumTypeName,
		InterfaceTypeName,
		FuncTypeName,
		AnyTypeName,
		VoidTypeName,
		NeverTypeName,
	} {
		if !IsMetaTypeName(name) {
			t.Errorf("%s is not recognized as a meta type", name)
		}
	}
}

func TestBuiltInMetaTypeKinds(t *testing.T) {
	for _, test := range []struct {
		name string
		kind MetaTypeKind
	}{
		{name: "map", kind: MapMetaType},
		{name: "tuple", kind: TupleMetaType},
		{name: "struct", kind: StructMetaType},
		{name: "enum", kind: EnumMetaType},
		{name: "interface", kind: InterfaceMetaType},
		{name: "any", kind: AnyMetaType},
		{name: "void", kind: VoidMetaType},
		{name: "never", kind: NeverMetaType},
		{name: "string", kind: PrimitiveMetaType},
		{name: "bool", kind: PrimitiveMetaType},
	} {
		meta := NewMetaTypeValue(TypeRef{Name: test.name}, nil)
		if meta.Kind != test.kind {
			t.Errorf(
				"%s kind = %q, want %q",
				test.name,
				meta.Kind,
				test.kind,
			)
		}
	}
}
