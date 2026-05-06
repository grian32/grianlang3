package cli

import (
	"embed"
	"fmt"
	"grianlang3/checker"
	"grianlang3/emitter"
	"grianlang3/lexer"
	"grianlang3/parser"
	"log"
	"os"
	"os/exec"
	"strings"
)

type BuildOpts struct {
	KeepLL      bool
	Dbg         bool
	NoExecBuild bool
}

func RunBuildCmd(builtinFs embed.FS, files []string, opts *BuildOpts) error {
	// TODO: really should do this in /tmp
	if err := os.RemoveAll("./lltemp"); err != nil {
		return err
	}
	if err := os.Mkdir("./lltemp", os.ModePerm); err != nil {
		return err
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
		err = safeRun(func() {
			if opts.Dbg {
				log.Printf("%s: %s\n", file, program.String())
			}
		})
		if err != nil {
			return err
		}
		if len(p.Errors) != 0 {
			for _, err := range p.Errors {
				log.Printf("parser error: %s:%s\n", file, err.String())
			}
			return fmt.Errorf("%s: exiting after parser errrors\n", file)
		}
		c := checker.New()
		c.Check(program)
		if len(c.Errors) != 0 {
			for _, err := range c.Errors {
				log.Printf("checker warning: %s:%s\n", file, err.String())
			}
		}

		e := emitter.New()
		err = safeRun(func() {
			e.Emit(program)
		})
		if len(e.Errors) != 0 {
			for _, err := range e.Errors {
				log.Printf("compiler error: %s:%s\n", file, err.String())
			}
			return fmt.Errorf("compiler errors\n")
		}
		if err != nil {
			log.Printf("%s: recovered emitting llvm ir: %s\n", file, err)
			return fmt.Errorf("compiler panic\n")
		}
		llvmIr := e.Module()

		fileName := fmt.Sprintf("./lltemp/%s.ll", file)
		llFile, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("%s: %w\n", file, err)
		}
		_, err = fmt.Fprintf(llFile, "%s", llvmIr)
		if err != nil {
			return fmt.Errorf("%s: %w\n", file, err)
		}
		err = llFile.Close()
		if err != nil {
			return fmt.Errorf("%s: %w\n", file, err)
		}

		llFiles = append(llFiles, fileName)

		for _, builtinModule := range e.BuiltinModules() {
			builtinModules[builtinModule] = struct{}{}
		}
	}

	for mod, _ := range builtinModules {
		// kinda dodgy fix but realistically its going to be one or two modules that vendor c stuff directly so no point bothering
		if mod == "ralloc.ll" {
			continue
		}
		modText, err := builtinFs.ReadFile(fmt.Sprintf("builtins/%s", mod))
		if err != nil {
			return fmt.Errorf("failed to read %s from builtin fs: %w\n", mod, err)
		}

		fileName := fmt.Sprintf("./lltemp/%s", mod)
		llFile, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w\n", llFile.Name(), err)
		}
		_, err = fmt.Fprintf(llFile, "%s", modText)
		if err != nil {
			return fmt.Errorf("failed to write %s: %w\n", llFile.Name(), err)
		}
		err = llFile.Close()
		if err != nil {
			return fmt.Errorf("failed to close %s: %w\n", llFile.Name(), err)
		}
		llFiles = append(llFiles, fileName)
	}

	llFiles = append(llFiles, "-o", "out")
	if !opts.NoExecBuild {
		cmd := exec.Command("clang", llFiles...)
		if opts.Dbg {
			fmt.Printf("executing: %s\n", strings.Join(cmd.Args, " "))
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error in clang exec, out: %s, err: %w", out, err)
		}
	}
	if !opts.KeepLL {
		if err := os.RemoveAll("./lltemp"); err != nil {
			return err
		}
	}

	return nil
}

func safeRun(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	fn()
	return nil
}
