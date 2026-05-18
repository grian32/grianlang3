package emitter

import (
	"grianlang3/lexer"
	"grianlang3/parser"
)

type Declare struct {
	Name       string
	ReturnType lexer.VarType
	ParamTypes []lexer.VarType
}

type StructDefinition struct {
	Name       string
	Fields     []lexer.VarType
	FieldNames map[string]int
}

type GlobalDefinition struct {
	Name string
	Type lexer.VarType
}

type importParser struct {
	declares    []Declare
	structDefns []StructDefinition
	globalDefns []GlobalDefinition
}

func (ip *importParser) findImports(node parser.Node) {
	switch node := node.(type) {
	case *parser.Program:
		for _, s := range node.Statements {
			ip.findImports(s)
		}
	case *parser.FunctionStatement:
		var paramTypes []lexer.VarType
		for _, p := range node.Params {
			paramTypes = append(paramTypes, p.Type)
		}

		ip.declares = append(ip.declares, Declare{
			Name:       node.Name.Value,
			ReturnType: node.Type,
			ParamTypes: paramTypes,
		})
	case *parser.StructStatement:
		// this is kinda stupid but passing around the struct stmt is icky
		ip.structDefns = append(ip.structDefns, StructDefinition{
			Name:       node.Name,
			Fields:     node.Types,
			FieldNames: node.Names,
		})
	case *parser.DefStatement:
		if node.Constant && node.Global {
			ip.globalDefns = append(ip.globalDefns, GlobalDefinition{
				Name: node.Name.Value,
				Type: node.Type,
			})
		}
	}
}

func findDeclares(file string) ([]Declare, []StructDefinition, []GlobalDefinition) {
	l := lexer.New(file)
	p := parser.New(l)
	program := p.ParseProgram()
	ip := importParser{}
	ip.findImports(program)
	return ip.declares, ip.structDefns, ip.globalDefns
}
