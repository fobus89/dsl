package parser

import "github.com/fobus89/dsl/value"

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

type functype = func(...value.Type) (value.Type, error)

type scope struct {
	parent    *scope
	values    MapType[string, value.Type]
	functions MapType[string, functype]
}

func NewCtxWithParent(parent *scope) *scope {
	return &scope{
		parent:    parent,
		values:    MapType[string, value.Type]{},
		functions: MapType[string, functype]{},
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

	return nil, false
}
