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

type ImportParser struct {
	declares    []Declare
	structDefns []StructDefinition
	globalDefns []GlobalDefinition
	// just the name since its all that it uses
	externStructDefns []string
}

func (ip *ImportParser) findImports(node parser.Node) {
	switch node := node.(type) {
	case *parser.Program:
		for _, s := range node.Statements {
			ip.findImports(s)
		}
	case *parser.FunctionStatement:
		if node.Private {
			break
		}

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
	case *parser.ExternFunctionStatement:
		if node.Private {
			break
		}

		var paramTypes []lexer.VarType
		for _, p := range node.Params {
			paramTypes = append(paramTypes, p.Type)
		}

		ip.declares = append(ip.declares, Declare{
			Name:       node.Name,
			ReturnType: node.ReturnType,
			ParamTypes: paramTypes,
		})
	case *parser.ExternStructStatement:
		if node.Private {
			break
		}
		ip.externStructDefns = append(ip.externStructDefns, node.Name)
	}
}

func findDeclares(file string) *ImportParser {
	l := lexer.New(file)
	p := parser.New(l)
	program := p.ParseProgram()
	ip := ImportParser{}
	ip.findImports(program)
	return &ip
}
