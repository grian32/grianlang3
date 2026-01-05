package lexer

import "grianlang3/util"

type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
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
	';': SEMICOLON,
	'=': ASSIGN,
	'(': LPAREN,
	')': RPAREN,
	'{': LBRACE,
	'}': RBRACE,
	',': COMMA,
}

func (l *Lexer) NextToken() Token {
	var tok Token

	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}

	sct, ok := singleCharToken[l.ch]
	if ok {
		tok = newToken(sct, l.ch)
		// ret early here is a bit of future proofing/opti
		l.readChar()
		return tok
	}

	switch l.ch {
	case '-':
		if l.peekChar() == '>' {
			l.readChar()
			l.readChar()
			tok.Type = ARROW
			tok.Literal = l.input[l.pos-2 : l.pos]
		}
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if util.IsDigit(l.ch) {
			tok.Literal = l.readInt()
			tok.Type = INT
			return tok
		}

		if util.IsAlpha(l.ch) {
			l.readChar()
			tok.Literal = l.readIdentifier()
			tok.Type, tok.VarType = identLookup(tok.Literal)
			return tok
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readInt() string {
	startPos := l.pos

	for util.IsDigit(l.ch) {
		l.readChar()
	}

	return l.input[startPos:l.pos]
}

func (l *Lexer) readIdentifier() string {
	startPos := l.pos - 1

	for util.IsAlphaNumeric(l.ch) {
		l.readChar()
	}

	return l.input[startPos:l.pos]
}

func newToken(tt TokenType, ch byte) Token {
	return Token{Type: tt, Literal: string(ch)}
}

func identLookup(lit string) (TokenType, VarType) {
	switch lit {
	case "int":
		return TYPE, Int
	case "int32":
		return TYPE, Int32
	case "def":
		return DEF, None
	case "fnc":
		return FNC, None
	case "return":
		return RETURN, None
	}

	return IDENTIFIER, None
}
