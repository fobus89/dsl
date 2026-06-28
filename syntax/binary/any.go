package binary_parser

import "reflect"

// {id:1} any {id:1,name:"user 1"} true
// {id:1,name:"user 1"} any {id:1} false
func mapAny(left, right map[string]any) bool {

	for k, lv := range left {

		rv, ok := right[k]
		{
			if !ok {
				continue
			}
		}

		if reflect.DeepEqual(lv, rv) {
			return true
		}
	}

	return false
}

// 1 any {id:1,name:"user 1"} true
// 2 any {id:1,age:2} true
// 3 any {id:1,age:2} false
func valAnyMap(val any, m map[string]any) bool {
	for _, v := range m {
		if reflect.DeepEqual(val, v) {
			return true
		}
	}

	return false
}

// 1 any 1 true
// 1 any 2 false
func valAnyVal(val1, val2 any) bool {
	return reflect.DeepEqual(val1, val2)
}

func mapAnySlice(left map[string]any, right []map[string]any) bool {
	for _, m := range right {
		if mapAny(left, m) {
			return true
		}
	}

	return false
}

func mapSliceAny(left []map[string]any, right map[string]any) bool {
	for _, m := range left {
		if mapAny(m, right) {
			return true
		}
	}

	return false
}

func mapSliceAnySlice(left, right []map[string]any) bool {
	for _, l := range left {
		for _, r := range right {
			if mapAny(l, r) {
				return true
			}
		}
	}

	return false
}

func sliceAny[T comparable](left, right []T) bool {
	if len(left) == 0 || len(right) == 0 {
		return false
	}

	set := make(map[T]struct{}, len(right))

	for _, v := range right {
		set[v] = struct{}{}
	}

	for _, v := range left {
		if _, ok := set[v]; ok {
			return true
		}
	}

	return false
}
