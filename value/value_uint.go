package value

// ========== is uint ===================
func (t Type) IsUnsignedInteger() bool {
	switch t.value.(type) {
	case uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func (t Type) IsUint8() bool {
	return Is[uint8](t.value)
}

func (t Type) IsUint16() bool {
	return Is[uint16](t.value)
}

func (t Type) IsUint32() bool {
	return Is[uint32](t.value)
}

func (t Type) IsUint64() bool {
	return Is[uint64](t.value)
}

func (t Type) IsUint() bool {
	return Is[uint](t.value)
}

// ========== to uint ====================
func (t Type) ToUint8() (uint8, bool) {
	return To[uint8](t.value)
}

func (t Type) ToUint16() (uint16, bool) {
	return To[uint16](t.value)
}

func (t Type) ToUint32() (uint32, bool) {
	return To[uint32](t.value)
}

func (t Type) ToUint64() (uint64, bool) {
	return To[uint64](t.value)
}

func (t Type) ToUint() (uint, bool) {
	return To[uint](t.value)
}

// ========== cast uint ===================
func (t Type) CastUint8() (uint8, bool) {
	v, ok := Cast[uint8](t.value)
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

func (t Type) CastUint16() (uint16, bool) {
	v, ok := Cast[uint16](t.value)
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

func (t Type) CastUint32() (uint32, bool) {
	v, ok := Cast[uint32](t.value)
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

func (t Type) CastUint64() (uint64, bool) {
	v, ok := Cast[uint64](t.value)
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

func (t Type) CastUint() (uint, bool) {
	v, ok := Cast[uint](t.value)
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

// ========== cast uint ===================
func (t Type) UnsafeCastUint8() uint8 {
	v, _ := t.CastUint8()
	return v
}

func (t Type) UnsafeCastUint16() uint16 {
	v, _ := t.CastUint16()
	return v
}

func (t Type) UnsafeCastUint32() uint32 {
	v, _ := t.CastUint32()
	return v
}

func (t Type) UnsafeCastUint64() uint64 {
	v, _ := t.CastUint64()
	return v
}

func (t Type) UnsafeCastUint() uint {
	v, _ := t.CastUint()
	return v
}
