package value

import "strconv"

func (t Type) IsString() bool {
	return Is[string](t.value)
}

func (t Type) ToString() (string, bool) {
	return To[string](t.value)
}

func (t Type) CastString() (string, bool) {
	switch v := t.value.(type) {
	case string:
		return v, true

	case int:
		return strconv.Itoa(v), true
	case int8:
		return strconv.FormatInt(int64(v), 10), true
	case int16:
		return strconv.FormatInt(int64(v), 10), true
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int64:
		return strconv.FormatInt(v, 10), true

	case uint:
		return strconv.FormatUint(uint64(v), 10), true
	case uint8:
		return strconv.FormatUint(uint64(v), 10), true
	case uint16:
		return strconv.FormatUint(uint64(v), 10), true
	case uint32:
		return strconv.FormatUint(uint64(v), 10), true
	case uint64:
		return strconv.FormatUint(v, 10), true

	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), true
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(v), true
	}

	return "", false
}

func (t Type) UnsafeCastString() string {
	v, _ := t.CastString()
	return v
}
