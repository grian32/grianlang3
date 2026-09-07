package checkedast

type Struct struct {
	Name string
	Id   StructID
	// todo: fields
}

type Function struct {
	Name string
	Id   FunctionID
	// todo: params
}

type Global struct {
	Name string
	Id   GlobalID
	// todo: type, mut,
}

type Program struct {
	Structs   []Struct
	Functions []Function
	Globals   []Global
}
