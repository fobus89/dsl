package value

// ========== is float ===================
func (t Type) IsFloat() bool {
	switch t.value.(type) {
	case float32, float64:
		return true
	}
	return false
}

func (t Type) IsFloat32() bool {
	return Is[float32](t.value)
}

func (t Type) IsFloat64() bool {
	return Is[float64](t.value)
}

// ========== to float ===================
func (t Type) ToFloat32() (float32, bool) {
	return To[float32](t.value)
}

func (t Type) ToFloat64() (float64, bool) {
	return To[float64](t.value)
}

// ========== cast float ===================
func (t Type) CastFloat32() (float32, bool) {
	v, ok := Cast[float32](t.value)
	{
		if ok {
			return v, ok
		}
	}

	switch v := t.value.(type) {
	case bool:

		if v {
			return 1, true
		}

		return 0, true
	}

	return 0, false
}

func (t Type) CastFloat64() (float64, bool) {
	v, ok := Cast[float64](t.value)
	{
		if ok {
			return v, ok
		}
	}

	switch v := t.value.(type) {
	case bool:

		if v {
			return 1, true
		}

		return 0, true
	}

	return 0, false
}

// ========== unsafe cast float ===================
func (t Type) UnsafeCastFloat64() float64 {
	v, _ := t.CastFloat64()
	return v
}

func (t Type) UnsafeCastFloat32() float32 {
	v, _ := t.CastFloat32()
	return v
}
