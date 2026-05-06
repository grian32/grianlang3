package cli

import (
	"fmt"
	"os/exec"
	"strings"
)

type ExDefOpts struct {
	OutFile string
	InFile  string
}

func RunExDef(opts *ExDefOpts) error {
	out, err := exec.Command("clang", "-dM", "-E", opts.InFile).CombinedOutput()
	if err != nil {
		return err
	}
	defines := make(map[string]string)
	for _, l := range strings.Split(string(out), "\n") {
		split := strings.Split(l, " ")
		if len(split) < 2 {
			continue
		}
		name := split[1]
		if strings.HasPrefix(name, "_") || strings.Contains(name, "(") {
			continue
		}
		rest := strings.Trim(strings.Join(split[2:], " "), " ")
		if rest == "" {
			continue
		}
		defines[name] = rest
	}
	// TODO: strip casts, go off of value, float = float, integer = int64, strings=char* and char = char, special case void* 0 etc, output non regular expressions(A=B) as is will have to look up types though

	for n, v := range defines {
		fmt.Printf("%s = %s\n", n, v)
	}
	return nil
}
