package parser

import (
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
}

func NewCtxWithParent(parent *scope) *scope {
	return &scope{
		parent:    parent,
		values:    MapType[string, value.Type]{},
		functions: MapType[string, functype]{},
		types:     MapType[string, ast.TypeDef]{},
		methods:   MapType[string, MapType[string, functype]]{},
	}
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
