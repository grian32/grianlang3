package cli

import (
	"fmt"
	"grianlang3/lexer"
	"grianlang3/util"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type ExDefOpts struct {
	OutFile string
	InFile  string
}

var castRegexp = regexp.MustCompile(`\([a-zA-Z_\* ]+\)`)

func RunExDef(opts *ExDefOpts) error {
	out, err := exec.Command("clang", "-dM", "-E", opts.InFile).CombinedOutput()
	if err != nil {
		return err
	}
	defines := util.NewOrderedMap[string, string]()
	for l := range strings.SplitSeq(string(out), "\n") {
		split := strings.Split(l, " ")
		if len(split) < 2 {
			continue
		}
		name := split[1]
		if strings.HasPrefix(name, "_") || strings.Contains(name, "(") {
			continue
		}
		rest := strings.TrimSpace(strings.Join(split[2:], " "))
		if rest == "" {
			continue
		}
		defines.Set(name, rest)
	}
	var gl3Out strings.Builder

	defines.Range(func(n, v string) {
		resolveDefine(n, defines)
	})

	var rangeErr error
	defines.Range(func(n, v string) {
		// linux/unix are compiler headers and shouldn't be included, need to look for similar ones in other OS
		if n == "linux" || n == "unix" {
			return
		}

		// NOTE: doing it this way rather than stripping at readtime because in the future id like to type based off of casts or alternatively special case for example void ptr cast to 0 as a nullptr
		castStripped := stripCasts(v)
		_, err := fmt.Fprintf(&gl3Out, "global const %s %s = %s\n", getTypeString(inferType(castStripped)), n, castStripped)
		if err != nil {
			rangeErr = err
			return
		}
	})
	if rangeErr != nil {
		return rangeErr
	}

	_, err = os.Stat(opts.OutFile)
	if err == nil {
		return fmt.Errorf("File %s aready exists", opts.OutFile)
	}

	file, err := os.Create(opts.OutFile)
	if err != nil {
		return err
	}
	_, err = file.WriteString(gl3Out.String())
	if err != nil {
		return err
	}
	return nil
}

func stripCasts(val string) string {
	return strings.TrimSpace(
		strings.TrimSuffix(
			strings.TrimPrefix(
				castRegexp.ReplaceAllString(val, ""),
				"("),
			")"),
	)
}

// this will resolve f.e #define A B or #define A "whatever" or #define A (OTHER / 5000)
func resolveDefine(name string, defines *util.OrderedMap[string, string]) {
	def, _ := defines.Get(name)
	parts := strings.Fields(def)
	var newV strings.Builder

	for _, p := range parts {
		stripped := stripCasts(p)
		if foundValue, ok := defines.Get(stripped); ok {
			newV.WriteString(foundValue)
		} else {
			newV.WriteString(stripped)
		}
	}
	defines.Set(name, strings.ReplaceAll(strings.TrimSpace(newV.String()), `""`, ``))

	def, _ = defines.Get(name)
	// if naive infer is string or some bullshit its prob busted but otherwise we can assume its a math op etc
	if inferType(def).Base == lexer.None {
		resolveDefine(name, defines)
	}
}

func inferType(val string) lexer.VarType {
	if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") && strings.Count(val, "\"") == 2 {
		return lexer.VarType{Base: lexer.Char, Pointer: 1}
	}

	if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") && strings.Count(val, "'") == 2 {
		return lexer.VarType{Base: lexer.Char}
	}

	if strings.ContainsAny(val, " \t") {
		return lexer.VarType{Base: lexer.None}
	}

	var hasDigit, hasOther, hasDot bool
	var otherFound string

	for _, c := range val {
		if util.IsDigit(byte(c)) {
			hasDigit = true
		} else if c == '.' {
			hasDot = true
		} else {
			hasOther = true
			otherFound += string(c)
		}
	}

	// support math ops and such
	if hasOther && hasDigit && util.ContainsOnly(otherFound, "+-*/") {
		return lexer.VarType{Base: lexer.Int}
	}

	if hasOther || !hasDigit {
		return lexer.VarType{Base: lexer.None}
	}

	if hasDot {
		return lexer.VarType{Base: lexer.Float}
	}

	return lexer.VarType{Base: lexer.Int}
}

func getTypeString(typ lexer.VarType) string {
	switch typ.Base {
	case lexer.Int:
		return "int"
	case lexer.Float:
		return "float"
	case lexer.Char:
		if typ.Pointer == 1 {
			return "char*"
		} else {
			return "char"
		}
	}

	return ""
}
