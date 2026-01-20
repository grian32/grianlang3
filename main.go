package main

import (
	"fmt"
	"grianlang3/emitter"
	"grianlang3/lexer"
	"grianlang3/parser"
	"os"
	"os/exec"
)

func main() {
	input, err := os.ReadFile("./test.gl3")
	if err != nil {
		fmt.Printf("failed to read input file: %v\n", err)
		return
	}
	fmt.Printf("from:\n%s\n", input)

	l := lexer.New(string(input))
	// printTokens(l)

	p := parser.New(l)

	program := p.ParseProgram()

	//fmt.Printf("%s\n", program.String())

	e := emitter.New()

	e.Emit(program, nil)

	file, err := os.Create("./test.ll")
	if err != nil {
		fmt.Printf("failed to write ll file: %v\n", err)
		return
	}
	defer file.Close()
	llvmIr := e.Module()
	fmt.Fprintf(file, "%s", llvmIr)
	//fmt.Printf("\n\nllvm ir:\n%s\n", llvmIr)

	// TODO: walk builtins/ automatically for compilation.. auto remove the .o files
	out, err := exec.Command("clang", "-c", "dbg.c", "-o", "dbg.o").CombinedOutput()
	if err != nil {
		fmt.Printf("out in clang exec: %s\n", out)
		fmt.Printf("err in clang exec: %v\n", err)
	}
	out, err = exec.Command("clang", "-c", "builtins/arrays.c", "-o", "array.o").CombinedOutput()
	if err != nil {
		fmt.Printf("out in clang exec: %s\n", out)
		fmt.Printf("err in clang exec: %v\n", err)
	}
	out, err = exec.Command("clang", "test.ll", "dbg.o", "array.o", "-o", "out").CombinedOutput()
	if err != nil {
		fmt.Printf("out in clang exec: %s\n", out)
		fmt.Printf("err in clang exec: %v\n", err)
	}
	execCmd := exec.Command("./out")
	output, err := execCmd.Output()
	if err != nil {
		fmt.Printf("err in binary exec: %v\n", err)
	} else {
		fmt.Println(string(output))
	}
	err = os.Remove("./out")
	if err != nil {
		fmt.Printf("failed to delete out file: %v\n", err)
		return
	}
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
