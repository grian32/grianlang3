package emitter

import (
	"grianlang3/parser"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type Emitter struct {
	m     *ir.Module
	entry *ir.Block
}

func New() *Emitter {
	e := &Emitter{m: ir.NewModule()}
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
	}

	return nil
}
