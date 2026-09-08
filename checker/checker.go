package checker

import (
	"fmt"
	"gl3/checkedast"
	"gl3/lexer"
	"gl3/parser"
	"gl3/util"
)

type Scope struct {
	Symbols map[string]checkedast.Symbol
	Parent  *Scope
}

type Checker struct {
	structs   []checkedast.Struct
	functions []checkedast.Function
	globals   []checkedast.Global

	symbols      map[string]checkedast.Symbol
	currentScope *Scope

	diagnostics []Diagnostic
}

func NewChecker() *Checker {
	return &Checker{symbols: make(map[string]checkedast.Symbol)}
}

func (c *Checker) CheckProgram(program *parser.Program) (*checkedast.Program, []Diagnostic) /* todo: diag */ {
	if !c.assignIDs(program) {
		goto end
	}

	if !c.populateFieldsFunctions(program) {
		goto end
	}

	if !c.checkBodies(program) {
		goto end
	}

end:
	return &checkedast.Program{
		Structs:   c.structs,
		Functions: c.functions,
		Globals:   c.globals,
	}, c.diagnostics
}

func (c *Checker) checkBodies(node parser.Node) bool {
	return true
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
			Name:    node.Name,
			Id:      id,
			Opaque:  false,
			Private: false,
		})
		c.symbols[node.Name] = id
	case *parser.ExternStructStatement:
		if c.isSymbolDuplicate(node.Name, node.Position()) {
			return false
		}

		id := checkedast.StructID(len(c.structs))
		c.structs = append(c.structs, checkedast.Struct{
			Name:    node.Name,
			Id:      id,
			Opaque:  true,
			Private: node.Private,
		})
		c.symbols[node.Name] = id
	case *parser.FunctionStatement:
		if c.isSymbolDuplicate(node.Name.Value, node.Position()) {
			return false
		}

		id := checkedast.FunctionID(len(c.functions))
		c.functions = append(c.functions, checkedast.Function{
			Name:     node.Name.Value,
			Id:       id,
			External: false,
			Private:  node.Private,
		})
		c.symbols[node.Name.Value] = id
	case *parser.ExternFunctionStatement:
		if c.isSymbolDuplicate(node.Name, node.Position()) {
			return false
		}

		id := checkedast.FunctionID(len(c.functions))
		c.functions = append(c.functions, checkedast.Function{
			Name:     node.Name,
			Id:       id,
			External: true,
			Private:  node.Private,
		})
		c.symbols[node.Name] = id
	case *parser.DefStatement:
		if !node.Global {
			break
		}

		if c.isSymbolDuplicate(node.Name.Value, node.Position()) {
			return false
		}

		id := checkedast.GlobalID(len(c.globals))
		c.globals = append(c.globals, checkedast.Global{
			Name:     node.Name.Value,
			Id:       id,
			Constant: node.Constant,
		})
		c.symbols[node.Name.Value] = id
	}

	return true
}

func (c *Checker) populateFieldsFunctions(node parser.Node) bool {
	// TODO: After all fields are populated, recursively check sizing for globals
	// and defined-function signatures, including forward-declared structs that
	// contain opaque fields indirectly. Pointers remain sized regardless of pointee.
	switch node := node.(type) {
	case *parser.Program:
		valid := true
		for _, s := range node.Statements {
			if !c.populateFieldsFunctions(s) {
				valid = false
			}
		}

		return valid
	case *parser.FunctionStatement:
		funcSymbol, _ := c.symbols[node.Name.Value]
		funcId := funcSymbol.(checkedast.FunctionID)
		funcAst := &c.functions[funcId]

		retType, ok := c.functionRetType(node.Type, node.Position(), node.Name.Value, false)
		if !ok {
			return false
		}

		funcAst.ReturnType = retType
		return c.checkFunctionParams(node.Params, funcAst, node.Name.Value)
	case *parser.ExternFunctionStatement:
		funcSymbol, _ := c.symbols[node.Name]
		funcId := funcSymbol.(checkedast.FunctionID)
		funcAst := &c.functions[funcId]

		retType, ok := c.functionRetType(node.ReturnType, node.Position(), node.Name, true)
		if !ok {
			return false
		}

		funcAst.ReturnType = retType
		return c.checkFunctionParams(node.Params, funcAst, node.Name)
	case *parser.DefStatement:
		if !node.Global {
			break
		}
		globalSymbol, _ := c.symbols[node.Name.Value]
		globalAst := &c.globals[globalSymbol.(checkedast.GlobalID)]

		gt, ok := checkedast.ConvertVarType(node.Type, c.symbols)
		if !ok {
			c.appendDiagnostic(node.Name.Position(), "invalid type for global `%s`", node.Name.Value)
			return false
		}

		if gt.Base == checkedast.Void && gt.Pointer == 0 {
			c.appendDiagnostic(node.Name.Position(), "none type is not allowed on global definitions, global `%s`", node.Name.Value)
			return false
		}

		if gt.Base == checkedast.StructType && gt.Pointer == 0 && c.structs[gt.Struct].Opaque {
			c.appendDiagnostic(node.Name.Position(), "opaque struct by value is not allowed on global definitions, global `%s`", node.Name.Value)
			return false
		}

		globalAst.Type = gt
	case *parser.StructStatement:
		structSymbol, _ := c.symbols[node.Name]
		structAst := &c.structs[structSymbol.(checkedast.StructID)]

		fieldsSucceded := true

		structAst.FieldNames = make(map[string]int)

		for _, field := range node.Fields {
			if _, exists := structAst.FieldNames[field.Name.Value]; exists {
				c.appendDiagnostic(field.Name.Position(), "duplicate field `%s` on struct `%s`", field.Name.Value, node.Name)
				fieldsSucceded = false
				continue
			}

			ft, ok := checkedast.ConvertVarType(field.Type, c.symbols)
			if !ok {
				c.appendDiagnostic(field.Name.Position(), "invalid field type for field `%s` on struct `%s`", field.Name.Value, node.Name)
				fieldsSucceded = false
				continue
			}

			if ft.Base == checkedast.Void && ft.Pointer == 0 {
				c.appendDiagnostic(field.Name.Position(), "none type not allowed on fields, field `%s` on struct `%s`", field.Name.Value, node.Name)
				fieldsSucceded = false
				continue
			}

			structAst.FieldNames[field.Name.Value] = len(structAst.Fields)
			structAst.Fields = append(structAst.Fields, checkedast.TypedName{
				Name: field.Name.Value,
				Type: ft,
			})
		}

		return fieldsSucceded
	}

	return true
}

func (c *Checker) checkFunctionParams(params []parser.FunctionParameter, funcAst *checkedast.Function, funcName string) bool {
	funcAst.ParameterNames = make(map[string]int)
	paramsSucceded := true

	for _, param := range params {
		if _, exists := funcAst.ParameterNames[param.Name.Value]; exists {
			c.appendDiagnostic(param.Name.Position(), "duplicate parameter `%s` on function `%s`", param.Name.Value, funcName)
			paramsSucceded = false
			continue
		}

		pt, ok := checkedast.ConvertVarType(param.Type, c.symbols)
		if !ok {
			c.appendDiagnostic(param.Name.Position(), "invalid parameter type for parameter `%s` on function `%s`", param.Name.Value, funcName)
			paramsSucceded = false
			continue
		}
		if pt.Base == checkedast.Void && pt.Pointer == 0 {
			c.appendDiagnostic(param.Name.Position(), "none type is not allowed on parameters, param `%s` on function `%s`", param.Name.Value, funcName)
			paramsSucceded = false
			continue
		}
		if pt.Base == checkedast.StructType && pt.Pointer == 0 {
			stct := c.structs[pt.Struct]
			if stct.Opaque {
				c.appendDiagnostic(param.Name.Position(), "opaque struct value types are not allowed on parameters, param `%s` on function `%s`", param.Name.Value, funcName)
				paramsSucceded = false
				continue
			}
		}

		funcAst.ParameterNames[param.Name.Value] = len(funcAst.Parameters)
		funcAst.Parameters = append(funcAst.Parameters, checkedast.TypedName{
			Name: param.Name.Value,
			Type: pt,
		})
	}

	return paramsSucceded
}

func (c *Checker) functionRetType(retType lexer.VarType, position *util.Position, funcName string, extern bool) (checkedast.Type, bool) {
	rt, ok := checkedast.ConvertVarType(retType, c.symbols)
	if !ok {
		c.appendDiagnostic(position, "invalid return type for function `%s`", funcName)
		return checkedast.Type{}, false
	}

	if !extern && rt.Base == checkedast.StructType && rt.Pointer == 0 && c.structs[rt.Struct].Opaque {
		c.appendDiagnostic(position, "opaque struct value return type is not allowed for function `%s`", funcName)
		return checkedast.Type{}, false
	}

	return rt, true
}

func (c *Checker) isSymbolDuplicate(name string, pos *util.Position) bool {
	if _, exists := c.symbols[name]; exists {
		c.appendDiagnostic(pos, "duplicate symbol `%s`", name)
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
