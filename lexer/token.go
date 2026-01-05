package lexer

type TokenType uint8

const (
	INT TokenType = iota
	PLUS
	SEMICOLON
	ASSIGN
	IDENTIFIER
	DEF
	FNC
	LPAREN
	RPAREN
	COMMA
	LBRACE
	RBRACE
	RETURN
	TYPE
	ARROW
	EOF
)

func (tt TokenType) String() string {
	switch tt {
	case INT:
		return "INT"
	case PLUS:
		return "PLUS"
	case SEMICOLON:
		return "SEMICOLON"
	case ASSIGN:
		return "ASSIGN"
	case IDENTIFIER:
		return "IDENTIFIER"
	case DEF:
		return "DEF"
	case FNC:
		return "FNC"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case COMMA:
		return "COMMA"
	case LBRACE:
		return "LBRACE"
	case RBRACE:
		return "RBRACE"
	case RETURN:
		return "RETURN"
	case TYPE:
		return "TYPE"
	case ARROW:
		return "ARROW"
	case EOF:
		return "EOF"
	default:
		return "UNKNOWN"
	}
}

type VarType uint8

const (
	None VarType = iota
	Int
	Int32
)

func (vt VarType) String() string {
	switch vt {
	case None:
		return "None"
	case Int:
		return "Int"
	case Int32:
		return "Int32"
	default:
		return "Unknown"
	}
}

type Token struct {
	Type    TokenType
	VarType VarType
	Literal string
}
