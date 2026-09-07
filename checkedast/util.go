package checkedast

import "gl3/lexer"

func ConvertVarType(vt lexer.VarType, symbols map[string]Symbol) (Type, bool) {
	if vt.IsStructType {
		id, ok := symbols[vt.StructName].(StructID)
		if !ok {
			return Type{}, false
		}
		return Type{Base: StructType, Pointer: vt.Pointer, Struct: id}, true
	}

	var base BaseType
	switch vt.Base {
	case lexer.Int:
		base = Int
	case lexer.Int32:
		base = Int32
	case lexer.Int16:
		base = Int16
	case lexer.Int8:
		base = Int8
	case lexer.Char:
		base = Char
	case lexer.Uint:
		base = Uint
	case lexer.Uint32:
		base = Uint32
	case lexer.Uint16:
		base = Uint16
	case lexer.Uint8:
		base = Uint8
	case lexer.Bool:
		base = Bool
	case lexer.Void:
		base = Void
	case lexer.Float:
		base = Float
	default:
		return Type{}, false
	}
	return Type{Base: base, Pointer: vt.Pointer}, true
}
