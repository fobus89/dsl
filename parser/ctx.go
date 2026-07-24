package parser

import (
	"strings"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/value"
)

type MapType[T comparable, E any] map[T]E

func (m MapType[T, E]) Get(key T) (E, bool) {
	tok, ok := m[key]

	if !ok {
		var none E
		return none, false
	}

	return tok, true
}

func (m MapType[T, E]) Set(key T, val E) {
	m[key] = val
}

type functype = ast.Func

type scope struct {
	parent    *scope
	values    MapType[string, value.Type]
	functions MapType[string, functype]
	types     MapType[string, ast.TypeDef]
	methods   MapType[string, MapType[string, functype]]
	generics  MapType[string, []ast.TypeParam]
	instances MapType[string, [][]ast.TypeRef]
}

func NewCtxWithParent(parent *scope) *scope {
	return &scope{
		parent:    parent,
		values:    MapType[string, value.Type]{},
		functions: MapType[string, functype]{},
		types:     MapType[string, ast.TypeDef]{},
		methods:   MapType[string, MapType[string, functype]]{},
		generics:  MapType[string, []ast.TypeParam]{},
		instances: MapType[string, [][]ast.TypeRef]{},
	}
}

func (s *scope) RegisterGenericInstance(
	key string,
	args []ast.TypeRef,
) {
	instances, _ := s.instances.Get(key)
	signature := genericArgsSignature(args)
	for _, existing := range instances {
		if genericArgsSignature(existing) == signature {
			return
		}
	}
	copyArgs := append([]ast.TypeRef(nil), args...)
	s.instances.Set(key, append(instances, copyArgs))
}

func (s *scope) GetGenericInstances(
	key string,
) [][]ast.TypeRef {
	if instances, ok := s.instances.Get(key); ok {
		return instances
	}
	if s.parent != nil {
		return s.parent.GetGenericInstances(key)
	}
	return nil
}

func genericArgsSignature(args []ast.TypeRef) string {
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		parts = append(parts, arg.String())
	}
	return strings.Join(parts, ",")
}

func (s *scope) SetGeneric(
	key string,
	params []ast.TypeParam,
) {
	s.generics.Set(key, params)
}

func (s *scope) GetGeneric(
	key string,
) ([]ast.TypeParam, bool) {
	if params, ok := s.generics.Get(key); ok {
		return params, true
	}
	if s.parent != nil {
		return s.parent.GetGeneric(key)
	}
	return nil, false
}

func NewCtx() *scope {
	return NewCtxWithParent(nil)
}

func (s *scope) SetValue(key string, val value.Type) {
	s.values.Set(key, val)
}

func (s *scope) GetValue(key string) (value.Type, bool) {

	v, ok := s.values.Get(key)
	{
		if ok {
			return v, true
		}
	}

	if s.parent != nil {
		return s.parent.GetValue(key)
	}

	return value.NewTypeNil(), false
}

func (s *scope) SetFunc(key string, fn functype) {
	s.functions.Set(key, fn)
}

func (s *scope) GetFunc(key string) (functype, bool) {

	v, ok := s.functions.Get(key)
	{
		if ok {
			return v, true
		}
	}

	if s.parent != nil {
		return s.parent.GetFunc(key)
	}

	return nil, false
}

func (s *scope) SetType(key string, def ast.TypeDef) {
	s.types.Set(key, def)
}

func (s *scope) GetType(key string) (ast.TypeDef, bool) {
	if def, ok := s.types.Get(key); ok {
		return def, true
	}

	if s.parent != nil {
		return s.parent.GetType(key)
	}

	return ast.TypeDef{}, false
}

func (s *scope) SetMethod(typeName, methodName string, fn functype) {
	typeMethods, ok := s.methods.Get(typeName)
	if !ok {
		typeMethods = MapType[string, functype]{}
		s.methods.Set(typeName, typeMethods)
	}
	typeMethods.Set(methodName, fn)
}

func (s *scope) GetMethod(typeName, methodName string) (functype, bool) {
	if typeMethods, ok := s.methods.Get(typeName); ok {
		if fn, ok := typeMethods.Get(methodName); ok {
			return fn, true
		}
	}

	if s.parent != nil {
		return s.parent.GetMethod(typeName, methodName)
	}

	return nil, false
}

func (s *scope) GetLocalCtx() ast.Ctx {
	return NewCtxWithParent(s)
}
