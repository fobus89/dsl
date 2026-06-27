package value

// ========== is float ===================
func (t Type) IsFloat() bool {
	switch t.Value.(type) {
	case float32, float64:
		return true
	}
	return false
}

func (t Type) IsFloat32() bool {
	return Is[float32](t.Value)
}

func (t Type) IsFloat64() bool {
	return Is[float64](t.Value)
}

// ========== to float ===================
func (t Type) ToFloat32() (float32, bool) {
	return To[float32](t.Value)
}

func (t Type) ToFloat64() (float64, bool) {
	return To[float64](t.Value)
}

// ========== cast float ===================
func (t Type) CastFloat32() (float32, bool) {
	return Cast[float32](t.Value)
}

func (t Type) CastFloat64() (float64, bool) {
	v, ok := Cast[float64](t.Value)
	{
		if ok {
			return v, ok
		}
	}

	switch v := t.Value.(type) {
	case bool:

		if v {
			return 1, true
		}

		return 0, true
	}

	return 0, false
}
