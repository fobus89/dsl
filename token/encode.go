package token

import (
	"encoding/json"
	"fmt"
)

var (
	_ json.Marshaler   = (*TokenType)(nil)
	_ json.Unmarshaler = (*TokenType)(nil)
)

func (t TokenType) MarshalJSON() ([]byte, error) {
	symbol, ok := tokenToStringMap.Get(t)
	{
		if !ok {
			return nil, fmt.Errorf("%s unknow token", t)
		}
	}

	return json.Marshal(symbol)
}

func (t TokenType) UnmarshalJSON([]byte) error {
	panic("unimplemented")
}

func (t TokenType) String() string {
	switch t {
	// Special
	case EOF:
		return "eof"
	case ILLEGAL:
		return "illegal"

	// Literals
	case STRING_LITERAL:
		return "string_literal"
	case STRING_FORMAT:
		return "string_format"
	case NUMBER_LITERAL:
		return "number_literal"
	case FLOAT_LITERAL:
		return "float_literal"
	case INT_LITERAL:
		return "int_literal"

	// Identifiers
	case IDENT:
		return "ident"

	// Arithmetic operators
	case PLUS:
		return "+"
	case MINUS:
		return "-"
	case STAR:
		return "*"
	case STAR_STAR:
		return "**"
	case SLASH:
		return "/"
	case PERCENT:
		return "%"
	case PLUS_PLUS:
		return "++"
	case MINUS_MINUS:
		return "--"

	// Bitwise operators
	case AMP:
		return "&"
	case PIPE:
		return "|"
	case CARET:
		return "^"
	case TILDE:
		return "~"
	case LT_LT:
		return "<<"
	case GT_GT:
		return ">>"

	// Comparison operators
	case EQ_EQ:
		return "=="
	case BANG_EQ:
		return "!="
	case GT:
		return ">"
	case LT:
		return "<"
	case GT_EQ:
		return ">="
	case LT_EQ:
		return "<="

	// Assignment operators
	case EQ:
		return "="
	case PLUS_EQ:
		return "+="
	case MINUS_EQ:
		return "-="
	case STAR_EQ:
		return "*="
	case SLASH_EQ:
		return "/="
	case PERCENT_EQ:
		return "%="
	case AMP_EQ:
		return "&="
	case PIPE_EQ:
		return "|="
	case LT_LT_EQ:
		return "<<="
	case GT_GT_EQ:
		return ">>="

	// Logical operators
	case BANG:
		return "!"
	case AMP_AMP:
		return "&&"
	case PIPE_PIPE:
		return "||"

	// Other operators
	case EQ_GT:
		return "=>"
	case MINUS_GT:
		return "->"
	case LT_MINUS:
		return "<-"
	case DOT_DOT:
		return ".."
	case DOT_DOT_EQ:
		return "..="
	case DOT_DOT_DOT:
		return "..."

	// Punctuation
	case QUESTION:
		return "?"
	case COLON:
		return ":"
	case COLON_COLON:
		return "::"
	case SEMICOLON:
		return ";"
	case COMMA:
		return ","
	case DOT:
		return "."
	case AT:
		return "@"
	case HASH:
		return "#"
	case DOLLAR:
		return "$"
	case LPARENT:
		return "("
	case RPARENT:
		return ")"
	case LBRACE:
		return "{"
	case RBRACE:
		return "}"
	case LBRACKET:
		return "["
	case RBRACKET:
		return "]"

	// Keywords
	case AS:
		return "as"
	case ASYNC:
		return "async"
	case AWAIT:
		return "await"
	case BETWEEN:
		return "between"
	case BREAK:
		return "break"
	case CATCH:
		return "catch"
	case CONST:
		return "const"
	case DEFER:
		return "defer"
	case DO:
		return "do"
	case ELSE:
		return "else"
	case ENUM:
		return "enum"
	case EXPORT:
		return "export"
	case FALSE:
		return "false"
	case FINALLY:
		return "finally"
	case FN:
		return "fn"
	case FOR:
		return "for"
	case FROM:
		return "from"
	case IF:
		return "if"
	case IMPL:
		return "impl"
	case IMPORT:
		return "import"
	case IN:
		return "in"
	case INTERFACE:
		return "interface"
	case IS:
		return "is"
	case LET:
		return "let"
	case LOOP:
		return "loop"
	case MATCH:
		return "match"
	case MUT:
		return "mut"
	case NIL:
		return "nil"
	case PUB:
		return "pub"
	case REF:
		return "ref"
	case RETURN:
		return "return"
	case SELF:
		return "self"
	case SPAWN:
		return "spawn"
	// case STRUCT:
	// 	return "struct"
	case SUPER:
		return "super"
	case THROW:
		return "throw"
	case TRUE:
		return "true"
	case TRY:
		return "try"
	case TYPE:
		return "type"
	case TYPEOF:
		return "typeof"
	case USE:
		return "use"
	case WHERE:
		return "where"
	case WHILE:
		return "while"
	case YIELD:
		return "yield"

	// SQL keywords
	case ALL:
		return "all"
	case ALTER:
		return "alter"
	case ASC:
		return "asc"
	case BY:
		return "by"
	case CREATE:
		return "create"
	case CROSS:
		return "cross"
	case DESC:
		return "desc"
	case DISTINCT:
		return "distinct"
	case DROP:
		return "drop"
	case EXISTS:
		return "exists"
	case FULL:
		return "full"
	case GROUP:
		return "group"
	case HAVING:
		return "having"
	case INNER:
		return "inner"
	case INTO:
		return "into"
	case JOIN:
		return "join"
	case LEFT:
		return "left"
	case LIKE:
		return "like"
	case LIMIT:
		return "limit"
	case OFFSET:
		return "offset"
	case ON:
		return "on"
	case ORDER:
		return "order"
	case OUTER:
		return "outer"
	case RIGHT:
		return "right"
	case SELECT:
		return "select"
	case SET:
		return "set"
	case TABLE:
		return "table"
	case UNION:
		return "union"
	case VALUES:
		return "values"

	// Built-in functions
	case PRINT:
		return "print"
	case PRINTLN:
		return "println"

	// Compile-time / macros
	case MACROS:
		return "macros"
	case EXPR:
		return "expr"
	case CONSTEXPR:
		return "constexpr"
	case COMPTIME_IDENT:
		return "comptime_ident"

	// Types
	case Bool:
		return "bool"
	case Int:
		return "int"
	case Int8:
		return "int8"
	case Int16:
		return "int16"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Uint:
		return "uint"
	case Uint8:
		return "uint8"
	case Uint16:
		return "uint16"
	case Uint32:
		return "uint32"
	case Uint64:
		return "uint64"
	case Float32:
		return "float32"
	case Float64:
		return "float64"
	case String:
		return "string"
	case Char:
		return "char"
	case Byte:
		return "byte"
	case Rune:
		return "rune"
	case Array:
		return "array"
	case Slice:
		return "slice"
	case Map:
		return "map"
	case Tuple:
		return "tuple"
	case Struct:
		return "struct"
	case Enum:
		return "enum"
	case Interface:
		return "interface"
	case Any:
		return "any"
	case Void:
		return "void"
	case Never:
		return "never"
	}

	return "UNKNOW_TOKEN"
}
