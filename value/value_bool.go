package value

// ========== is float ===================
func (t Type) IsBool() bool {
	switch t.value.(type) {
	case bool:
		return true
	}
	return false
}

// ========== to float ===================
func (t Type) ToBool() (bool, bool) {
	return To[bool](t.value)
}

// ========== cast float ===================
func (t Type) CastBool() (bool, bool) {
	if v, ok := t.ToBool(); ok {
		return v, ok
	}

	if t.IsNil() || t.IsUndefined() {
		return false, true
	}

	switch {
	case t.IsNumber():
		return t.UnsafeCastFloat64() != 0, true
	case t.IsString():
		return t.value != "", true
	}

	return true, true
}

// ========== cast float ===================
func (t Type) UnsafeCastBool() bool {
	v, _ := t.CastBool()
	return v
}
