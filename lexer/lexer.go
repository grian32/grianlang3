package lexer

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
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if l.ch >= '0' && l.ch <= '9' {
			tok.Literal = l.readInt()
			tok.Type = INT
			return tok
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readInt() string {
	startPos := l.pos

	for l.ch >= '0' && l.ch <= '9' {
		l.readChar()
	}

	return l.input[startPos:l.pos]
}

func newToken(tt TokenType, ch byte) Token {
	return Token{Type: tt, Literal: string(ch)}
}
