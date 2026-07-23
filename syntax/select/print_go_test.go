package select_parser_test

import (
	"go/ast"
	"go/importer"
	goparser "go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

func TestPrintGOGeneratesStandaloneForLoop(t *testing.T) {
	p := newSelectTestParser(`
		select id, name as username
		from users
		where active and id > 1
		limit 2
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(generated, "Select(") {
		t.Fatalf("generated code still depends on Select helper:\n%s", generated)
	}
	for _, expected := range []string{
		"for _, _value := range _rows",
		`_get(_row, "active")`,
		`_get(_row, "id")`,
		`_out["username"]`,
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated code does not contain %q:\n%s", expected, generated)
		}
	}

	source := `package generated
import (
	"fmt"
	"math"
	"reflect"
)
func run(users []any) any {
	return ` + generated + `
}`

	files := token.NewFileSet()
	file, err := goparser.ParseFile(files, "generated.go", source, 0)
	if err != nil {
		t.Fatalf("generated invalid Go syntax: %v\n%s", err, source)
	}

	config := types.Config{Importer: importer.Default()}
	if _, err := config.Check("generated", files, []*ast.File{file}, nil); err != nil {
		t.Fatalf("generated Go does not type-check: %v\n%s", err, source)
	}
}
