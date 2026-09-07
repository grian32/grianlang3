package checker

import (
	"gl3/checkedast"
	"gl3/lexer"
	"gl3/parser"
	"reflect"
	"strings"
	"testing"
)

type expectedDiagnostic struct {
	messageContains []string
	line            uint32
}

var declarationTests = []struct {
	name        string
	source      string
	want        checkedast.Program
	symbols     map[string]checkedast.Symbol
	diagnostics []expectedDiagnostic
}{
	{name: "empty"},
	{
		name: "interleaved declarations",
		source: `
global int32 first = 1i32
fnc alpha() -> int32 { return 1i32 }
struct First { int32 value }
global const int32 second = 2i32
struct Second { First* previous }
fnc beta(int32 input) -> int32 {
    def int32 local = input
    return local
}
`,
		want: checkedast.Program{
			Structs: []checkedast.Struct{
				{Name: "First", Id: 0},
				{Name: "Second", Id: 1},
			},
			Functions: []checkedast.Function{
				{Name: "alpha", Id: 0},
				{Name: "beta", Id: 1},
			},
			Globals: []checkedast.Global{
				{Name: "first", Id: 0},
				{Name: "second", Id: 1},
			},
		},
		symbols: map[string]checkedast.Symbol{
			"First": checkedast.StructID(0), "Second": checkedast.StructID(1),
			"alpha": checkedast.FunctionID(0), "beta": checkedast.FunctionID(1),
			"first": checkedast.GlobalID(0), "second": checkedast.GlobalID(1),
		},
	},
	{
		name: "other statements are not declarations",
		source: `
import "io"
def int32 topLevelLocal = 0i32
fnc probe(int32 parameter) -> int32 {
    def int32 inside = parameter
    return inside
}
`,
		want: checkedast.Program{
			Functions: []checkedast.Function{{Name: "probe", Id: 0}},
		},
		symbols: map[string]checkedast.Symbol{
			"probe": checkedast.FunctionID(0),
		},
	},
	{
		name: "duplicate struct then struct",
		source: `struct repeated { int32 value }
struct repeated { int32 value }`,
		diagnostics: []expectedDiagnostic{
			{messageContains: []string{"duplicate", "repeated"}, line: 2},
		},
	},
	{
		name: "duplicate struct then function",
		source: `struct repeated { int32 value }
fnc repeated() -> int32 { return 0i32 }`,
		diagnostics: []expectedDiagnostic{
			{messageContains: []string{"duplicate", "repeated"}, line: 2},
		},
	},
	{
		name: "duplicate struct then global",
		source: `struct repeated { int32 value }
global int32 repeated = 0i32`,
		diagnostics: []expectedDiagnostic{
			{messageContains: []string{"duplicate", "repeated"}, line: 2},
		},
	},
	{
		name: "duplicate function then struct",
		source: `fnc repeated() -> int32 { return 0i32 }
struct repeated { int32 value }`,
		diagnostics: []expectedDiagnostic{
			{messageContains: []string{"duplicate", "repeated"}, line: 2},
		},
	},
	{
		name: "duplicate function then function",
		source: `fnc repeated() -> int32 { return 0i32 }
fnc repeated() -> int32 { return 0i32 }`,
		diagnostics: []expectedDiagnostic{
			{messageContains: []string{"duplicate", "repeated"}, line: 2},
		},
	},
	{
		name: "duplicate function then global",
		source: `fnc repeated() -> int32 { return 0i32 }
global int32 repeated = 0i32`,
		diagnostics: []expectedDiagnostic{
			{messageContains: []string{"duplicate", "repeated"}, line: 2},
		},
	},
	{
		name: "duplicate global then struct",
		source: `global int32 repeated = 0i32
struct repeated { int32 value }`,
		diagnostics: []expectedDiagnostic{
			{messageContains: []string{"duplicate", "repeated"}, line: 2},
		},
	},
	{
		name: "duplicate global then function",
		source: `global int32 repeated = 0i32
fnc repeated() -> int32 { return 0i32 }`,
		diagnostics: []expectedDiagnostic{
			{messageContains: []string{"duplicate", "repeated"}, line: 2},
		},
	},
	{
		name: "duplicate global then global",
		source: `global int32 repeated = 0i32
global int32 repeated = 0i32`,
		diagnostics: []expectedDiagnostic{
			{messageContains: []string{"duplicate", "repeated"}, line: 2},
		},
	},
	{
		name: "multiple duplicate names",
		source: `struct repeated { int32 value }
struct repeated { int32 other }
global int32 counter = 0i32
global int32 counter = 1i32`,
		diagnostics: []expectedDiagnostic{
			{messageContains: []string{"duplicate", "repeated"}, line: 2},
			{messageContains: []string{"duplicate", "counter"}, line: 4},
		},
	},
}

// CheckProgram is the entry point; its returned program is the declaration output.
func TestCheckProgram(t *testing.T) {
	for _, test := range declarationTests {
		t.Run(test.name, func(t *testing.T) {
			program := parseDeclarations(t, test.source)
			c := &Checker{}
			got, diagnostics := c.CheckProgram(program)

			assertDiagnostics(t, diagnostics, test.diagnostics)
			if len(test.diagnostics) != 0 {
				// Recovery output is not part of the error-case contract.
				return
			}
			assertDeclarations(t, got, &test.want)
			assertSymbols(t, c.symbols, test.symbols)
		})
	}
}

func parseDeclarations(t *testing.T, source string) *parser.Program {
	t.Helper()
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if len(p.Errors) != 0 {
		t.Fatalf("invalid test fixture: %v", p.Errors)
	}
	return program
}

func assertDeclarations(t *testing.T, got, want *checkedast.Program) {
	t.Helper()
	if got == nil {
		t.Fatal("CheckProgram returned nil")
	}
	assertSlice(t, "structs", got.Structs, want.Structs)
	assertSlice(t, "functions", got.Functions, want.Functions)
	assertSlice(t, "globals", got.Globals, want.Globals)
}

func assertSlice[T any](t *testing.T, name string, got, want []T) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: want %d entries, got %d", name, len(want), len(got))
	}
	for i := range want {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("%s[%d]: want %+v, got %+v", name, i, want[i], got[i])
		}
	}
}

func assertSymbols(t *testing.T, got, want map[string]checkedast.Symbol) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("symbols: want %v, got %v", want, got)
	}
	for name, expected := range want {
		actual, ok := got[name]
		if !ok || actual != expected {
			t.Errorf("symbol %q: want %T(%v), got %T(%v), present=%v",
				name, expected, expected, actual, actual, ok)
		}
	}
}

// Match diagnostics without requiring a particular reporting order. Fragments
// keep these tests independent of punctuation and surrounding explanation.
func assertDiagnostics(t *testing.T, got []Diagnostic, want []expectedDiagnostic) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("diagnostics: want %d, got %d: %+v", len(want), len(got), got)
	}
	matched := make([]bool, len(got))
	for _, expected := range want {
		found := false
		for i, actual := range got {
			if matched[i] || actual.Position == nil || actual.Position.StartLine != expected.line {
				continue
			}
			matches := true
			for _, fragment := range expected.messageContains {
				if !strings.Contains(actual.Message, fragment) {
					matches = false
					break
				}
			}
			if matches {
				matched[i] = true
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing diagnostic containing %q at line %d; got %+v", expected.messageContains, expected.line, got)
		}
	}
}
