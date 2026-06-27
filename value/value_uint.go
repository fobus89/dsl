package value

// ========== is uint ===================
func (t Type) IsUnsignedInteger() bool {
	switch t.Value.(type) {
	case uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func (t Type) IsUint8() bool {
	return Is[uint8](t.Value)
}

func (t Type) IsUint16() bool {
	return Is[uint16](t.Value)
}

func (t Type) IsUint32() bool {
	return Is[uint32](t.Value)
}

func (t Type) IsUint64() bool {
	return Is[uint64](t.Value)
}

func (t Type) IsUint() bool {
	return Is[uint](t.Value)
}

// ========== cast uint ===================
func (t Type) CastUint8() (uint8, bool) {
	return Cast[uint8](t.Value)
}

func (t Type) CastUint16() (uint16, bool) {
	return Cast[uint16](t.Value)
}

func (t Type) CastUint32() (uint32, bool) {
	return Cast[uint32](t.Value)
}

func (t Type) CastUint64() (uint64, bool) {
	return Cast[uint64](t.Value)
}

func (t Type) CastUint() (uint, bool) {
	return Cast[uint](t.Value)
}

// ========== to uint ====================
func (t Type) ToUint8() (uint8, bool) {
	return To[uint8](t.Value)
}

func (t Type) ToUint16() (uint16, bool) {
	return To[uint16](t.Value)
}

func (t Type) ToUint32() (uint32, bool) {
	return To[uint32](t.Value)
}

func (t Type) ToUint64() (uint64, bool) {
	return To[uint64](t.Value)
}

func (t Type) ToUint() (uint, bool) {
	return To[uint](t.Value)
}
