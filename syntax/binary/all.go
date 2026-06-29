package binary_parser

import "reflect"

func valAllSlice(val any, slice any) (any, bool) {
	return valAnySlice(val, slice)
}

func sliceAllVal(slice any, val any) (any, bool) {
	v := reflect.ValueOf(slice)
	for i := 0; i < v.Len(); i++ {
		if !reflect.DeepEqual(v.Index(i).Interface(), val) {
			return nil, false
		}
	}

	return slice, true
}

func sliceAll[T comparable](left, right []T) (any, bool) {
	if len(left) == 0 {
		return left, true
	}

	set := make(map[T]struct{}, len(right))

	for _, v := range right {
		set[v] = struct{}{}
	}

	for _, v := range left {
		if _, ok := set[v]; !ok {
			return nil, false
		}
	}

	return left, true
}

func mapAll(left, right map[string]any) (any, bool) {
	for k, lv := range left {
		rv, ok := right[k]
		if !ok {
			return nil, false
		}

		if !reflect.DeepEqual(lv, rv) {
			return nil, false
		}
	}

	return right, true
}

func valAllMap(val any, m map[string]any) (any, bool) {
	for _, v := range m {
		if !reflect.DeepEqual(val, v) {
			return nil, false
		}
	}

	return val, true
}

func mapAllSlice(left map[string]any, right []map[string]any) (any, bool) {
	var out []map[string]any

	for _, m := range right {
		if v, ok := mapAll(left, m); ok {
			out = append(out, v.(map[string]any))
		}
	}

	if len(out) == 0 {
		return nil, false
	}

	return out, true
}

func mapSliceAll(left []map[string]any, right map[string]any) (any, bool) {
	for _, m := range left {
		if _, ok := mapAll(m, right); !ok {
			return nil, false
		}
	}

	return right, true
}

func mapSliceAllSlice(left, right []map[string]any) (any, bool) {
	var out []map[string]any

	for _, l := range left {
		found := false

		for _, r := range right {
			if v, ok := mapAll(l, r); ok {
				out = append(out, v.(map[string]any))
				found = true
				break
			}
		}

		if !found {
			return nil, false
		}
	}

	return out, true
}
