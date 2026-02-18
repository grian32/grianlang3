package parser

import (
	"grianlang3/lexer"
	"testing"
)

type InputOutput struct {
	input  string
	output string
}

func TestLiterals(t *testing.T) {
	tests := map[string]InputOutput{
		"int64": {
			"4",
			"4(Int);",
		},
		"int32": {
			"4i32",
			"4(Int32);",
		},
		"int16": {
			"4i16",
			"4(Int16);",
		},
		"int8": {
			"4i8",
			"4(Int8);",
		},
		"uint32": {
			"4u32",
			"4(Uint32);",
		},
		"uint16": {
			"4u16",
			"4(Uint16);",
		},
		"uint8": {
			"4u8",
			"4(Uint8);",
		},
		"uint64": {
			"4u64",
			"4(Uint);",
		},
		"float": {
			"1.5",
			"1.5(Float);",
		},
		"true bool": {
			"true",
			"true;",
		},
		"false bool": {
			"false",
			"false;",
		},
		"char": {
			"'a'",
			"97(Int8);",
		},
		"string": {
			"\"hello\"",
			"\"hello\000\";",
		},
	}

	runTests(t, tests)
}

func TestDefStatement(t *testing.T) {
	tests := map[string]InputOutput{
		"int def": {
			"def int x = 7",
			"def Int x = 7(Int);",
		},
		"int32 def": {
			"def int32 x = 7i32",
			"def Int32 x = 7(Int32);",
		},
		"uint32 def": {
			"def uint32 x = 7u32",
			"def Uint32 x = 7(Uint32);",
		},
		"float def": {
			"def float x = 1.5",
			"def Float x = 1.5(Float);",
		},
		"bool def": {
			"def bool x = true",
			"def Bool x = true;",
		},
		"char def": {
			"def char x = 'a'",
			"def Char x = 97(Int8);",
		},
		"string def": {
			"def char* x = \"hello\"",
			"def Char* x = \"hello\000\";",
		},
		"multi pointer": {
			"def char*** x = \"hello\"",
			"def Char*** x = \"hello\000\";",
		},
	}

	runTests(t, tests)
}

func TestFuncStatement(t *testing.T) {
	tests := map[string]InputOutput{
		"basic main func": {
			"fnc main() -> int32 { \n return 0i32 \n }",
			"fnc main() -> Int32 { return 0(Int32) };",
		},
		"mult params": {
			"fnc stuff(int8 x, int32** other) -> none { \n }",
			"fnc stuff(Int8 x, Int32** other) -> Void {  };",
		},
		"no params void": {
			"fnc foo() -> none { \n }",
			"fnc foo() -> Void {  };",
		},
		"non void ret with params": {
			"fnc x(int8 x, int32** other) -> int8 { \n return x; \n }",
			"fnc x(Int8 x, Int32** other) -> Int8 { return x };",
		},
		"bool ret": {
			"fnc isok() -> bool { \n return true; \n }",
			"fnc isok() -> Bool { return true };",
		},
		"char pointer ret": {
			"fnc greet() -> char* { \n return \"hi\"; \n }",
			"fnc greet() -> Char* { return \"hi\000\" };",
		},
	}

	runTests(t, tests)
}

func TestImportStatement(t *testing.T) {
	tests := map[string]InputOutput{
		"test std module": {
			"import \"arrays\"",
			"import \"arrays\";",
		},
		"test gl3 file": {
			"import \"stuff.gl3\"",
			"import \"stuff.gl3\";",
		},
	}

	runTests(t, tests)
}

func runTests(t *testing.T, tests map[string]InputOutput) {
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			l := lexer.New(test.input)
			p := New(l)
			program := p.ParseProgram()
			if len(p.Errors) != 0 {
				t.Errorf("got parser errors: %v", p.Errors)
			}
			if program.String() != test.output {
				t.Errorf("wanted: %s, got: %s", test.output, program.String())
			}
		})
	}
}
