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

	switch {
	case t.IsNumber():
		return t.value != 0, true
	case t.IsString():
		return t.value != "", true
	}

	return false, false
}

// ========== cast float ===================
func (t Type) UnsafeCastBool() bool {
	v, _ := t.CastBool()
	return v
}
