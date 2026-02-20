package lexer

import (
	"grianlang3/util"
)

type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte

	currLine uint32
	currCh   uint32
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()

	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}

	l.pos = l.readPos
	l.readPos += 1
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPos]
	}
}

var singleCharToken = map[byte]TokenType{
	'+': PLUS,
	'*': ASTERISK,
	'/': SLASH,
	';': SEMICOLON,
	'(': LPAREN,
	')': RPAREN,
	'{': LBRACE,
	'}': RBRACE,
	'[': LBRACKET,
	']': RBRACKET,
	',': COMMA,
	'.': DOT,
	':': COLON,
}

func (l *Lexer) NextToken() Token {
	var tok Token

	for {
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			if l.ch == '\n' {
				l.currLine++
				l.currCh = 0
			}
			l.readChar()
		}

		if l.ch == '/' && l.peekChar() == '/' {
			for l.ch != '\n' {
				l.readChar()
			}
			l.currLine++
			l.currCh = 0
			continue
		}

		break
	}

	sct, ok := singleCharToken[l.ch]
	if ok {
		tok = newToken(sct, l.ch, l.currLine, l.currCh)
		// ret early here is a bit of future proofing/opti
		l.readChar()
		return tok
	}

	switch l.ch {
	case '-':
		tok = l.doubleCharToken('>', MINUS, ARROW, l.currLine, l.currCh)
	case '&':
		tok = l.doubleCharToken('&', AMPERSAND, LAND, l.currLine, l.currCh)
	case '|':
		tok = l.doubleCharToken('|', UNKNOWN, LOR, l.currLine, l.currCh)
	case '=':
		tok = l.doubleCharToken('=', ASSIGN, EQ, l.currLine, l.currCh)
	case '!':
		tok = l.doubleCharToken('=', NOT, NOTEQ, l.currLine, l.currCh)
	case '<':
		tok = l.doubleCharToken('=', LT, LTEQ, l.currLine, l.currCh)
	case '>':
		tok = l.doubleCharToken('=', GT, GTEQ, l.currLine, l.currCh)
	case 0:
		tok.Literal = ""
		tok.Type = EOF
		tok.Position = PositionInfo{
			Line: l.currLine,
			Col:  l.currCh,
		}
	default:
		if l.ch == '0' && l.peekChar() == 'x' {
			tok.Position = PositionInfo{
				Line: l.currLine,
				Col:  l.currCh,
			}
			tok.Literal = l.readHexaInt()
			tok.Type = INT
			return tok
		}

		if util.IsDigit(l.ch) {
			tok.Position = PositionInfo{
				Line: l.currLine,
				Col:  l.currCh,
			}
			tok.Literal, tok.Type = l.readNumber()
			return tok
		}

		if util.IsAlpha(l.ch) {
			tok.Position = PositionInfo{
				Line: l.currLine,
				Col:  l.currCh,
			}
			l.readChar()
			tok.Literal = l.readIdentifier()
			tok.Type, tok.VarType.Base = identLookup(tok.Literal)
			return tok
		}
		if l.ch == '"' {
			tok.Position = PositionInfo{
				Line: l.currLine,
				Col:  l.currCh,
			}
			l.readChar()
			tok.Type = STRING
			//tok.VarType = VarType{Base: Int8, Pointer: 1};
			tok.Literal = l.readString()
		}

		if l.ch == '\'' {
			tok.Position = PositionInfo{
				Line: l.currLine,
				Col:  l.currCh,
			}
			l.readChar()
			tok.Type = CHAR
			if l.ch == '\'' {
				tok.Literal = ""
			} else {
				tok.Literal = string(l.ch)
				l.readChar() // skip next '
			}
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readString() string {
	startPos := l.pos

	for l.ch != '"' {
		l.readChar()
	}

	return l.input[startPos:l.pos]
}

func (l *Lexer) readHexaInt() string {
	startPos := l.pos
	l.readChar()
	l.readChar()

	for util.IsHexaNumeral(l.ch) {
		l.readChar()
	}

	return l.input[startPos:l.pos]
}

func (l *Lexer) readNumber() (string, TokenType) {
	startPos := l.pos
	tt := INT

	for util.IsDigit(l.ch) || l.ch == '.' {
		if l.ch == '.' {
			tt = FLOAT
		}
		l.readChar()
	}

	return l.input[startPos:l.pos], tt
}

func (l *Lexer) readIdentifier() string {
	startPos := l.pos - 1

	for util.IsAlphaNumeric(l.ch) {
		l.readChar()
	}

	return l.input[startPos:l.pos]
}

func newToken(tt TokenType, ch byte, currLine, currCh uint32) Token {
	return Token{Type: tt, Literal: string(ch), Position: PositionInfo{Line: currLine, Col: currCh}}
}

func (l *Lexer) doubleCharToken(char2 byte, tt TokenType, tt2 TokenType, currLine, currCh uint32) Token {
	var tok Token

	if l.peekChar() == char2 {
		l.readChar()
		l.readChar()
		tok.Type = tt2
		tok.Literal = l.input[l.pos-2 : l.pos]
		tok.Position = PositionInfo{
			Line: currLine,
			Col:  currCh,
		}
		return tok
	}

	if tt != UNKNOWN {
		tok = newToken(tt, l.ch, currLine, currCh)
	}

	return tok
}

func identLookup(lit string) (TokenType, BaseVarType) {
	switch lit {
	case "int":
		return TYPE, Int
	case "int32":
		return TYPE, Int32
	case "int16":
		return TYPE, Int16
	case "int8":
		return TYPE, Int8
	case "uint":
		return TYPE, Uint
	case "uint8":
		return TYPE, Uint8
	case "uint16":
		return TYPE, Uint16
	case "uint32":
		return TYPE, Uint32
	case "none":
		return TYPE, Void
	case "def":
		return DEF, None
	case "fnc":
		return FNC, None
	case "return":
		return RETURN, None
	case "bool":
		return TYPE, Bool
	case "true":
		return TRUE, None
	case "false":
		return FALSE, None
	case "float":
		return TYPE, Float
	case "as":
		return AS, None
	case "sizeof":
		return SIZEOF, None
	case "import":
		return IMPORT, None
	case "char":
		return TYPE, Char
	case "if":
		return IF, None
	case "else":
		return ELSE, None
	case "while":
		return WHILE, None
	case "struct":
		return STRUCT, None
	}

	return IDENTIFIER, None
}
