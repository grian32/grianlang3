package cli

import (
	"fmt"
	"grianlang3/lexer"
	"grianlang3/util"
	"os/exec"
	"regexp"
	"strings"
)

type ExDefOpts struct {
	OutFile string
	InFile  string
}

var castRegexp = regexp.MustCompile(`\([a-zA-Z_ ]+\)`)

func RunExDef(opts *ExDefOpts) error {
	out, err := exec.Command("clang", "-dM", "-E", opts.InFile).CombinedOutput()
	if err != nil {
		return err
	}
	defines := make(map[string]string)
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
		defines[name] = rest
	}
	// TODO: strip casts, go off of value, float = float, integer = int64, strings=char* and char = char, special case void* 0 etc, output non regular expressions(A=B) as is will have to look up types though

	for n, v := range defines {
		// NOTE: doing it this way rather than stripping at readtime because in the future id like to type based off of casts or alternatively special case for example void ptr cast to 0 as a nullptr
		castStripped := stripCasts(v)
		fmt.Printf("%s = %s of type %s\n", n, castStripped, inferType(castStripped).String())
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

func inferType(val string) lexer.VarType {
	var currBaseType lexer.BaseVarType

	if strings.Contains(val, "\"") {
		return lexer.VarType{Base: lexer.Char, Pointer: 1}
	}

	if strings.Contains(val, "'") {
		return lexer.VarType{Base: lexer.Char}
	}

	var hasDigit, hasOther, hasDot bool

	for _, c := range val {
		if util.IsDigit(byte(c)) {
			hasDigit = true
		} else if c == '.' {
			hasDot = true
		} else {
			hasOther = true
		}
	}

	if hasDigit {
		currBaseType = lexer.Int
	}

	if hasDigit && hasDot {
		currBaseType = lexer.Float
	}

	if hasOther {
		currBaseType = lexer.None
	}

	return lexer.VarType{Base: currBaseType}
}
