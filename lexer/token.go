package lexer

type TokenType uint8

const (
	INT TokenType = iota
	PLUS
	EOF
)

type Token struct {
	Type TokenType
	Literal string
}
