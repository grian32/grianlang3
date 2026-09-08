package checker

import (
	"gl3/checkedast"
	"testing"
)

var scopeTests = []struct {
	name        string
	source      string
	want        checkedast.Program
	symbols     map[string]checkedast.Symbol
	diagnostics []expectedDiagnostic
}{
	{
		name: "parameters and locals resolve to local IDs",
		source: `fnc copy(int32 input) -> int32 {
    def int32 output = input
    return output
}`,
		want: checkedast.Program{Functions: []checkedast.Function{{
			Name: "copy", Id: 0,
			Parameters:     []checkedast.TypedName{{Name: "input", Type: checkedast.Type{Base: checkedast.Int32}}},
			ParameterNames: map[string]int{"input": 0},
			ReturnType:     checkedast.Type{Base: checkedast.Int32},
			Locals: []checkedast.Local{
				{Name: "input", Type: checkedast.Type{Base: checkedast.Int32}},
				{Name: "output", Type: checkedast.Type{Base: checkedast.Int32}},
			},
			Body: checkedast.Block{Statements: []checkedast.Stmt{
				&checkedast.LocalDeclaration{ID: 1, Initializer: &checkedast.LocalRef{ExprInfo: checkedast.ExprInfo{ResultType: checkedast.Type{Base: checkedast.Int32}}, ID: 0}},
				&checkedast.Return{Value: &checkedast.LocalRef{ExprInfo: checkedast.ExprInfo{ResultType: checkedast.Type{Base: checkedast.Int32}}, ID: 1}},
			}},
		}}},
		symbols: map[string]checkedast.Symbol{"copy": checkedast.FunctionID(0)},
	},
	{
		name: "local initializer resolves outer name before shadowing",
		source: `global int32 value = 7i32
fnc copy() -> int32 {
    def int32 value = value
    return value
}
fnc read() -> int32 { return value }`,
		want: checkedast.Program{
			Globals: []checkedast.Global{{Name: "value", Id: 0, Type: checkedast.Type{Base: checkedast.Int32},
				Initializer: &checkedast.IntegerLiteral{ExprInfo: checkedast.ExprInfo{ResultType: checkedast.Type{Base: checkedast.Int32}}, Value: 7}}},
			Functions: []checkedast.Function{
				{Name: "copy", Id: 0, ReturnType: checkedast.Type{Base: checkedast.Int32},
					Locals: []checkedast.Local{{Name: "value", Type: checkedast.Type{Base: checkedast.Int32}}},
					Body: checkedast.Block{Statements: []checkedast.Stmt{
						&checkedast.LocalDeclaration{ID: 0, Initializer: &checkedast.GlobalRef{ExprInfo: checkedast.ExprInfo{ResultType: checkedast.Type{Base: checkedast.Int32}}, ID: 0}},
						&checkedast.Return{Value: &checkedast.LocalRef{ExprInfo: checkedast.ExprInfo{ResultType: checkedast.Type{Base: checkedast.Int32}}, ID: 0}},
					}}},
				{Name: "read", Id: 1, ReturnType: checkedast.Type{Base: checkedast.Int32},
					Body: checkedast.Block{Statements: []checkedast.Stmt{
						&checkedast.Return{Value: &checkedast.GlobalRef{ExprInfo: checkedast.ExprInfo{ResultType: checkedast.Type{Base: checkedast.Int32}}, ID: 0}},
					}}},
			},
		},
		symbols: map[string]checkedast.Symbol{"value": checkedast.GlobalID(0), "copy": checkedast.FunctionID(0), "read": checkedast.FunctionID(1)},
	},
	{
		name: "parameter names and IDs are function local",
		source: `fnc first(int32 value) -> int32 { return value }
fnc second(bool value) -> bool { return value }`,
		want: checkedast.Program{Functions: []checkedast.Function{
			{Name: "first", Id: 0,
				Parameters:     []checkedast.TypedName{{Name: "value", Type: checkedast.Type{Base: checkedast.Int32}}},
				ParameterNames: map[string]int{"value": 0}, ReturnType: checkedast.Type{Base: checkedast.Int32},
				Locals: []checkedast.Local{{Name: "value", Type: checkedast.Type{Base: checkedast.Int32}}},
				Body: checkedast.Block{Statements: []checkedast.Stmt{
					&checkedast.Return{Value: &checkedast.LocalRef{ExprInfo: checkedast.ExprInfo{ResultType: checkedast.Type{Base: checkedast.Int32}}, ID: 0}},
				}}},
			{Name: "second", Id: 1,
				Parameters:     []checkedast.TypedName{{Name: "value", Type: checkedast.Type{Base: checkedast.Bool}}},
				ParameterNames: map[string]int{"value": 0}, ReturnType: checkedast.Type{Base: checkedast.Bool},
				Locals: []checkedast.Local{{Name: "value", Type: checkedast.Type{Base: checkedast.Bool}}},
				Body: checkedast.Block{Statements: []checkedast.Stmt{
					&checkedast.Return{Value: &checkedast.LocalRef{ExprInfo: checkedast.ExprInfo{ResultType: checkedast.Type{Base: checkedast.Bool}}, ID: 0}},
				}}},
		}},
		symbols: map[string]checkedast.Symbol{"first": checkedast.FunctionID(0), "second": checkedast.FunctionID(1)},
	},
	{
		name: "local names cannot leak between functions",
		source: `fnc first() -> int32 {
    def int32 hidden = 1i32
    return hidden
}
fnc second() -> int32 {
    return hidden
}`,
		diagnostics: []expectedDiagnostic{{messageContains: []string{"unknown", "hidden"}, line: 6}},
	},
	{
		name: "parameter names cannot leak between functions",
		source: `fnc first(int32 hidden) -> int32 { return hidden }
fnc second() -> int32 { return hidden }`,
		diagnostics: []expectedDiagnostic{{messageContains: []string{"unknown", "hidden"}, line: 2}},
	},
	{
		name: "local is unavailable in its own initializer",
		source: `fnc sample() -> int32 {
    def int32 value = value
    return 0i32
}`,
		diagnostics: []expectedDiagnostic{{messageContains: []string{"unknown", "value"}, line: 2}},
	},
	{
		name: "local is unavailable before its declaration",
		source: `fnc sample() -> int32 {
    def int32 first = later
    def int32 later = 1i32
    return 0i32
}`,
		diagnostics: []expectedDiagnostic{{messageContains: []string{"unknown", "later"}, line: 2}},
	},
	{
		name: "duplicate locals in one scope",
		source: `fnc sample() -> int32 {
    def int32 value = 1i32
    def int32 value = 2i32
    return 0i32
}`,
		diagnostics: []expectedDiagnostic{{messageContains: []string{"duplicate", "value"}, line: 3}},
	},
}

func TestCheckProgramScopes(t *testing.T) {
	for _, test := range scopeTests {
		t.Run(test.name, func(t *testing.T) {
			program := parseDeclarations(t, test.source)
			c := NewChecker()
			got, diagnostics := c.CheckProgram(program)
			assertDiagnostics(t, diagnostics, test.diagnostics)
			if len(test.diagnostics) != 0 {
				return
			}
			assertDeclarations(t, got, &test.want)
			assertSymbols(t, c.symbols, test.symbols)
		})
	}
}
