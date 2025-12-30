package main

import (
	"fmt"
	"grianlang3/lexer"
)

func main() {
	input := "1+4+3"

	tokens := []lexer.Token{}

	l := lexer.New(input)

	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == lexer.EOF {
			break
		}
	}

	fmt.Printf("%v\n", tokens)
}
