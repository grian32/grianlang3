package emitter

import (
	"fmt"
	"grianlang3/lexer"

	_ "embed"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

type BuiltinDef struct {
	RetType   types.Type
	RetGlType lexer.VarType
	Params    []types.Type
}

func NewBuiltinDef(ret types.Type, glRet lexer.VarType, param ...types.Type) BuiltinDef {
	return BuiltinDef{RetType: ret, RetGlType: glRet, Params: param}
}

func newVt(base lexer.BaseVarType) lexer.VarType {
	return lexer.VarType{Base: base, Pointer: 0}
}

func newVtPtr(base lexer.BaseVarType, ptr uint8) lexer.VarType {
	return lexer.VarType{Base: base, Pointer: ptr}
}

var builtinModules = map[string]map[string]BuiltinDef{
	"dbg": {
		"dbg_i64":   NewBuiltinDef(types.Void, newVt(lexer.Void), types.I64),
		"dbg_i32":   NewBuiltinDef(types.Void, newVt(lexer.Void), types.I32),
		"dbg_i16":   NewBuiltinDef(types.Void, newVt(lexer.Void), types.I16),
		"dbg_i8":    NewBuiltinDef(types.Void, newVt(lexer.Void), types.I8),
		"dbg_u64":   NewBuiltinDef(types.Void, newVt(lexer.Void), types.I64),
		"dbg_u32":   NewBuiltinDef(types.Void, newVt(lexer.Void), types.I32),
		"dbg_u16":   NewBuiltinDef(types.Void, newVt(lexer.Void), types.I16),
		"dbg_u8":    NewBuiltinDef(types.Void, newVt(lexer.Void), types.I8),
		"dbg_float": NewBuiltinDef(types.Void, newVt(lexer.Void), types.Float),
		"dbg_bool":  NewBuiltinDef(types.Void, newVt(lexer.Void), types.I1),
	},
	//"malloc":    NewBuiltinDef(types.I8Ptr, newVtPtr(lexer.Int8, 1), types.I64),
	"arrays": {
		"arr_new":  NewBuiltinDef(types.I8Ptr, newVtPtr(lexer.None, 1), types.I64),
		"arr_push": NewBuiltinDef(types.Void, newVt(lexer.None), types.I8Ptr, types.I8Ptr),
		"arr_free": NewBuiltinDef(types.Void, newVt(lexer.None), types.I8Ptr),
	},
}

func AddBuiltinModule(e *Emitter, moduleName string) error {
	builtins, ok := builtinModules[moduleName]
	if !ok {
		return fmt.Errorf("couldn't find module with name %s", moduleName)
	}

	for name, typing := range builtins {
		var params []*ir.Param
		for _, p := range typing.Params {
			params = append(params, ir.NewParam("", p))
		}
		fnc := e.m.NewFunc(name, typing.RetType, params...)
		e.functions[name] = fnc
		e.functionGlReturnTypes[name] = typing.RetGlType
	}
	e.builtinModules = append(e.builtinModules, fmt.Sprintf("%s.ll", moduleName))

	return nil
}
