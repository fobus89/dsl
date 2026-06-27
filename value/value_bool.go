package value

// ========== is float ===================
func (t Type) IsBool() bool {
	switch t.Value.(type) {
	case bool:
		return true
	}
	return false
}

// ========== to float ===================
func (t Type) ToBool() (bool, bool) {
	return To[bool](t.Value)
}

// ========== cast float ===================
func (t Type) CastBool() (bool, bool) {
	if v, ok := t.ToBool(); ok {
		return v, ok
	}

	switch {
	case t.IsNumber():
		return t.Value != 0, true
	case t.IsString():
		return t.Value != "", true
	}

	return false, false
}
