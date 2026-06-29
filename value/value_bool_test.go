package value

import "testing"

func TestCastBoolFalsyValues(t *testing.T) {
	tests := []Type{
		NewType(0),
		NewType(-0),
		NewType(int64(0)),
		NewType(float64(0)),
		NewType(float64(-0)),
		NewTypeNil(),
		NewTypeUndefined(),
	}

	for _, tt := range tests {
		got, ok := tt.CastBool()
		if !ok {
			t.Fatalf("expected bool cast to be supported for %#v", tt.Any())
		}

		if got {
			t.Fatalf("expected %#v to be falsy", tt.Any())
		}
	}
}

func TestCastBoolTruthyValues(t *testing.T) {
	tests := []Type{
		NewType(1),
		NewType(-1),
		NewType("hello"),
		NewType(map[string]any{}),
		NewType([]int{}),
	}

	for _, tt := range tests {
		got, ok := tt.CastBool()
		if !ok {
			t.Fatalf("expected bool cast to be supported for %#v", tt.Any())
		}

		if !got {
			t.Fatalf("expected %#v to be truthy", tt.Any())
		}
	}
}
