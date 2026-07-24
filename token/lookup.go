package token

import (
	"slices"
)

type MapType[T comparable, E any] map[T]E

func (m MapType[T, E]) Get(key T) (E, bool) {
	tok, ok := m[key]

	if !ok {
		var none E
		return none, false
	}

	return tok, true
}

var symbolKeys = func() []string {
	keys := make([]string, 0, len(symbolMap))
	{
		for key := range symbolMap {
			keys = append(keys, key)
		}
	}

	slices.SortFunc(keys, func(a, b string) int {
		return len(b) - len(a)
	})

	return keys
}()

var symbolMap = func() MapType[string, TokenType] {
	m := make(MapType[string, TokenType])

	for i := symbol_start + 1; i < symbol_end; i++ {

		// Skip internal markers
		if slices.Contains(markers, i) {
			continue
		}
		m[i.String()] = i
	}

	return m
}()

var keywordMap = func() MapType[string, TokenType] {
	m := make(MapType[string, TokenType])

	for i := keyword_start + 1; i < keyword_end; i++ {
		// Skip internal markers
		if slices.Contains(markers, i) {
			continue
		}

		m[i.String()] = i
	}

	return m
}()

var tokenToStringMap = func() MapType[TokenType, string] {
	m := make(MapType[TokenType, string])

	// Add all tokens from token_start to token_end
	for i := token_start + 1; i < token_end; i++ {

		// Skip internal markers
		if slices.Contains(markers, i) {
			continue
		}

		m[i] = i.String()
	}

	return m
}()

var stringToTokenMap = func() MapType[string, TokenType] {
	m := make(MapType[string, TokenType])

	// Reverse mapping from tokenToStringMap
	for token, str := range tokenToStringMap {
		m[str] = token
	}

	return m
}()

func StringToToken(key string) TokenType {
	token, _ := stringToTokenMap.Get(key)

	return token
}

func Symbol(key string) (TokenType, bool) {
	if keyword, ok := symbolMap.Get(key); ok {
		return keyword, true
	}
	return ILLEGAL, false
}

func Keyword(key string) (TokenType, bool) {
	if keyword, ok := keywordMap.Get(key); ok {
		return keyword, true
	}
	return ILLEGAL, false
}

func LookupReservedToken(key string) (TokenType, bool) {
	if keyword, ok := keywordMap.Get(key); ok {
		return keyword, true
	}

	for _, bounds := range [][2]TokenType{
		{builtin_start, builtin_end},
		{compiletime_start, compiletime_end},
		{type_start, type_end},
	} {
		for kind := bounds[0] + 1; kind < bounds[1]; kind++ {
			if MarkersContains(kind) || kind.String() != key {
				continue
			}
			// Collections are expressed as []T and [N]T. The old
			// array/slice pseudo-types remain identifiers.
			if kind == Array || kind == Slice {
				return ILLEGAL, false
			}
			return kind, true
		}
	}
	return ILLEGAL, false
}

func Symbols() []string {
	return symbolKeys
}
