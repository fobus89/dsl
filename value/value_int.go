package value

func (t Type) IsNumber() bool {
	switch t.Value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	}

	return false
}

// ========== is int ========================
func (t Type) IsInteger() bool {
	switch t.Value.(type) {
	case int, int8, int16, int32, int64:
		return true
	}
	return false
}

func (t Type) IsInt8() bool {
	return Is[int8](t.Value)
}

func (t Type) IsInt16() bool {
	return Is[int16](t.Value)
}

func (t Type) IsInt32() bool {
	return Is[int32](t.Value)
}

func (t Type) IsInt64() bool {
	return Is[int64](t.Value)
}

func (t Type) IsInt() bool {
	return Is[int](t.Value)
}

// ========== to int ========================
func (t Type) ToInt8() (int8, bool) {
	return To[int8](t.Value)
}

func (t Type) ToInt16() (int16, bool) {
	return To[int16](t.Value)
}

func (t Type) ToInt32() (int32, bool) {
	return To[int32](t.Value)
}

func (t Type) ToInt64() (int64, bool) {
	return To[int64](t.Value)
}

func (t Type) ToInt() (int, bool) {
	return To[int](t.Value)
}

// ========== cast int ========================
func (t Type) CastInt8() (int8, bool) {
	return Cast[int8](t.Value)
}

func (t Type) CastInt16() (int16, bool) {
	return Cast[int16](t.Value)
}

func (t Type) CastInt32() (int32, bool) {
	return Cast[int32](t.Value)
}

func (t Type) CastInt64() (int64, bool) {
	return Cast[int64](t.Value)
}

func (t Type) CastInt() (int, bool) {
	return Cast[int](t.Value)
}
