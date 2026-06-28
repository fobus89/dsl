package binary_parser

import "reflect"

func sliceAll[T comparable](left, right []T) bool {
	if len(left) == 0 {
		return true
	}

	set := make(map[T]struct{}, len(right))

	for _, v := range right {
		set[v] = struct{}{}
	}

	for _, v := range left {
		if _, ok := set[v]; !ok {
			return false
		}
	}

	return true
}

func mapAll(left, right map[string]any) bool {
	for k, lv := range left {
		rv, ok := right[k]
		if !ok {
			return false
		}

		if !reflect.DeepEqual(lv, rv) {
			return false
		}
	}

	return true
}

func valAllMap(val any, m map[string]any) bool {
	for _, v := range m {
		if !reflect.DeepEqual(val, v) {
			return false
		}
	}

	return true
}

func mapAllSlice(left map[string]any, right []map[string]any) bool {
	for _, m := range right {
		if !mapAll(left, m) {
			return false
		}
	}

	return true
}

func mapSliceAll(left []map[string]any, right map[string]any) bool {
	for _, m := range left {
		if !mapAll(m, right) {
			return false
		}
	}

	return true
}

func mapSliceAllSlice(left, right []map[string]any) bool {
	for _, l := range left {
		found := false

		for _, r := range right {
			if mapAll(l, r) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}
