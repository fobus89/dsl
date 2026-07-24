package ast

import "testing"

func TestParseRecursiveTypeRef(t *testing.T) {
	for _, input := range []string{
		"[][][][]int",
		"[2][3][]*User",
		"*[4][]string",
	} {
		ref := ParseTypeRef(input)
		if got := ref.String(); got != input {
			t.Fatalf("ParseTypeRef(%q).String() = %q", input, got)
		}
	}
}
