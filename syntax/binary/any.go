package binary_parser

import "reflect"

// {id:1} any {id:1,name:"user 1"} returns right
func mapAny(left, right map[string]any) (any, bool) {
	for k, lv := range left {
		rv, ok := right[k]
		{
			if !ok {
				return nil, false
			}
		}

		if !reflect.DeepEqual(lv, rv) {
			return nil, false
		}
	}

	return right, true
}

func valAnyMap(val any, m map[string]any) (any, bool) {
	for _, v := range m {
		if reflect.DeepEqual(val, v) {
			return v, true
		}
	}

	return nil, false
}

func valAnyVal(val1, val2 any) (any, bool) {
	if reflect.DeepEqual(val1, val2) {
		return val1, true
	}

	return nil, false
}

func valAnySlice(val any, slice any) (any, bool) {
	v := reflect.ValueOf(slice)
	for i := 0; i < v.Len(); i++ {
		if reflect.DeepEqual(val, v.Index(i).Interface()) {
			return val, true
		}
	}

	return nil, false
}

func sliceAnyVal(slice any, val any) (any, bool) {
	return valAnySlice(val, slice)
}

func mapAnySlice(left map[string]any, right []map[string]any) (any, bool) {
	for _, m := range right {
		if v, ok := mapAny(left, m); ok {
			return v, true
		}
	}

	return nil, false
}

func mapSliceAny(left []map[string]any, right map[string]any) (any, bool) {
	for _, m := range left {
		if v, ok := mapAny(m, right); ok {
			return v, true
		}
	}

	return nil, false
}

func mapSliceAnySlice(left, right []map[string]any) (any, bool) {
	for _, l := range left {
		for _, r := range right {
			if v, ok := mapAny(l, r); ok {
				return v, true
			}
		}
	}

	return nil, false
}

func sliceAny[T comparable](left, right []T) (any, bool) {
	if len(left) == 0 || len(right) == 0 {
		return nil, false
	}

	set := make(map[T]struct{}, len(right))

	for _, v := range right {
		set[v] = struct{}{}
	}

	for _, v := range left {
		if _, ok := set[v]; ok {
			return v, true
		}
	}

	return nil, false
}
