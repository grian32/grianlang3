package checker

import (
	"fmt"
	"gl3/checkedast"
	"gl3/parser"
	"gl3/util"
)

type Checker struct {
	structs   []checkedast.Struct
	functions []checkedast.Function
	globals   []checkedast.Global

	symbols map[string]checkedast.Symbol

	diagnostics []Diagnostic
}

func NewChecker() *Checker {
	return &Checker{symbols: make(map[string]checkedast.Symbol)}
}

func (c *Checker) CheckProgram(program *parser.Program) (*checkedast.Program, []Diagnostic) /* todo: diag */ {
	if !c.assignIDs(program) {
		goto end
	}

end:
	return &checkedast.Program{
		Structs:   c.structs,
		Functions: c.functions,
		Globals:   c.globals,
	}, c.diagnostics
}

func (c *Checker) assignIDs(node parser.Node) bool {
	switch node := node.(type) {
	case *parser.Program:
		valid := true
		for _, s := range node.Statements {
			if !c.assignIDs(s) {
				valid = false
			}
		}

		return valid
	case *parser.StructStatement:
		if c.isSymbolDuplicate(node.Name, node.Position()) {
			return false
		}

		id := checkedast.StructID(len(c.structs))
		c.structs = append(c.structs, checkedast.Struct{
			Name: node.Name,
			Id:   id,
		})
		c.symbols[node.Name] = id
	case *parser.FunctionStatement:
		if c.isSymbolDuplicate(node.Name.Value, node.Position()) {
			return false
		}

		id := checkedast.FunctionID(len(c.functions))
		c.functions = append(c.functions, checkedast.Function{
			Name: node.Name.Value,
			Id:   id,
		})
		c.symbols[node.Name.Value] = id
	case *parser.DefStatement:
		if !node.Global {
			break
		}

		if c.isSymbolDuplicate(node.Name.Value, node.Position()) {
			return false
		}

		id := checkedast.GlobalID(len(c.globals))
		c.globals = append(c.globals, checkedast.Global{
			Name: node.Name.Value,
			Id:   id,
		})
		c.symbols[node.Name.Value] = id
	}

	return true
}

func (c *Checker) isSymbolDuplicate(name string, pos *util.Position) bool {
	if _, exists := c.symbols[name]; exists {
		c.appendDiagnostic(pos, "duplicate symbol %s", name)
		return true
	}

	return false
}

func (c *Checker) appendDiagnostic(pos *util.Position, msg string, v ...any) {
	c.diagnostics = append(c.diagnostics, Diagnostic{
		Message:  fmt.Sprintf(msg, v...),
		Position: pos,
	})
}
