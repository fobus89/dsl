package matchstmt_parser_test

import (
	astgo "go/ast"
	goparser "go/parser"
	gotoken "go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/fobus89/dsl/ast"
	"github.com/fobus89/dsl/parser"
	binary_parser "github.com/fobus89/dsl/syntax/binary"
	call_parser "github.com/fobus89/dsl/syntax/call"
	comparison_parser "github.com/fobus89/dsl/syntax/comparison"
	funcdecl_parser "github.com/fobus89/dsl/syntax/func_decl"
	let_parser "github.com/fobus89/dsl/syntax/let"
	literal_parser "github.com/fobus89/dsl/syntax/literal"
	logical_parser "github.com/fobus89/dsl/syntax/logical"
	matchstmt_parser "github.com/fobus89/dsl/syntax/match_stmt"
	member_parser "github.com/fobus89/dsl/syntax/member"
	typedecl_parser "github.com/fobus89/dsl/syntax/type_decl"
)

type testParser interface {
	parser.Parser
	Parse() ([]ast.Expr, error)
}

func newMatchParser(input string) testParser {
	p := parser.NewParser(input)
	literal_parser.RegisterParser(p)
	binary_parser.RegisterParser(p)
	comparison_parser.RegisterParser(p)
	logical_parser.RegisterParser(p)
	call_parser.RegisterParser(p)
	funcdecl_parser.RegisterParser(p)
	member_parser.RegisterParser(p)
	typedecl_parser.RegisterParser(p)
	matchstmt_parser.RegisterParser(p)
	let_parser.RegisterParser(p)
	return p
}

func TestEnumReceiverMethod(t *testing.T) {
	p := newMatchParser(`
		enum Message {
			Quit,
			Write(string),
		}

		fn (m: Message) describe() string {
			return match m {
				Message.Quit => "quit",
				Message.Write(text) => text,
			}
		}

		let msg = Message.Write("hello")
		let result = msg.describe()
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	for _, expr := range exprs {
		if _, err := expr.Eval(p.Ctx()); err != nil {
			t.Fatal(err)
		}
	}
	result, ok := p.Ctx().GetValue("result")
	if !ok || result.UnsafeCastString() != "hello" {
		t.Fatalf("result = %#v, want hello", result.Any())
	}

	parts := make([]string, 0, len(exprs))
	for _, expr := range exprs {
		printed, err := expr.PrintGO(p.Ctx())
		if err != nil {
			t.Fatal(err)
		}
		parts = append(parts, printed)
	}
	if !strings.Contains(
		parts[1],
		"func Message_describe(m Message) string",
	) {
		t.Fatalf("unexpected enum method:\n%s", parts[1])
	}
	if !strings.Contains(
		parts[3],
		"result := Message_describe(msg)",
	) {
		t.Fatalf("unexpected enum method call:\n%s", parts[3])
	}

	source := "package generated\n\n" +
		parts[0] + "\n\n" +
		parts[1] + "\n\nfunc run() {\n" +
		parts[2] + "\n" +
		parts[3] + "\n_ = result\n}"
	typeCheckGeneratedGo(t, source)
}

func TestStructEnumVariantAndNestedTuplePattern(t *testing.T) {
	p := newMatchParser(`
		enum Message {
			Text(string),
		}

		enum Event {
			Move { point: (int, int) },
			Nested(Message),
		}

		let event = Event.Move { point: (3, 4) }
		let result = match event {
			Event.Move { point: (x, y) } if x > 0 => y,
			Event.Move { point: (_, _) } => 0,
			Event.Nested(Message.Text(_)) => 0,
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	for _, expr := range exprs {
		if _, err := expr.Eval(p.Ctx()); err != nil {
			t.Fatal(err)
		}
	}
	result, ok := p.Ctx().GetValue("result")
	if !ok || result.Any() != int64(4) {
		t.Fatalf("result = %#v, want 4", result.Any())
	}

	parts := make([]string, 0, len(exprs))
	for _, expr := range exprs {
		printed, err := expr.PrintGO(p.Ctx())
		if err != nil {
			t.Fatal(err)
		}
		parts = append(parts, printed)
	}
	for _, expected := range []string{
		"Event(EventMove{point:",
		"x := __matchVariant0.point.V0",
		"y := __matchVariant0.point.V1",
		".(MessageText)",
	} {
		joined := strings.Join(parts, "\n")
		if !strings.Contains(joined, expected) {
			t.Fatalf(
				"generated nested pattern misses %q:\n%s",
				expected,
				joined,
			)
		}
	}

	source := "package generated\n\n" +
		parts[0] + "\n\n" +
		parts[1] + "\n\nfunc run() {\n" +
		parts[2] + "\n" +
		parts[3] + "\n_ = result\n}"
	typeCheckGeneratedGo(t, source)
}

func TestEnumMatchPayloadGuardAndExhaustiveness(t *testing.T) {
	p := newMatchParser(`
		enum Message {
			Quit,
			Move(int, int),
			Write(string),
		}

		let msg = Message.Move(12, 4)
		let result = match msg {
			Message.Quit => "quit",
			Message.Move(x, _) if x > 10 => "large move",
			Message.Move(_, _) => "move",
			Message.Write(text) => text,
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	for _, expr := range exprs {
		if _, err := expr.Eval(p.Ctx()); err != nil {
			t.Fatal(err)
		}
	}

	result, ok := p.Ctx().GetValue("result")
	if !ok || result.UnsafeCastString() != "large move" {
		t.Fatalf(
			"result = %#v, want large move",
			result.Any(),
		)
	}

	constructed, err := exprs[1].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(
		constructed,
		"Message(MessageMove{V0: 12, V1: 4})",
	) {
		t.Fatalf("unexpected enum construction:\n%s", constructed)
	}

	generated, err := exprs[2].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"result := func() string",
		"__matchValue.(MessageMove)",
		"x := __matchVariant0.V0",
		"if (x > 10)",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf(
				"generated enum match misses %q:\n%s",
				expected,
				generated,
			)
		}
	}

	enumDecl, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	source := "package generated\n\n" +
		enumDecl + "\n\nfunc run() {\n" +
		constructed + "\n" +
		generated + "\n_ = result\n}"
	typeCheckGeneratedGo(t, source)
}

func typeCheckGeneratedGo(t *testing.T, source string) {
	t.Helper()
	files := gotoken.NewFileSet()
	file, err := goparser.ParseFile(
		files,
		"generated.go",
		source,
		goparser.AllErrors,
	)
	if err != nil {
		t.Fatalf("generated Go does not parse: %v\n%s", err, source)
	}
	if _, err := (&types.Config{}).Check(
		"generated",
		files,
		[]*astgo.File{file},
		nil,
	); err != nil {
		t.Fatalf(
			"generated Go does not type-check: %v\n%s",
			err,
			source,
		)
	}
}

func TestEnumMatchReportsMissingVariant(t *testing.T) {
	p := newMatchParser(`
		enum State { Idle, Running, Stopped }
		let state = State.Idle
		let result = match state {
			State.Idle => 0,
			State.Running => 1,
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[1].PrintGO(p.Ctx()); err != nil {
		t.Fatal(err)
	}
	_, err = exprs[2].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(err.Error(), "State.Stopped") {
		t.Fatalf("expected missing variant error, got %v", err)
	}
}

func TestEnumPayloadTypeIsChecked(t *testing.T) {
	p := newMatchParser(`
		enum Message { Write(string) }
		let msg = Message.Write(42)
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	_, err = exprs[1].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(err.Error(), "expects string, got int") {
		t.Fatalf("expected enum payload type error, got %v", err)
	}
}

func TestEnumPatternRejectsDifferentSubjectType(t *testing.T) {
	p := newMatchParser(`
		enum State { Idle }
		let number = 1
		let result = match number {
			State.Idle => 0,
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[1].PrintGO(p.Ctx()); err != nil {
		t.Fatal(err)
	}
	_, err = exprs[2].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(
			err.Error(),
			"State.Idle is incompatible with int",
		) {
		t.Fatalf("expected enum subject type error, got %v", err)
	}
}

func TestMatchGuardMustBeBool(t *testing.T) {
	p := newMatchParser(`
		let result = match 12 {
			n if n => 1,
			_ => 0,
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	_, err = exprs[0].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(err.Error(), "guard must be bool, got int") {
		t.Fatalf("expected bool guard error, got %v", err)
	}
}

func TestLetMatchLiteralAndWildcard(t *testing.T) {
	p := newMatchParser(`
		let result = match 2 {
			1 => "one",
			2 => "two",
			_ => "other",
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}
	result, ok := p.Ctx().GetValue("result")
	if !ok || result.UnsafeCastString() != "two" {
		t.Fatalf("result = %#v, want two", result.Any())
	}

	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"result := func() string",
		"__matchValue := 2",
		"if __matchValue == 2",
		`return "two"`,
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf(
				"generated match misses %q:\n%s",
				expected,
				generated,
			)
		}
	}
}

func TestLetMatchBindingGuard(t *testing.T) {
	p := newMatchParser(`
		let result = match 12 {
			n if n > 10 => n,
			_ => 0,
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exprs[0].Eval(p.Ctx()); err != nil {
		t.Fatal(err)
	}
	result, ok := p.Ctx().GetValue("result")
	if !ok || result.Any() != int64(12) {
		t.Fatalf("result = %#v, want 12", result.Any())
	}

	generated, err := exprs[0].PrintGO(p.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(
		generated,
		"n := __matchValue",
	) || !strings.Contains(
		generated,
		"if (n > 10)",
	) || !strings.Contains(generated, "return n") {
		t.Fatalf("unexpected guarded match:\n%s", generated)
	}
}

func TestLetMatchRejectsIncompatibleArmTypes(t *testing.T) {
	p := newMatchParser(`
		let result = match 1 {
			1 => "one",
			_ => 0,
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	_, err = exprs[0].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(
			err.Error(),
			"match arms return incompatible types string and int",
		) {
		t.Fatalf("expected incompatible arm error, got %v", err)
	}
}

func TestMatchRequiresFallback(t *testing.T) {
	p := newMatchParser(`
		let result = match 1 {
			1 => "one",
		}
	`)

	exprs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	_, err = exprs[0].PrintGO(p.Ctx())
	if err == nil ||
		!strings.Contains(err.Error(), "non-exhaustive match") {
		t.Fatalf("expected exhaustive match error, got %v", err)
	}
}
