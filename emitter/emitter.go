package emitter

import (
	"fmt"
	"grianlang3/lexer"
	"grianlang3/parser"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type Emitter struct {
	m *ir.Module
	// var, vartypes, params get reset after each function is emitted.
	variables  map[string]*ir.InstAlloca
	varTypes   map[string]types.Type
	parameters map[string]*ir.Param

	functions      map[string]*ir.Func
	functionBlocks map[string]*ir.Block
}

func New() *Emitter {
	e := &Emitter{m: ir.NewModule()}
	e.variables = make(map[string]*ir.InstAlloca)
	e.varTypes = make(map[string]types.Type)
	e.functions = make(map[string]*ir.Func)
	e.functionBlocks = make(map[string]*ir.Block)
	e.parameters = make(map[string]*ir.Param)

	fnc := e.m.NewFunc("dbg_i64", types.Void, ir.NewParam("val", types.I64))
	e.functions["dbg_i64"] = fnc
	fnc = e.m.NewFunc("dbg_i32", types.Void, ir.NewParam("val", types.I64))
	e.functions["dbg_i32"] = fnc
	// main := e.m.NewFunc("main", types.I32)
	// entry := main.NewBlock("")
	// e.entry = entry
	return e
}

func (e *Emitter) Module() *ir.Module {
	return e.m
}

func (e *Emitter) Emit(node parser.Node, entry *ir.Block) value.Value {
	switch node := node.(type) {
	case *parser.Program:
		var last value.Value
		for _, s := range node.Statements {
			last = e.Emit(s, entry)
		}

		return last
	case *parser.ExpressionStatement:
		return e.Emit(node.Expression, entry)
	case *parser.IntegerLiteral:
		// sorta unsafe cast i think, will crash tho if incompatible so same behaviour? maybe better to have an error msg tho
		return constant.NewInt(varTypeToLlvm(node.Type).(*types.IntType), node.Value)
	case *parser.InfixExpression:
		left := e.Emit(node.Left, entry)
		right := e.Emit(node.Right, entry)

		switch node.Operator {
		case "+":
			return entry.NewAdd(left, right)
		case "-":
			return entry.NewSub(left, right)
		case "*":
			return entry.NewMul(left, right)
		case "/":
			// TODO: def behaviour for /0, intmin/-1
			return entry.NewSDiv(left, right)
		}
	case *parser.DefStatement:
		lt := varTypeToLlvm(node.Type)
		right := e.Emit(node.Right, entry)

		vPtr := entry.NewAlloca(lt)
		e.variables[node.Name.Value] = vPtr
		e.varTypes[node.Name.Value] = lt
		entry.NewStore(right, vPtr)
		return right
	case *parser.AssignmentStatement:
		vPtr, ok := e.variables[node.Name.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find variable of name %s used in var assignment", node.Name.Value)
		}
		right := e.Emit(node.Right, entry)
		entry.NewStore(right, vPtr)
		return right
	case *parser.IdentifierExpression:
		if param, ok := e.parameters[node.Value]; ok {
			return param
		}

		vPtr, ok := e.variables[node.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find variable of name %s used in var ref", node.Value)
		}
		vType, ok := e.varTypes[node.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find variable type of name %s used in var ref", node.Value)
		}
		return entry.NewLoad(vType, vPtr)
	case *parser.CallExpression:
		var args []value.Value

		for _, a := range node.Params {
			args = append(args, e.Emit(a, entry))
		}

		fncPtr, ok := e.functions[node.Function.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find function with name %s", node.Function.Value)
		}

		return entry.NewCall(fncPtr, args...)
	case *parser.FunctionStatement:
		retType := varTypeToLlvm(node.Type)
		var paramTypes []*ir.Param

		for _, p := range node.Params {
			irParam := ir.NewParam(
				p.Name.Value,
				varTypeToLlvm(p.Type),
			)
			e.parameters[p.Name.Value] = irParam
			paramTypes = append(paramTypes, irParam)
		}

		fncPtr := e.m.NewFunc(node.Name.Value, retType, paramTypes...)
		e.functions[node.Name.Value] = fncPtr
		block := fncPtr.NewBlock("")
		e.functionBlocks[node.Name.Value] = block

		foundRet := false

		for _, s := range node.Body.Statements {
			// ehh???
			if _, ok := s.(*parser.ReturnStatement); !foundRet && ok {
				foundRet = true
			}
			e.Emit(s, block)
		}

		if !foundRet {
			block.NewRet(nil)
		}

		e.parameters = make(map[string]*ir.Param)
		e.variables = make(map[string]*ir.InstAlloca)
		e.varTypes = make(map[string]types.Type)
		return fncPtr
	case *parser.ReturnStatement:
		val := e.Emit(node.Expr, entry)
		entry.NewRet(val)
		return val
	}

	return nil
}

func varTypeToLlvm(vt lexer.VarType) types.Type {
	switch vt {
	case lexer.None:
		return nil
	case lexer.Int:
		return types.I64
	case lexer.Int32:
		return types.I32
	case lexer.Void:
		return types.Void
	}
	return nil
}
