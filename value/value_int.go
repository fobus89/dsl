package value

func (t Type) IsNumber() bool {
	switch t.value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	}

	return false
}

// ========== is int ========================
func (t Type) IsInteger() bool {
	switch t.value.(type) {
	case int, int8, int16, int32, int64:
		return true
	}
	return false
}

func (t Type) IsInt8() bool {
	return Is[int8](t.value)
}

func (t Type) IsInt16() bool {
	return Is[int16](t.value)
}

func (t Type) IsInt32() bool {
	return Is[int32](t.value)
}

func (t Type) IsInt64() bool {
	return Is[int64](t.value)
}

func (t Type) IsInt() bool {
	return Is[int](t.value)
}

// ========== to int ========================
func (t Type) ToInt8() (int8, bool) {
	return To[int8](t.value)
}

func (t Type) ToInt16() (int16, bool) {
	return To[int16](t.value)
}

func (t Type) ToInt32() (int32, bool) {
	return To[int32](t.value)
}

func (t Type) ToInt64() (int64, bool) {
	return To[int64](t.value)
}

func (t Type) ToInt() (int, bool) {
	return To[int](t.value)
}

// ========== cast int ========================
func (t Type) CastInt8() (int8, bool) {
	v, ok := Cast[int8](t.value)
	{
		if ok {
			return v, true
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

func (t Type) CastInt16() (int16, bool) {
	v, ok := Cast[int16](t.value)
	{
		if ok {
			return v, true
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

func (t Type) CastInt32() (int32, bool) {
	v, ok := Cast[int32](t.value)
	{
		if ok {
			return v, true
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

func (t Type) CastInt64() (int64, bool) {
	v, ok := Cast[int64](t.value)
	{
		if ok {
			return v, true
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

func (t Type) CastInt() (int, bool) {
	v, ok := Cast[int](t.value)
	{
		if ok {
			return v, true
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

// ========== unsafe cast int ========================

func (t Type) UnsafeCastInt8() int8 {
	v, _ := t.CastInt8()
	return v
}

func (t Type) UnsafeCastInt16() int16 {
	v, _ := t.CastInt16()
	return v
}

func (t Type) UnsafeCastInt32() int32 {
	v, _ := t.CastInt32()
	return v
}

func (t Type) UnsafeCastInt64() int64 {
	v, _ := t.CastInt64()
	return v
}

func (t Type) UnsafeCastInt() int {
	v, _ := t.CastInt()
	return v
}
