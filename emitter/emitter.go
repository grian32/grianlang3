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
	variables         map[string]*ir.InstAlloca
	varTypes          map[string]types.Type
	varGlTypes        map[string]lexer.VarType
	parameters        map[string]*ir.Param
	parametersGlTypes map[string]lexer.VarType

	functions             map[string]*ir.Func
	functionBlocks        map[string]*ir.Block
	functionGlReturnTypes map[string]lexer.VarType
}

func New() *Emitter {
	e := &Emitter{m: ir.NewModule()}
	e.variables = make(map[string]*ir.InstAlloca)
	e.varTypes = make(map[string]types.Type)
	e.functions = make(map[string]*ir.Func)
	e.functionBlocks = make(map[string]*ir.Block)
	e.parameters = make(map[string]*ir.Param)
	e.functionGlReturnTypes = make(map[string]lexer.VarType)
	e.parametersGlTypes = make(map[string]lexer.VarType)
	e.varGlTypes = make(map[string]lexer.VarType)

	AddBuiltins(e)
	//fnc = e.m.NewFunc("malloc", types.I32Ptr, ir.NewParam("val", types.I64Ptr))
	//e.functions["malloc"] = fnc
	return e
}

func (e *Emitter) Module() *ir.Module {
	return e.m
}

// TODO: maybe look into ditching this, serves its purposes, but feels a bit wasteful
var infixIntOpTypes = map[lexer.BaseVarType]struct{}{
	lexer.Int8:   {},
	lexer.Int16:  {},
	lexer.Int32:  {},
	lexer.Int:    {},
	lexer.Uint8:  {},
	lexer.Uint16: {},
	lexer.Uint32: {},
	lexer.Uint:   {},
}

var llvmIntTypes = map[types.Type]struct{}{
	types.I1:  {},
	types.I8:  {},
	types.I16: {},
	types.I32: {},
	types.I64: {},
}

var varTypeIntTypes = map[lexer.BaseVarType]struct{}{
	lexer.Bool:  {},
	lexer.Int8:  {},
	lexer.Int16: {},
	lexer.Int32: {},
	lexer.Int:   {},
}

var glTypeSInts = map[lexer.BaseVarType]struct{}{
	lexer.Int8:  {},
	lexer.Int16: {},
	lexer.Int32: {},
	lexer.Int:   {},
}

var glTypeUInts = map[lexer.BaseVarType]struct{}{
	lexer.Uint8:  {},
	lexer.Uint16: {},
	lexer.Uint32: {},
	lexer.Uint:   {},
}

func (e *Emitter) Emit(node parser.Node, entry *ir.Block) (value.Value, lexer.VarType) {
	switch node := node.(type) {
	case *parser.Program:
		var last value.Value
		var lastType lexer.VarType
		for _, s := range node.Statements {
			last, lastType = e.Emit(s, entry)
		}

		return last, lastType
	case *parser.ExpressionStatement:
		return e.Emit(node.Expression, entry)
	case *parser.IntegerLiteral:
		/**
		fnc main() -> int32 {
			    def uint x = 18446744073709551616u64;
			    dbg_u64(x);
			    return 0i32;
			}
		NOTE: vaguely strange behaviour on such input, seems to always trunc to -1 even if ...17 or ...18 so on?
		maybe llvm quirk, not sure, don't think there's much more i can do as far as my handling of ints goes, not like
		go has uint128
		*/
		if _, ok := glTypeSInts[node.Type.Base]; ok {
			return constant.NewInt(varTypeToLlvm(node.Type).(*types.IntType), node.Value), node.Type
		} else if _, ok := glTypeUInts[node.Type.Base]; ok {
			return constant.NewInt(varTypeToLlvm(node.Type).(*types.IntType), int64(node.UValue)), node.Type
		}
	case *parser.BooleanExpression:
		return constant.NewInt(types.I1, boolToI1(node.Value)), lexer.VarType{Base: lexer.Bool, Pointer: 0}
	case *parser.FloatLiteral:
		return constant.NewFloat(types.Float, float64(node.Value)), lexer.VarType{Base: lexer.Float, Pointer: 0}
	case *parser.InfixExpression:
		// TODO: clean this pile of shit up
		left, leftVt := e.Emit(node.Left, entry)
		right, rightVt := e.Emit(node.Right, entry)

		leftType := left.Type()
		rightType := right.Type()

		_, leftIntBaseOk := infixIntOpTypes[leftVt.Base]
		_, rightIntBaseOk := infixIntOpTypes[rightVt.Base]
		leftIntOk := leftIntBaseOk && leftVt.Pointer == 0
		rightIntOk := rightIntBaseOk && rightVt.Pointer == 0

		// TODO: extend when doubles etc
		_, leftFloatOk := leftType.(*types.FloatType)
		_, rightFloatOk := rightType.(*types.FloatType)

		if ptr, ok := leftType.(*types.PointerType); ok && rightIntOk {
			if node.Operator == "+" {
				return entry.NewGetElementPtr(ptr.ElemType, left, right), leftVt
			} else if node.Operator == "-" {
				intType := rightType.(*types.IntType)
				zero := constant.NewInt(intType, 0)
				negRight := entry.NewSub(zero, right)
				return entry.NewGetElementPtr(ptr.ElemType, left, negRight), leftVt
			}
		}

		if leftIntOk && rightIntOk && leftVt == rightVt {
			switch node.Operator {
			case "+":
				return entry.NewAdd(left, right), leftVt
			case "-":
				return entry.NewSub(left, right), leftVt
			case "*":
				return entry.NewMul(left, right), leftVt
			case "==":
				return entry.NewICmp(enum.IPredEQ, left, right), leftVt
			case "!=":
				return entry.NewICmp(enum.IPredNE, left, right), leftVt
			}

			if _, ok := glTypeUInts[leftVt.Base]; ok {
				switch node.Operator {
				case "/":
					// TODO: def behaviour for /0, intmin/-1
					return entry.NewUDiv(left, right), leftVt
				case "<":
					return entry.NewICmp(enum.IPredULT, left, right), leftVt
				case ">":
					return entry.NewICmp(enum.IPredUGT, left, right), leftVt
				case "<=":
					return entry.NewICmp(enum.IPredULE, left, right), leftVt
				case ">=":
					return entry.NewICmp(enum.IPredUGE, left, right), leftVt
				}
			} else if _, ok := glTypeSInts[leftVt.Base]; ok {
				switch node.Operator {
				case "/":
					// TODO: def behaviour for /0, intmin/-1
					return entry.NewSDiv(left, right), leftVt
				case "<":
					return entry.NewICmp(enum.IPredSGT, left, right), leftVt
				case ">":
					return entry.NewICmp(enum.IPredSGT, left, right), leftVt
				case "<=":
					return entry.NewICmp(enum.IPredSLE, left, right), leftVt
				case ">=":
					return entry.NewICmp(enum.IPredSGE, left, right), leftVt
				}
			}
		}

		if leftFloatOk && rightFloatOk {
			switch node.Operator {
			case "+":
				return entry.NewFAdd(left, right), leftVt
			case "-":
				return entry.NewFSub(left, right), leftVt
			case "*":
				return entry.NewFMul(left, right), leftVt
			case "/":
				// TODO: def behaviour for /0, intmin/-1
				return entry.NewFDiv(left, right), leftVt
			case "<":
				return entry.NewFCmp(enum.FPredOLT, left, right), leftVt
			case ">":
				return entry.NewFCmp(enum.FPredOGT, left, right), leftVt
			case "<=":
				return entry.NewFCmp(enum.FPredOLE, left, right), leftVt
			case ">=":
				return entry.NewFCmp(enum.FPredOGE, left, right), leftVt
			case "==":
				return entry.NewFCmp(enum.FPredOEQ, left, right), leftVt
			case "!=":
				return entry.NewFCmp(enum.FPredONE, left, right), leftVt
			}
		}

		if leftType == types.I1 && rightType == types.I1 {
			switch node.Operator {
			case "&&":
				return entry.NewAnd(left, right), leftVt
			case "||":
				return entry.NewOr(left, right), leftVt
			}
		}

		fmt.Printf("compile error: operator %s invalid for types %T(%s), %T(%s)", node.Operator, node.Left, node.Left.String(), node.Right, node.Right.String())
	case *parser.PrefixExpression:
		switch node.Operator {
		case "!":
			right, rt := e.Emit(node.Right, entry)
			// TODO: check right == i1
			trueVal := constant.NewInt(types.I1, 1)
			return entry.NewXor(right, trueVal), rt
		case "-":
			right, rt := e.Emit(node.Right, entry)
			_, rightIntOk := glTypeSInts[rt.Base]
			_, rightFloatOk := right.Type().(*types.FloatType)
			if rightIntOk {
				zero := constant.NewInt(right.Type().(*types.IntType), 0)
				return entry.NewSub(zero, right), rt
			} else if rightFloatOk {
				zero := constant.NewFloat(types.Float, 0)
				return entry.NewFSub(zero, right), rt
			}
		}
	case *parser.DefStatement:
		lt := varTypeToLlvm(node.Type)
		right, vt := e.Emit(node.Right, entry)

		vPtr := entry.NewAlloca(lt)
		e.variables[node.Name.Value] = vPtr
		e.varTypes[node.Name.Value] = lt
		e.varGlTypes[node.Name.Value] = vt
		entry.NewStore(right, vPtr)
		return right, vt
	case *parser.AssignmentExpression:
		if ident, ok := node.Left.(*parser.IdentifierExpression); ok {
			vPtr, ok := e.variables[ident.Value]
			if !ok {
				fmt.Printf("compile error: couldn't find variable of name %s used in var assignment", ident.Value)
			}
			right, vt := e.Emit(node.Right, entry)
			entry.NewStore(right, vPtr)
			return right, vt
		} else if _, ok := node.Left.(*parser.DereferenceExpression); ok {
			ptr, _ := e.emitAddress(node.Left, entry)
			right, vt := e.Emit(node.Right, entry)
			entry.NewStore(right, ptr)
			return right, vt
		}
	case *parser.IdentifierExpression:
		if param, ok := e.parameters[node.Value]; ok {
			return param, e.parametersGlTypes[node.Value]
		}

		vPtr, ok := e.variables[node.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find variable of name %s used in var ref", node.Value)
		}
		vType, ok := e.varTypes[node.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find variable type of name %s used in var ref", node.Value)
		}
		return entry.NewLoad(vType, vPtr), e.varGlTypes[node.Value]
	case *parser.CallExpression:
		var args []value.Value

		for _, a := range node.Params {
			val, _ := e.Emit(a, entry)
			args = append(args, val)
		}

		fncPtr, ok := e.functions[node.Function.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find function with name %s", node.Function.Value)
		}

		return entry.NewCall(fncPtr, args...), e.functionGlReturnTypes[node.Function.Value]
	case *parser.FunctionStatement:
		retType := varTypeToLlvm(node.Type)
		var paramTypes []*ir.Param

		for _, p := range node.Params {
			irParam := ir.NewParam(
				p.Name.Value,
				varTypeToLlvm(p.Type),
			)
			e.parameters[p.Name.Value] = irParam
			e.parametersGlTypes[p.Name.Value] = p.Type
			paramTypes = append(paramTypes, irParam)
		}

		fncPtr := e.m.NewFunc(node.Name.Value, retType, paramTypes...)
		e.functions[node.Name.Value] = fncPtr
		block := fncPtr.NewBlock("")
		e.functionBlocks[node.Name.Value] = block
		e.functionGlReturnTypes[node.Name.Value] = node.Type

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
		return fncPtr, node.Type
	case *parser.ReturnStatement:
		val, vt := e.Emit(node.Expr, entry)
		entry.NewRet(val)
		return val, vt
	case *parser.ReferenceExpression:
		vPtr, ok := e.variables[node.Var.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find variable with name %s in reference expr", node.Var.Value)
		}
		return vPtr, e.varGlTypes[node.Var.Value]
	case *parser.DereferenceExpression:
		ptr, vt := e.Emit(node.Var, entry)

		ptrTy, ok := ptr.Type().(*types.PointerType)
		if !ok {
			fmt.Printf("compile error: cannot deref non-ptr type %v\n", ptrTy)
		}

		return entry.NewLoad(ptrTy.ElemType, ptr), vt
	case *parser.CastExpression:
		src, lt := e.Emit(node.Expr, entry)
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
				return entry.NewZExt(src, dstType), node.Type
			}

			dstSize := getSizeForVarType(node.Type)
			srcSize := getSizeForLlvmType(src.Type())

			if srcSize < dstSize {
				return entry.NewSExt(src, dstType), node.Type
			} else if srcSize > dstSize {
				return entry.NewTrunc(src, dstType), node.Type
			} else {
				// same size, no cast necessary
				return src, lt
			}
		} else if leftIntOk && node.Type.Pointer > 0 {
			return entry.NewIntToPtr(src, dstType), node.Type
		} else if _, ok := src.Type().(*types.PointerType); ok && rightIntOk && node.Type.Pointer == 0 {
			if node.Type.Base != lexer.Int {
				// non 64 bit which is llvm default on most (i.e 64bit) systems
				fmt.Printf("compile warning: pointer to int cast may truncate")
			}
			return entry.NewPtrToInt(src, varTypeToLlvm(node.Type)), node.Type
		} else if leftIntOk && rightFloatOk {
			return entry.NewSIToFP(src, dstType), node.Type
		} else if leftFloatOk && rightIntOk {
			return entry.NewFPToSI(src, dstType), node.Type
		} else if leftPtrOk && rightPtrOk {
			/** c bitcast behaviour
			  int x = 1073741941; // 0100 0000 0000 0000 0000 0000 0111 0101 = 1073741941; as float = 2.somethingsomething
			  int* xx = &x;
			  float* fx = (float*)xx;
			  printf("%.100f", *fx); == 2.somethingsomething
			  return 0;
			*/
			return entry.NewBitCast(src, dstType), node.Type
		}
	case *parser.SizeofExpression:
		return constant.NewInt(types.I64, getSizeForVarType(node.Type)), lexer.VarType{Base: lexer.Uint, Pointer: 0}
	case *parser.ArrayLiteral:
		newFnc, ok := e.functions["arr_new"]
		if !ok {
			fmt.Printf("compiler error: cannot find arr_new while emitting array literal\n")
		}
		push, ok := e.functions["arr_push"]
		if !ok {
			fmt.Printf("compiler error: cannot find arr_push while emitting array literal\n")
		}

		sizeInt := constant.NewInt(types.I64, getSizeForVarType(node.Type))
		newCall := entry.NewCall(newFnc, sizeInt)
		node.Type.Pointer++
		newCallCasted := entry.NewBitCast(newCall, varTypeToLlvm(node.Type))
		ptr := entry.NewAlloca(varTypeToLlvm(node.Type))
		entry.NewStore(newCallCasted, ptr)
		for _, elem := range node.Items {
			v, _ := e.Emit(elem, entry)
			entry.NewCall(push, ptr, v)
		}

		// possibly dangerous?
		return newCallCasted, node.Type
	}

	return nil, lexer.VarType{}
}

// emitAdress literally only necessary because i need the vptr from ident expr, deref expr is same as e.emit lol
func (e *Emitter) emitAddress(node parser.Node, entry *ir.Block) (value.Value, lexer.VarType) {
	switch node := node.(type) {
	case *parser.IdentifierExpression:
		if param, ok := e.parameters[node.Value]; ok {
			return param, e.parametersGlTypes[node.Value]
		}
		vPtr, ok := e.variables[node.Value]
		if !ok {
			fmt.Printf("compile error: couldn't find variable with name %s in deref assignment", node.Value)
		}
		vt := e.varGlTypes[node.Value]
		vt.Pointer++
		return vPtr, vt
	case *parser.DereferenceExpression:
		ptr, t := e.Emit(node.Var, entry)
		return ptr, t
	default:
		fmt.Printf("compile error: invalid node type for emitAddress\n")
		return nil, lexer.VarType{}
	}
}

func varTypeToLlvm(vt lexer.VarType) types.Type {
	var baseType types.Type
	switch vt.Base {
	case lexer.None:
		baseType = nil
	case lexer.Int, lexer.Uint:
		baseType = types.I64
	case lexer.Int32, lexer.Uint32:
		baseType = types.I32
	case lexer.Int8, lexer.Uint8:
		baseType = types.I8
	case lexer.Int16, lexer.Uint16:
		baseType = types.I16
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
	case lexer.Bool, lexer.Int8, lexer.Uint8:
		return 1
	case lexer.Int16, lexer.Uint16:
		return 2
	case lexer.Int32, lexer.Float, lexer.Uint32:
		return 4
	case lexer.Int, lexer.Uint:
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
