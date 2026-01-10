package emitter

import (
	"fmt"
	"grianlang3/lexer"
	"grianlang3/parser"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
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
	fnc = e.m.NewFunc("dbg_float", types.Void, ir.NewParam("val", types.Float))
	e.functions["dbg_float"] = fnc
	fnc = e.m.NewFunc("dbg_bool", types.Void, ir.NewParam("val", types.I64Ptr))
	e.functions["dbg_bool"] = fnc
	//fnc = e.m.NewFunc("malloc", types.I32Ptr, ir.NewParam("val", types.I64Ptr))
	//e.functions["malloc"] = fnc
	return e
}

func (e *Emitter) Module() *ir.Module {
	return e.m
}

// TODO: maybe look into ditching this, serves its purposes, but feels a bit wasteful
var infixIntOpTypes = map[types.Type]struct{}{
	types.I64: {},
	types.I32: {},
}

var llvmIntTypes = map[types.Type]struct{}{
	types.I1:  {},
	types.I32: {},
	types.I64: {},
}

var varTypeIntTypes = map[lexer.BaseVarType]struct{}{
	lexer.Bool:  {},
	lexer.Int:   {},
	lexer.Int32: {},
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
	case *parser.BooleanExpression:
		return constant.NewInt(types.I1, boolToI1(node.Value))
	case *parser.FloatLiteral:
		return constant.NewFloat(types.Float, float64(node.Value))
	case *parser.InfixExpression:
		left := e.Emit(node.Left, entry)
		right := e.Emit(node.Right, entry)

		leftType := left.Type()
		rightType := right.Type()

		_, leftIntOk := infixIntOpTypes[leftType]
		_, rightIntOk := infixIntOpTypes[rightType]

		// TODO: extend when doubles etc
		_, leftFloatOk := leftType.(*types.FloatType)
		_, rightFloatOk := rightType.(*types.FloatType)

		if ptr, ok := leftType.(*types.PointerType); ok && rightIntOk {
			if node.Operator == "+" {
				return entry.NewGetElementPtr(ptr.ElemType, left, right)
			} else if node.Operator == "-" {
				intType := rightType.(*types.IntType)
				zero := constant.NewInt(intType, 0)
				negRight := entry.NewSub(zero, right)
				return entry.NewGetElementPtr(ptr.ElemType, left, negRight)
			}
		}

		if leftIntOk && rightIntOk && leftType == rightType {
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
			case "<":
				return entry.NewICmp(enum.IPredSLT, left, right)
			case ">":
				return entry.NewICmp(enum.IPredSGT, left, right)
			case "<=":
				return entry.NewICmp(enum.IPredSLE, left, right)
			case ">=":
				return entry.NewICmp(enum.IPredSGE, left, right)
			case "==":
				return entry.NewICmp(enum.IPredEQ, left, right)
			case "!=":
				return entry.NewICmp(enum.IPredNE, left, right)
			}
		}

		if leftFloatOk && rightFloatOk {
			switch node.Operator {
			case "+":
				return entry.NewFAdd(left, right)
			case "-":
				return entry.NewFSub(left, right)
			case "*":
				return entry.NewFMul(left, right)
			case "/":
				// TODO: def behaviour for /0, intmin/-1
				return entry.NewFDiv(left, right)
			case "<":
				return entry.NewFCmp(enum.FPredOLT, left, right)
			case ">":
				return entry.NewFCmp(enum.FPredOGT, left, right)
			case "<=":
				return entry.NewFCmp(enum.FPredOLE, left, right)
			case ">=":
				return entry.NewFCmp(enum.FPredOGE, left, right)
			case "==":
				return entry.NewFCmp(enum.FPredOEQ, left, right)
			case "!=":
				return entry.NewFCmp(enum.FPredONE, left, right)
			}
		}

		if leftType == types.I1 && rightType == types.I1 {
			switch node.Operator {
			case "&&":
				return entry.NewAnd(left, right)
			case "||":
				return entry.NewOr(left, right)
			}
		}

		fmt.Printf("compile error: operator %s invalid for types %T(%s), %T(%s)", node.Operator, node.Left, node.Left.String(), node.Right, node.Right.String())
		return nil
	case *parser.PrefixExpression:
		switch node.Operator {
		case "!":
			right := e.Emit(node.Right, entry)
			// TODO: check right == i1
			trueVal := constant.NewInt(types.I1, 1)
			return entry.NewXor(right, trueVal)
		case "-":
			right := e.Emit(node.Right, entry)
			_, rightIntOk := llvmIntTypes[right.Type()]
			_, rightFloatOk := right.Type().(*types.FloatType)
			if rightIntOk {
				zero := constant.NewInt(right.Type().(*types.IntType), 0)
				return entry.NewSub(zero, right)
			} else if rightFloatOk {
				zero := constant.NewFloat(types.Float, 0)
				return entry.NewFSub(zero, right)
			}
		}
	case *parser.DefStatement:
		lt := varTypeToLlvm(node.Type)
		right := e.Emit(node.Right, entry)

		vPtr := entry.NewAlloca(lt)
		e.variables[node.Name.Value] = vPtr
		e.varTypes[node.Name.Value] = lt
		entry.NewStore(right, vPtr)
		return right
	case *parser.AssignmentExpression:
		if ident, ok := node.Left.(*parser.IdentifierExpression); ok {
			vPtr, ok := e.variables[ident.Value]
			if !ok {
				fmt.Printf("compile error: couldn't find variable of name %s used in var assignment", ident.Value)
			}
			right := e.Emit(node.Right, entry)
			entry.NewStore(right, vPtr)
			return right
		} else if _, ok := node.Left.(*parser.DereferenceExpression); ok {
			ptr := e.emitAddress(node.Left, entry)
			right := e.Emit(node.Right, entry)
			entry.NewStore(right, ptr)
			return right
		}

		return nil
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
	case *parser.ReferenceExpression:
		vPtr, ok := e.variables[node.Var.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find variable with name %s in reference expr", node.Var.Value)
		}
		return vPtr
	case *parser.DereferenceExpression:
		ptr := e.Emit(node.Var, entry)

		ptrTy, ok := ptr.Type().(*types.PointerType)
		if !ok {
			fmt.Printf("compile error: cannot deref non-ptr type %v\n", ptrTy)
		}

		return entry.NewLoad(ptrTy.ElemType, ptr)
	case *parser.CastExpression:
		src := e.Emit(node.Expr, entry)
		srcType := src.Type()
		dstType := varTypeToLlvm(node.Type)
		// TODO: could be faster if bare bool comparison? prob
		_, leftIntOk := llvmIntTypes[srcType]
		_, rightIntOk := varTypeIntTypes[node.Type.Base]

		_, leftFloatOk := src.Type().(*types.FloatType)
		rightFloatOk := node.Type.Base == lexer.Float && node.Type.Pointer == 0

		_, leftPtrOk := src.Type().(*types.PointerType)
		rightPtrOk := node.Type.Pointer > 0

		if leftIntOk && rightIntOk && node.Type.Pointer == 0 {
			if srcType == types.I1 {
				return entry.NewZExt(src, dstType)
			}

			dstSize := getSizeForVarType(node.Type)
			srcSize := getSizeForLlvmType(src.Type())

			if srcSize < dstSize {
				return entry.NewSExt(src, dstType)
			} else if srcSize > dstSize {
				return entry.NewTrunc(src, dstType)
			} else {
				// same size, no cast necessary
				return src
			}
		} else if leftIntOk && node.Type.Pointer > 0 {
			return entry.NewIntToPtr(src, dstType)
		} else if _, ok := src.Type().(*types.PointerType); ok && rightIntOk && node.Type.Pointer == 0 {
			if node.Type.Base != lexer.Int {
				// non 64 bit which is llvm default on most (i.e 64bit) systems
				fmt.Printf("compile warning: pointer to int cast may truncate")
			}
			return entry.NewPtrToInt(src, varTypeToLlvm(node.Type))
		} else if leftIntOk && rightFloatOk {
			return entry.NewSIToFP(src, dstType)
		} else if leftFloatOk && rightIntOk {
			return entry.NewFPToSI(src, dstType)
		} else if leftPtrOk && rightPtrOk {
			/** c bitcast behaviour
			  int x = 1073741941; // 0100 0000 0000 0000 0000 0000 0111 0101 = 1073741941; as float = 2.somethingsomething
			  int* xx = &x;
			  float* fx = (float*)xx;
			  printf("%.100f", *fx); == 2.somethingsomething
			  return 0;
			*/
			return entry.NewBitCast(src, dstType)
		}

	}

	return nil
}

// emitAdress literally only necessary because i need the vptr from ident expr, deref expr is same as e.emit lol
func (e *Emitter) emitAddress(node parser.Node, entry *ir.Block) value.Value {
	switch node := node.(type) {
	case *parser.IdentifierExpression:
		if param, ok := e.parameters[node.Value]; ok {
			return param
		}
		vPtr, ok := e.variables[node.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find variable with name %s in deref assignment", node.Value)
		}
		return vPtr
	case *parser.DereferenceExpression:
		ptr := e.Emit(node.Var, entry)
		return ptr
	default:
		fmt.Printf("compile error: invalid node type for emitAddress\n")
		return nil
	}
}

func varTypeToLlvm(vt lexer.VarType) types.Type {
	var baseType types.Type
	switch vt.Base {
	case lexer.None:
		baseType = nil
	case lexer.Int:
		baseType = types.I64
	case lexer.Int32:
		baseType = types.I32
	case lexer.Void:
		baseType = types.Void
	case lexer.Bool:
		baseType = types.I1
	case lexer.Float:
		baseType = types.Float
	}

	if baseType != nil {
		for _ = range vt.Pointer {
			baseType = &types.PointerType{ElemType: baseType}
		}
	}

	return baseType
}

func getSizeForVarType(vt lexer.VarType) int64 {
	if vt.Pointer > 0 {
		return 8
	}
	switch vt.Base {
	case lexer.Bool:
		return 1
	case lexer.Int32, lexer.Float:
		return 4
	case lexer.Int:
		return 8
	}

	return 0
}

func getSizeForLlvmType(lt types.Type) int64 {
	switch lt := lt.(type) {
	case *types.IntType:
		if lt.BitSize < 8 {
			return 1
		}
		return int64(lt.BitSize / 8)
	case *types.PointerType:
		return 8
	}

	return 0
}

// returns a i64 as thats what the emit integer const function expects, means one less conv
func boolToI1(b bool) int64 {
	if b {
		return 1
	}

	return 0
}
