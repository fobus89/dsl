// Package token defines token types and utilities for the foo_lang lexer.
// It includes all language keywords, operators, literals, and special symbols.
package token

import "slices"

type TokenType int

const (
	token_start TokenType = iota

	// Special
	EOF     // end of file
	ILLEGAL // unknown token

	// Literals
	literal_start
	STRING_LITERAL // "string"
	STRING_FORMAT  // "{expr}"
	NUMBER_LITERAL // int64 | float64
	FLOAT_LITERAL  // 0.0-9.0
	INT_LITERAL    // 0-9
	literal_end

	// Identifiers
	IDENT // identifier

	// Symbols (operators and punctuation)
	symbol_start

	// Arithmetic operators
	arithmetic_start
	PLUS        // +
	MINUS       // -
	STAR        // *
	STAR_STAR   // **
	SLASH       // /
	PERCENT     // %
	PLUS_PLUS   // ++
	MINUS_MINUS // --
	arithmetic_end

	// Comparison operators
	comparison_start
	EQ_EQ   // ==
	BANG_EQ // !=
	GT      // >
	LT      // <
	GT_EQ   // >=
	LT_EQ   // <=
	comparison_end

	// Assignment operators
	assignment_start
	EQ         // =
	PLUS_EQ    // +=
	MINUS_EQ   // -=
	STAR_EQ    // *=
	SLASH_EQ   // /=
	PERCENT_EQ // %=
	assignment_end

	AMP  // & |
	PIPE // | |

	// Logical operators
	logical_start
	BANG      // !
	AMP_AMP   // && | and
	PIPE_PIPE // || | or
	logical_end

	// Punctuation
	punctuation_start
	QUESTION    // ?
	COLON       // :
	COLON_COLON // ::
	SEMICOLON   // ;
	COMMA       // ,
	DOT         // .
	AT          // @
	HASH        // #
	DOLLAR      // $
	LPARENT     // (
	RPARENT     // )
	LBRACE      // {
	RBRACE      // }
	LBRACKET    // [
	RBRACKET    // ]
	punctuation_end

	symbol_end

	// Keywords
	keyword_start
	AND       // &&
	OR        // ||
	FALSE     // false
	TRUE      // true
	AS        // as
	ASYNC     // async
	AWAIT     // await
	BETWEEN   // between
	BREAK     // break
	CATCH     // catch
	CONST     // const
	DEFER     // defer
	DO        // do
	ELSE      // else
	ENUM      // enum
	EXPORT    // export
	FINALLY   // finally
	FN        // fn
	FOR       // for
	FROM      // from
	IF        // if
	IMPL      // impl
	IMPORT    // import
	IN        // in
	INTERFACE // interface
	IS        // is
	LET       // let
	LOOP      // loop
	MATCH     // match
	MUT       // mut
	NIL       // nil
	PUB       // pub
	REF       // ref
	RETURN    // return
	SELF      // self
	SPAWN     // spawn

	// STRUCT    // struct
	SUPER  // super
	THROW  // throw
	TRY    // try
	TYPE   // type
	TYPEOF // typeof
	USE    // use
	WHERE  // where
	WHILE  // while
	YIELD  // yield

	// SQL keywords
	ALL      // all
	ALTER    // alter
	ASC      // asc
	BY       // by
	CREATE   // create
	CROSS    // cross
	DESC     // desc
	DISTINCT // distinct
	DROP     // drop
	EXISTS   // exists
	FULL     // full
	GROUP    // group
	HAVING   // having
	INNER    // inner
	INTO     // into
	JOIN     // join
	LEFT     // left
	LIKE     // like
	LIMIT    // limit
	OFFSET   // offset
	ON       // on
	ORDER    // order
	OUTER    // outer
	RIGHT    // right
	SELECT   // select
	SET      // set
	TABLE    // table
	UNION    // union
	VALUES   // values
	keyword_end

	// Built-in functions
	builtin_start
	PRINT   // print
	PRINTLN // println
	builtin_end

	// Compile-time / macros
	compiletime_start
	MACROS         // macro fn
	EXPR           // expr
	CONSTEXPR      // $constexpr
	COMPTIME_IDENT // $"{expr}string"
	compiletime_end

	// Types
	type_start
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Float32
	Float64
	String
	Char
	Byte
	Rune
	Array
	Slice
	Map
	Tuple
	Struct
	Enum
	Interface
	Any
	Void
	Never
	type_end

	token_end
)

var markers = []TokenType{
	token_start, token_end,
	literal_start, literal_end,
	symbol_start, symbol_end,
	arithmetic_start, arithmetic_end,
	comparison_start, comparison_end,
	assignment_start, assignment_end,
	logical_start, logical_end,
	punctuation_start, punctuation_end,
	keyword_start, keyword_end,
	builtin_start, builtin_end,
	compiletime_start, compiletime_end,
	type_start, type_end,
}

// MarkersContains reports whether the given TokenType is a range marker.
// Exported for use by the LSP to skip markers when iterating token ranges.
func MarkersContains(t TokenType) bool {
	return slices.Contains(markers, t)
}

func (t TokenType) IsValid() bool {
	return token_start < t && t < token_end && !MarkersContains(t)
}

func (t TokenType) IsLiteral() bool {
	return literal_start < t && t < literal_end
}

func (t TokenType) IsSymbol() bool {
	return symbol_start < t && t < symbol_end && !slices.Contains(markers, t)
}

func (t TokenType) IsArithmetic() bool {
	return arithmetic_start < t && t < arithmetic_end
}

func (t TokenType) IsComparison() bool {
	return comparison_start < t && t < comparison_end
}

func (t TokenType) IsAssignment() bool {
	return assignment_start < t && t < assignment_end
}

func (t TokenType) IsLogical() bool {
	return logical_start < t && t < logical_end
}

func (t TokenType) IsPunctuation() bool {
	return punctuation_start < t && t < punctuation_end
}

func (t TokenType) IsKeyword() bool {
	return keyword_start < t && t < keyword_end && !slices.Contains(markers, t)
}

func (t TokenType) IsBuiltin() bool {
	return builtin_start < t && t < builtin_end
}

func (t TokenType) IsCompiletime() bool {
	return compiletime_start < t && t < compiletime_end
}

func (t TokenType) IsType() bool {
	return type_start < t && t < type_end
}

// EachKeyword yields all valid keyword TokenTypes.
// Exported for the LSP to auto-generate completion items.
func EachKeyword(yield func(TokenType) bool) {
	for i := keyword_start + 1; i < keyword_end; i++ {
		if MarkersContains(i) {
			continue
		}
		if !yield(i) {
			return
		}
	}
}

// EachType yields all valid type TokenTypes.
func EachType(yield func(TokenType) bool) {
	for i := type_start + 1; i < type_end; i++ {
		if MarkersContains(i) {
			continue
		}
		if !yield(i) {
			return
		}
	}
}

// EachBuiltin yields all valid builtin TokenTypes.
func EachBuiltin(yield func(TokenType) bool) {
	for i := builtin_start + 1; i < builtin_end; i++ {
		if MarkersContains(i) {
			continue
		}
		if !yield(i) {
			return
		}
	}
}

// EachCompiletime yields all valid compile-time TokenTypes.
func EachCompiletime(yield func(TokenType) bool) {
	for i := compiletime_start + 1; i < compiletime_end; i++ {
		if MarkersContains(i) {
			continue
		}
		if !yield(i) {
			return
		}
	}
}
