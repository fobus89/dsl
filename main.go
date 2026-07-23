package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"

	"github.com/fobus89/dsl/parser"
	all_parser "github.com/fobus89/dsl/syntax/all"
	any_parser "github.com/fobus89/dsl/syntax/any"
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	call_parser "github.com/fobus89/dsl/syntax/call"
	comparison_parser "github.com/fobus89/dsl/syntax/comparison"
	funcdecl_parser "github.com/fobus89/dsl/syntax/func_decl"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	logical_parser "github.com/fobus89/dsl/syntax/logical"
	map_parser "github.com/fobus89/dsl/syntax/map"
	member_parser "github.com/fobus89/dsl/syntax/member"
	select_parser "github.com/fobus89/dsl/syntax/select"
	typedecl_parser "github.com/fobus89/dsl/syntax/type_decl"
	unary_parser "github.com/fobus89/dsl/syntax/unary"
	"github.com/fobus89/dsl/value"
)

func main() {
	p := parser.NewParser(`
		type User struct { name: *string = "hello" }
		fn (u: *User) Name() String { return u.name + (2121 * 2 /2 * (1+22)) }
		u = User{}
	`)

	p.SetValue("id", value.NewType(12211))

	slice1 := []int{11, 7}

	slices.Reverse(slice1)
	p.SetValue("testarray1", value.NewType(slice1))
	p.SetValue("testarray2", value.NewType([]int{4, 2, 3, 7, 5, 6, 1, 22}))

	p.SetFunc("get_template", func(vals ...value.Type) (value.Type, error) {
		if len(vals) != 1 {
			return value.NewTypeNil(), fmt.Errorf("len() expects exactly 1 argument, got %d", len(vals))
		}

		templ := value.NewType(map[string]any{
			"templ_id":  vals[0].Any(),
			"fio":       "eshmatov toshmat boltayevich",
			"pin":       "eshmat",
			"firstname": "eshmatov",
			"name":      "toshmat",
			"lastname":  "boltayevich",
			"age":       40,
		})

		return value.NewType(templ), nil
	})

	p.SetFunc("len", func(vals ...value.Type) (value.Type, error) {
		if len(vals) != 1 {
			return value.NewTypeNil(), fmt.Errorf("len() expects exactly 1 argument, got %d", len(vals))
		}

		return value.NewType(vals[0].Len()), nil
	})

	p.SetFunc("get", func(vals ...value.Type) (value.Type, error) {
		if len(vals) != 1 {
			return value.NewTypeNil(), fmt.Errorf("get() expects exactly 1 argument, got %d", len(vals))
		}

		url, ok := vals[0].CastString()
		{
			if !ok {
				return value.NewTypeNil(), errors.New("get() expects a string URL")
			}
		}

		resp, err := http.Get(url)
		{
			if err != nil {
				return value.NewTypeNil(), err
			}
			defer resp.Body.Close()
		}

		body, err := io.ReadAll(resp.Body)
		{
			if err != nil {
				return value.NewTypeNil(), err
			}
		}

		return value.NewType(string(body)), nil
	})

	p.SetFunc("json", func(vals ...value.Type) (value.Type, error) {
		if len(vals) != 1 {
			return value.NewTypeNil(), fmt.Errorf("get() expects exactly 1 argument, got %d", len(vals))
		}

		str, ok := vals[0].CastString()
		{
			if !ok {
				return value.NewTypeNil(), errors.New("get() expects a string URL")
			}
		}

		var m any

		if err := json.Unmarshal([]byte(str), &m); err != nil {
			return value.NewTypeNil(), err
		}

		return value.NewType(m), nil
	})

	p.SetFunc("min", func(vals ...value.Type) (value.Type, error) {
		minVal := vals[0].UnsafeCastFloat64()

		for _, v := range vals {
			tmp := v.UnsafeCastFloat64()

			if minVal > tmp {
				minVal = tmp
			}
		}

		return value.NewType(minVal), nil
	})

	p.SetFunc("stringify", func(vals ...value.Type) (value.Type, error) {
		if len(vals) != 1 {
			return value.NewTypeNil(), fmt.Errorf("get() expects exactly 1 argument, got %d", len(vals))
		}

		data, err := json.MarshalIndent(vals[0].Any(), "", " ")
		{
			if err != nil {
				return value.NewTypeNil(), err
			}
		}

		return value.NewType(string(data)), nil
	})

	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	comparison_parser.RegisterParser(p)
	any_parser.RegisterParser(p)
	all_parser.RegisterParser(p)
	assignment_parser.RegisterParser(p)
	call_parser.RegisterParser(p)
	funcdecl_parser.RegisterParser(p)
	map_parser.RegisterParser(p)
	member_parser.RegisterParser(p)
	select_parser.RegisterParser(p)
	typedecl_parser.RegisterParser(p)
	unary_parser.RegisterParser(p)
	logical_parser.RegisterParser(p)

	exprs, err := p.Parse()
	{
		if err != nil {
			log.Fatalln(err)
		}
	}

	for _, expr := range exprs {
		// v, err := expr.Eval(p.Ctx())
		// {
		// 	if err != nil {
		// 		fmt.Println(err)
		// 	} else if v.Any() != nil {
		// 		fmt.Println(v.Any())
		// 	}
		// }

		code, err := expr.PrintGO(p.Ctx())
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(code)
	}

	// res, _ := p.GetValue("result")

	// fmt.Println(res.UnsafeCastString())

}
