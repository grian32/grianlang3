package lexer

import "strings"

type TokenType uint8

const (
	UNKNOWN TokenType = iota
	INT
	FLOAT
	PLUS
	MINUS
	ASTERISK
	SLASH
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
	AMPERSAND
	TRUE
	FALSE
	NOT
	LAND
	LOR
	EQ
	NOTEQ
	LT
	LTEQ
	GT
	GTEQ
	AS
	EOF
)

func (tt TokenType) String() string {
	switch tt {
	case INT:
		return "INT"
	case PLUS:
		return "PLUS"
	case SLASH:
		return "SLASH"
	case ASTERISK:
		return "ASTERISK"
	case MINUS:
		return "MINUS"
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
	case AMPERSAND:
		return "AMPERSAND"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	case NOT:
		return "NOT"
	case LAND:
		return "LAND"
	case LOR:
		return "LOR"
	case EQ:
		return "EQ"
	case NOTEQ:
		return "NOTEQ"
	case LT:
		return "LT"
	case LTEQ:
		return "LTEQ"
	case GT:
		return "GT"
	case GTEQ:
		return "GTEQ"
	case EOF:
		return "EOF"
	case AS:
		return "AS"
	case FLOAT:
		return "FLOAT"
	default:
		return "UNKNOWN"
	}
}

type VarType struct {
	Base    BaseVarType
	Pointer uint8
}

type BaseVarType uint8

const (
	None BaseVarType = iota
	Int
	Int32
	Bool
	Void
	Float
)

func (bvt BaseVarType) String() string {
	switch bvt {
	case None:
		return "None"
	case Void:
		return "Void"
	case Int:
		return "Int"
	case Int32:
		return "Int32"
	case Bool:
		return "Bool"
	default:
		return "Unknown"
	}
}

func (vt VarType) String() string {
	var bvt strings.Builder
	bvt.WriteString(vt.Base.String())

	for _ = range vt.Pointer {
		bvt.WriteString("*")
	}

	return bvt.String()
}

type Token struct {
	Type    TokenType
	VarType VarType
	Literal string
}
