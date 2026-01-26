package main

import (
	"embed"
	"fmt"
	"grianlang3/emitter"
	"grianlang3/lexer"
	"grianlang3/parser"
	"log"
	"os"
	"os/exec"
)

//go:embed builtins/*.ll
var builtinFs embed.FS

func main() {
	if err := os.Mkdir("./lltemp", os.ModePerm); err != nil {
		log.Fatal(err)
	}
	files := os.Args[1:]
	keepll := false
	if files[0] == "--keepll" {
		keepll = true
		files = os.Args[2:]
	}
	var llFiles []string
	builtinModules := map[string]struct{}{}
	for _, file := range files {
		input, err := os.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}
		l := lexer.New(string(input))
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors) != 0 {
			for _, err := range p.Errors {
				log.Printf("%s: parser error: %s\n", file, err)
			}
			log.Printf("%s: %s", file, program.String())
			log.Fatalf("%s: exiting after parser errors\n", file)
		}
		e := emitter.New()
		e.Emit(program, nil)
		llvmIr := e.Module()
		fileName := fmt.Sprintf("./lltemp/%s.ll", file)
		llFile, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("%s: %v\n", file, err)
		}
		_, err = fmt.Fprintf(llFile, "%s", llvmIr)
		if err != nil {
			log.Fatalf("%s: %v\n", file, err)
		}
		err = llFile.Close()
		if err != nil {
			log.Fatalf("%s: %v\n", file, err)
		}

		llFiles = append(llFiles, fileName)

		for _, builtinModule := range e.BuiltinModules() {
			builtinModules[builtinModule] = struct{}{}
		}
	}

	for mod, _ := range builtinModules {
		modText, err := builtinFs.ReadFile(fmt.Sprintf("builtins/%s", mod))
		if err != nil {
			log.Fatalf("failed to read %s from builtin fs: %v\n", mod, err)
		}
		fileName := fmt.Sprintf("./lltemp/%s", mod)
		llFile, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("failed to create %s: %v\n", llFile, err)
		}
		_, err = fmt.Fprintf(llFile, "%s", modText)
		if err != nil {
			log.Fatalf("failed to write %s: %v\n", llFile, err)
		}
		err = llFile.Close()
		if err != nil {
			log.Fatalf("failed to close %s: %v\n", llFile, err)
		}
		llFiles = append(llFiles, fileName)
	}

	llFiles = append(llFiles, "-o", "out")
	out, err := exec.Command("clang", llFiles...).CombinedOutput()
	if err != nil {
		fmt.Printf("out in clang exec: %s\n", out)
		fmt.Printf("err in clang exec: %v\n", err)
		log.Fatalf("existing after err in clang exec\n")
	}
	if !keepll {
		err := os.RemoveAll("./lltemp")
		if err != nil {
			log.Fatal(err)
		}
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
