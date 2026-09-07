package checkedast

type BaseType uint8

const (
	Invalid BaseType = iota
	Int
	Int32
	Int16
	Int8
	Char
	Uint
	Uint32
	Uint16
	Uint8
	Bool
	Void
	Float
	StructType
)

type Type struct {
	Base    BaseType
	Pointer uint8
	Struct  StructID
}

type TypedName struct {
	Name string
	Type Type
}

type Struct struct {
	Name       string
	Id         StructID
	Fields     []TypedName
	FieldNames map[string]int
}

type Function struct {
	Name           string
	Id             FunctionID
	Parameters     []TypedName
	ParameterNames map[string]int
	ReturnType     Type
	Private        bool
}

type Global struct {
	Name     string
	Id       GlobalID
	Constant bool
	Type     Type
}

type Program struct {
	Structs   []Struct
	Functions []Function
	Globals   []Global
}
