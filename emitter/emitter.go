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
	m         *ir.Module
	entry     *ir.Block
	variables map[string]*ir.InstAlloca
	varTypes  map[string]types.Type
}

func New() *Emitter {
	e := &Emitter{m: ir.NewModule()}
	e.variables = make(map[string]*ir.InstAlloca)
	e.varTypes = make(map[string]types.Type)
	e.m.NewFunc("dbg_i64", types.Void, ir.NewParam("val", types.I64))
	main := e.m.NewFunc("main", types.I32)
	entry := main.NewBlock("")
	e.entry = entry
	return e
}

func (e *Emitter) Module() *ir.Module {
	return e.m
}

func (e *Emitter) Emit(node parser.Node) value.Value {
	switch node := node.(type) {
	case *parser.Program:
		var last value.Value
		for _, s := range node.Statements {
			last = e.Emit(s)
		}
		if last != nil {
			dbgFunc := e.m.Funcs[0]
			e.entry.NewCall(dbgFunc, last)
		}

		e.entry.NewRet(constant.NewInt(types.I32, 0))
		return last
	case *parser.ExpressionStatement:
		return e.Emit(node.Expression)
	case *parser.IntegerLiteral:
		return constant.NewInt(types.I64, node.Value)
	case *parser.InfixExpression:
		left := e.Emit(node.Left)
		right := e.Emit(node.Right)

		switch node.Operator {
		case "+":
			return e.entry.NewAdd(left, right)
		}
	case *parser.DefStatement:
		lt := varTypeToLlvm(node.Type)
		right := e.Emit(node.Right)

		vPtr := e.entry.NewAlloca(lt)
		e.variables[node.Name.Value] = vPtr
		e.varTypes[node.Name.Value] = lt
		e.entry.NewStore(right, vPtr)
		return right
	case *parser.AssignmentStatement:
		vPtr, ok := e.variables[node.Name.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find variable of name %s used in var assignment", node.Name)
		}
		right := e.Emit(node.Right)
		e.entry.NewStore(right, vPtr)
		return right
	case *parser.IdentifierExpression:
		vPtr, ok := e.variables[node.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find variable of name %s used in var ref", node.Value)
		}
		vType, ok := e.varTypes[node.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find variable type of name %s used in var ref", node.Value)
		}
		return e.entry.NewLoad(vType, vPtr)
	}

	return nil
}

func varTypeToLlvm(vt lexer.VarType) types.Type {
	switch vt {
	case lexer.None:
		return nil
	case lexer.Int:
		return types.I64
	}
	return nil
}
