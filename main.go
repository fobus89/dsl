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
	assignment_parser "github.com/fobus89/dsl/syntax/assignment"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	call_parser "github.com/fobus89/dsl/syntax/call"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	logical_parser "github.com/fobus89/dsl/syntax/logical"
	member_parser "github.com/fobus89/dsl/syntax/member"
	select_parser "github.com/fobus89/dsl/syntax/select"
	"github.com/fobus89/dsl/value"
)

func main() {
	p := parser.NewParser(`

		testarray1 any testarray2

		user1 = select
			id,name,username,
			address.street.zipcode
		from json(get("https://jsonplaceholder.typicode.com/users/"))

		user1

		user2 = select
			id as id2,
			name
		from json(get("https://jsonplaceholder.typicode.com/users/1"))

		stringify(user1)	
		`)

	slice1 := []int{11, 7}

	slices.Reverse(slice1)
	p.SetValue("testarray1", value.NewType(slice1))
	p.SetValue("testarray2", value.NewType([]int{4, 2, 3, 7, 5, 6, 1, 22}))

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
	assignment_parser.RegisterParser(p)
	call_parser.RegisterParser(p)
	member_parser.RegisterParser(p)
	select_parser.RegisterParser(p)
	logical_parser.RegisterParser(p)

	exprs, err := p.Parse()
	{
		if err != nil {
			log.Fatalln(err)
		}
	}

	for _, expr := range exprs {
		v, err := expr.Eval(p.Ctx())
		{
			if err != nil {
				fmt.Println(err)
			} else if v.Any() != nil {
				fmt.Println(v.Any())
			}
		}
	}
}
