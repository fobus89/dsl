package value

type Undefined struct{}

func NewUndefined() Undefined {
	return Undefined{}
}

func (Undefined) String() string {
	return "undefined"
}

func (Undefined) MarshalJSON() ([]byte, error) {
	return []byte(`"undefined"`), nil
}

func NewTypeUndefined() Type {
	return NewType(NewUndefined())
}

func (t Type) IsUndefined() bool {
	_, ok := t.value.(Undefined)
	return ok
}
