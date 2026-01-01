package lexer

type TokenType uint8

const (
	INT TokenType = iota
	PLUS
	SEMICOLON
	ASSIGN
	IDENTIFIER
	DEF
	TYPE
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
	case TYPE:
		return "TYPE"
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
)

func (vt VarType) String() string {
	switch vt {
	case None:
		return "None"
	case Int:
		return "Int"
	default:
		return "Unknown"
	}
}

type Token struct {
	Type    TokenType
	VarType VarType
	Literal string
}
