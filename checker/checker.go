package checker

import (
	"gl3/checkedast"
	"gl3/parser"
)

type Checker struct {
	structs   []checkedast.Struct
	functions []checkedast.Function
	globals   []checkedast.Global

	symbols map[string]checkedast.Symbol
}

func (c *Checker) CheckProgram(program *parser.Program) (*checkedast.Program, []Diagnostic) /* todo: diag */ {
	c.assignIDs(program)

	return &checkedast.Program{}, nil
}

func (c *Checker) assignIDs(program *parser.Program) {

}
