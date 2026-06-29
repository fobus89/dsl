package select_parser_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	call_parser "github.com/fobus89/dsl/syntax/call"
	comparison_parser "github.com/fobus89/dsl/syntax/comparison"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	logical_parser "github.com/fobus89/dsl/syntax/logical"
	member_parser "github.com/fobus89/dsl/syntax/member"
	select_parser "github.com/fobus89/dsl/syntax/select"
	unary_parser "github.com/fobus89/dsl/syntax/unary"
	"github.com/fobus89/dsl/value"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newSelectTestParser(input string) testParser {
	p := parser.NewParser(input)

	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	comparison_parser.RegisterParser(p)
	assignment_parser.RegisterParser(p)
	call_parser.RegisterParser(p)
	member_parser.RegisterParser(p)
	select_parser.RegisterParser(p)
	unary_parser.RegisterParser(p)
	logical_parser.RegisterParser(p)

	return p
}

func TestSelectMemberField(t *testing.T) {
	p := newSelectTestParser(`
		select id, address.street from users
	`)

	p.Ctx().SetValue("users", value.NewType([]any{
		map[string]any{
			"id": 1,
			"address": map[string]any{
				"street": "Victor Plains",
			},
		},
	}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	rows := got.Any().([]map[string]any)
	if rows[0]["street"] != "Victor Plains" {
		t.Fatalf("expected nested member in output, got %#v", rows[0])
	}
}

func TestSelectExpressionFieldWithAlias(t *testing.T) {
	p := newSelectTestParser(`
		select id as pin, 1+2 as sum, address.geo from users
	`)

	p.Ctx().SetValue("users", value.NewType(map[string]any{
		"id": 1,
		"address": map[string]any{
			"geo": map[string]any{
				"lat": "-37.3159",
				"lng": "81.1496",
			},
		},
	}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	row := got.Any().(map[string]any)
	if row["pin"] != 1 {
		t.Fatalf("expected pin alias, got %#v", row)
	}

	if row["sum"] != float64(3) {
		t.Fatalf("expected expression alias sum, got %#v", row)
	}

	if row["geo"] == nil {
		t.Fatalf("expected geo member, got %#v", row)
	}
}

func TestSelectStarMap(t *testing.T) {
	p := newSelectTestParser(`
		select * from users
	`)

	p.Ctx().SetValue("users", value.NewType(map[string]any{
		"id":       1,
		"name":     "Leanne Graham",
		"username": "Bret",
	}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	row := got.Any().(map[string]any)
	if row["id"] != 1 || row["name"] != "Leanne Graham" || row["username"] != "Bret" {
		t.Fatalf("expected full row, got %#v", row)
	}
}

func TestSelectStarSliceWhereLimit(t *testing.T) {
	p := newSelectTestParser(`
		select * from users where active == true limit 1
	`)

	p.Ctx().SetValue("users", value.NewType([]any{
		map[string]any{
			"id":     1,
			"name":   "Leanne Graham",
			"active": true,
		},
		map[string]any{
			"id":     2,
			"name":   "Ervin Howell",
			"active": true,
		},
		map[string]any{
			"id":     3,
			"name":   "Clementine Bauch",
			"active": false,
		},
	}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	rows := got.Any().([]map[string]any)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %#v", rows)
	}

	if rows[0]["id"] != 1 || rows[0]["name"] != "Leanne Graham" || rows[0]["active"] != true {
		t.Fatalf("expected full first active row, got %#v", rows[0])
	}
}

func TestSelectDeepMemberField(t *testing.T) {
	p := newSelectTestParser(`
		user1 = select
			id,name,username,
			address.street.geo.lat.lng
		from json(get("https://jsonplaceholder.typicode.com/users/"))
	`)

	p.Ctx().SetFunc("get", func(vals ...value.Type) (value.Type, error) {
		return value.NewType(`[{
			"id": 1,
			"name": "Leanne Graham",
			"username": "Bret",
			"address": {
				"street": {
					"geo": {
						"lat": {
							"lng": "Kulas Light"
						}
					}
				}
			}
		}]`), nil
	})

	p.Ctx().SetFunc("json", func(vals ...value.Type) (value.Type, error) {
		if len(vals) != 1 {
			return value.NewTypeNil(), fmt.Errorf("json() expects exactly 1 argument")
		}

		var data any
		if err := json.Unmarshal([]byte(vals[0].UnsafeCastString()), &data); err != nil {
			return value.NewTypeNil(), err
		}

		return value.NewType(data), nil
	})

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range exprs {
		_, err := v.Eval(p.Ctx())
		if err != nil {
			t.Fatal(err)
		}
	}

	got, ok := p.Ctx().GetValue("user1")
	if !ok {
		t.Fatal("expected user1 in context")
	}

	rows, ok := got.Any().([]map[string]any)
	if !ok {
		t.Fatalf("expected select output, got %#v", got.Any())
	}

	if rows[0]["lng"] != "Kulas Light" {
		t.Fatalf("expected deep nested member in output, got %#v", rows[0])
	}

}

func TestSelectWhereMember(t *testing.T) {
	p := newSelectTestParser(`
		select name as username from users where address.city == "Gwenborough"
	`)

	p.Ctx().SetValue("users", value.NewType([]any{
		map[string]any{
			"name": "Leanne Graham",
			"address": map[string]any{
				"city": "Gwenborough",
			},
		},
		map[string]any{
			"name": "Ervin Howell",
			"address": map[string]any{
				"city": "Wisokyburgh",
			},
		},
	}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	rows := got.Any().([]map[string]any)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %#v", rows)
	}

	if rows[0]["username"] != "Leanne Graham" {
		t.Fatalf("expected aliased field, got %#v", rows[0])
	}
}

func TestSelectMissingDeepMemberReturnsNil(t *testing.T) {
	p := newSelectTestParser(`
		select id, address.street.zipcode from users
	`)

	p.Ctx().SetValue("users", value.NewType([]any{
		map[string]any{
			"id": 1,
			"address": map[string]any{
				"street": "Kulas Light",
			},
		},
	}))

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got, err := exprs[0].Eval(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}

	rows := got.Any().([]map[string]any)
	if rows[0]["zipcode"] != nil {
		t.Fatalf("expected invalid nested member to be nil, got %#v", rows[0])
	}
}
