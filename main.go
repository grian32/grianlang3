package main

import (
	"fmt"
	"grianlang3/lexer"
	"grianlang3/parser"
)

func main() {
	input := "1+4+3"

	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()

	for _, s := range program.Statements {
		fmt.Printf("%s\n", s.String())
	}
}
