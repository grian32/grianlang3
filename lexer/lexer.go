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

func (l *Lexer) NextToken() Token {
	var tok Token

	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}

	switch l.ch {
	case '+':
		tok = newToken(PLUS, l.ch)
	case ';':
		tok = newToken(SEMICOLON, l.ch)
	case '=':
		tok = newToken(EQUALS, l.ch)
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
	startPos := l.pos

	for util.IsAlpha(l.ch) {
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
	case "def":
		return DEF, None
	}

	return IDENTIFIER, None
}
