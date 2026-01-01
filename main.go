package main

import (
	"fmt"
	"grianlang3/lexer"
	"os"
)

func main() {
	input, err := os.ReadFile("./test.gl3")
	if err != nil {
		fmt.Printf("failed to read input file: %v\n", err)
		return
	}
	fmt.Printf("from:\n%s\n", input)

	l := lexer.New(string(input))
	printTokens(l)

	//p := parser.New(l)
	//
	//program := p.ParseProgram()
	//
	//e := emitter.New()
	//
	//e.Emit(program)
	//
	//file, err := os.Open("./test.ll")
	//if err != nil {
	//	fmt.Printf("failed to wirte ll file: %v\n", err)
	//	return
	//}
	//llvmIr := e.Module()
	//fmt.Fprintf(file, "%s", llvmIr)
	//fmt.Printf("llvm ir:\n%s\n", llvmIr)
	//
	//_ = exec.Command("clang", "-c dbg.c -o dbg.o")
	//_ = exec.Command("clang", "test.ll dbg.o -o out")
	//execCmd := exec.Command("./out")
	//output, err := execCmd.Output()
	//if err != nil {
	//	fmt.Printf("err in binary exec: %v", err)
	//} else {
	//	fmt.Println(string(output))
	//}
}

// temp debug function
func printTokens(l *lexer.Lexer) {
	for {
		tok := l.NextToken()
		if tok.Type == lexer.EOF {
			fmt.Printf("%v\n", tok)
			break
		}
		fmt.Printf("%v\n", tok)
	}
}
