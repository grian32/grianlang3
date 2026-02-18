package parser

import (
	"grianlang3/lexer"
	"testing"
	"time"
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
		"std module": {
			"import \"arrays\"",
			"import \"arrays\";",
		},
		"gl3 file": {
			"import \"stuff.gl3\"",
			"import \"stuff.gl3\";",
		},
	}

	runTests(t, tests)
}

func TestIfStatement(t *testing.T) {
	tests := map[string]InputOutput{
		"basic true": {
			"if true { \n  \n }",
			"if true {  };",
		},
		"basic false": {
			"if false { \n  \n }",
			"if false {  };",
		},
		"expr cond": {
			"if x > 5i32 { \n \n }",
			"if (x > 5(Int32)) {  };",
		},
		"expr cond with body": {
			"if x == 7u32 { \n stuff()\n }",
			"if (x == 7(Uint32)) { stuff(); };",
		},
		"with else empty": {
			"if true { \n \n } else { \n \n }",
			"if true {  } else {  };",
		},
		"with else body": {
			"if x < 1 { \n return 0 \n } else { \n return 1 \n }",
			"if (x < 1(Int)) { return 0(Int) } else { return 1(Int) };",
		},
		"nested if": {
			"if true { \n if false { \n \n } \n }",
			"if true { if false {  } };",
		},
		"logical condition": {
			"if x > 5 && y < 2 { \n \n }",
			"if ((x > 5(Int)) && (y < 2(Int))) {  };",
		},
	}

	runTests(t, tests)
}

func TestWhileStatement(t *testing.T) {
	tests := map[string]InputOutput{
		"basic true": {
			"while true { \n  \n }",
			"while true {  };",
		},
		"basic false": {
			"while false { \n  \n }",
			"while false {  };",
		},
		"expr cond": {
			"while x > 5i32 { \n \n }",
			"while (x > 5(Int32)) {  };",
		},
		"expr cond with body": {
			"while x == 7u32 { \n stuff()\n }",
			"while x == 7(Uint32) { stuff(); };",
		},
		"nested while": {
			"while true { \n while false { \n \n } \n }",
			"while true { while false {  } };",
		},
		"logical condition": {
			"while x > 5 && y < 2 { \n \n }",
			"while ((x > 5(Int)) && (y < 2(Int))) {  };",
		},
	}

	runTests(t, tests)
}

func TestInfixExpression(t *testing.T) {
	tests := map[string]InputOutput{
		"plus": {
			"5i32 + 3i32",
			"(5(Int32) + 3(Int32));",
		},
		"minus": {
			"8u32 - 2u32",
			"(8(Uint32) - 2(Uint32));",
		},
		"asterisk": {
			"x * 4i16",
			"(x * 4(Int16));",
		},
		"slash": {
			"12i32 / 3i32",
			"(12(Int32) / 3(Int32));",
		},
		"logical and": {
			"ok && ready",
			"(ok && ready);",
		},
		"logical or": {
			"ok || ready",
			"(ok || ready);",
		},
		"equals": {
			"x == 7i8",
			"(x == 7(Int8));",
		},
		"not equals": {
			"x != 7i8",
			"(x != 7(Int8));",
		},
		"less than": {
			"x < 5i32",
			"(x < 5(Int32));",
		},
		"greater than": {
			"x > 5i32",
			"(x > 5(Int32));",
		},
		"less than or equal": {
			"x <= 5i32",
			"(x <= 5(Int32));",
		},
		"greater than or equal": {
			"x >= 5i32",
			"(x >= 5(Int32));",
		},
		"dot": {
			"player.health",
			"(player . health);",
		},
		"array index literal": {
			"a[1]",
			"*(a + 1(Int));",
		},
		"array index identifier": {
			"items[i]",
			"*(items + i);",
		},
		"mismatched types": {
			"1 == true",
			"(1(Int) == true);",
		},
	}

	runTests(t, tests)
}

func TestPrecedenceExpression(t *testing.T) {
	tests := map[string]InputOutput{
		"product before sum": {
			"1 + 2 * 3",
			"(1(Int) + (2(Int) * 3(Int)));",
		},
		"grouped sum": {
			"(1 + 2) * 3",
			"((1(Int) + 2(Int)) * 3(Int));",
		},
		"logical and equality": {
			"1 + 2 == 3 && 4 > 5",
			"(((1(Int) + 2(Int)) == 3(Int)) && (4(Int) > 5(Int)));",
		},
		"prefix then sum": {
			"-1 + 2",
			"((-1(Int)) + 2(Int));",
		},
		"comparison before equality": {
			"1 + 2 < 3 == 4",
			"(((1(Int) + 2(Int)) < 3(Int)) == 4(Int));",
		},
		// TODO: add tests for precendences around sizeof, cast, call, etc once those are more clearly defined rather than just slapped onto something
	}

	runTests(t, tests)
}

func TestAssignmentExpression(t *testing.T) {
	tests := map[string]InputOutput{
		"identifier assign": {
			"x = 1i32",
			"x = 1(Int32);",
		},
		"deref assign": {
			"*ptr = 3i8",
			"*ptr = 3(Int8);",
		},
		"dot assign": {
			"player.health = 0",
			"(player . health) = 0(Int);",
		},
		"array index assign": {
			"items[i] = 2u16",
			"*(items + i) = 2(Uint16);",
		},
	}

	runTests(t, tests)
}

func TestCastExpression(t *testing.T) {
	tests := map[string]InputOutput{
		"int to float": {
			"1i32 as float",
			"1(Int32) as Float;",
		},
		"ident to pointer": {
			"x as int8*",
			"x as Int8*;",
		},
	}

	runTests(t, tests)
}

func TestPrefixExpression(t *testing.T) {
	tests := map[string]InputOutput{
		"not true": {
			"!true",
			"(!true);",
		},
		"not identifier": {
			"!ready",
			"(!ready);",
		},
		"not infix": {
			"!(x == 1i8)",
			"(!(x == 1(Int8)));",
		},
		"neg int": {
			"-5i32",
			"(-5(Int32));",
		},
		"neg identifier": {
			"-count",
			"(-count);",
		},
		"neg infix": {
			"-(x + 2i32)",
			"(-(x + 2(Int32)));",
		},
	}

	runTests(t, tests)
}

func TestCallExpression(t *testing.T) {
	tests := map[string]InputOutput{
		"empty args": {
			"foo()",
			"foo();",
		},
		"single arg": {
			"foo(1i32)",
			"foo(1(Int32));",
		},
		"multi args": {
			"foo(1i8, x, true)",
			"foo(1(Int8), x, true);",
		},
		"infix arg": {
			"sum(1i32 + 2i32)",
			"sum((1(Int32) + 2(Int32)));",
		},
		"nested call": {
			"outer(inner(2u16))",
			"outer(inner(2(Uint16)));",
		},
		"string and char args": {
			"print(\"hi\", 'a')",
			"print(\"hi\000\", 97(Int8));",
		},
	}

	runTests(t, tests)
}

func TestSizeofExpression(t *testing.T) {
	tests := map[string]InputOutput{
		"sizeof builtin type": {
			"sizeof int32",
			"sizeof Int32;",
		},
		"sizeof pointer": {
			"sizeof char*",
			"sizeof Char*;",
		},
	}

	runTests(t, tests)
}

func TestStructInitializationExpression(t *testing.T) {
	tests := map[string]InputOutput{
		"basic init": {
			"Player:{1i32, true}",
			"Player:{1(Int32),true};",
		},
		"single elem init": {
			"Cooked:{1i32}",
			"Cooked:{1(Int32)};",
		},
		"empty init": {
			"Empty:{}",
			"Empty:{};",
		},
	}

	runTests(t, tests)
}

func TestArrayLiteralExpression(t *testing.T) {
	tests := map[string]InputOutput{
		"int array": {
			"[int32; 1i32, 2i32]",
			"[Int32;1(Int32),2(Int32)];",
		},
		"empty array": {
			"[uint8;]",
			"[Uint8;];",
		},
	}

	runTests(t, tests)
}

func TestReferenceAndDereferenceExpression(t *testing.T) {
	tests := map[string]InputOutput{
		"reference identifier": {
			"&x",
			"&x;",
		},
		"dereference identifier": {
			"*ptr",
			"*ptr;",
		},
		"multi dereference identifier": {
			"***ptr",
			"***ptr;",
		},
	}

	runTests(t, tests)
}

func TestStructStatement(t *testing.T) {
	tests := map[string]struct {
		input  string
		name   string
		fields map[string]lexer.VarType
	}{
		"single field struct": {
			input: "struct Player { int32 health }",
			name:  "Player",
			fields: map[string]lexer.VarType{
				"health": {Base: lexer.Int32},
			},
		},
		"multi field struct": {
			input: "struct Player { int32 health char* name bool alive }",
			name:  "Player",
			fields: map[string]lexer.VarType{
				"health": {Base: lexer.Int32},
				"name":   {Base: lexer.Char, Pointer: 1},
				"alive":  {Base: lexer.Bool},
			},
		},
		"empty struct": {
			input:  "struct Empty { }",
			name:   "Empty",
			fields: map[string]lexer.VarType{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panicked: %v", r)
				}
			}()
			l := lexer.New(test.input)
			p := New(l)
			program := p.ParseProgram()
			if len(p.Errors) != 0 {
				t.Fatalf("got parser errors: %v", p.Errors)
			}
			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}
			stmt, ok := program.Statements[0].(*StructStatement)
			if !ok {
				t.Fatalf("expected StructStatement, got %T", program.Statements[0])
			}
			if stmt.Name != test.name {
				t.Fatalf("expected struct name %q, got %q", test.name, stmt.Name)
			}
			if len(stmt.Types) != len(test.fields) {
				t.Fatalf("expected %d fields, got %d", len(test.fields), len(stmt.Types))
			}
			if len(stmt.Names) != len(test.fields) {
				t.Fatalf("expected %d field names, got %d", len(test.fields), len(stmt.Names))
			}
			for fieldName, expectedType := range test.fields {
				idx, ok := stmt.Names[fieldName]
				if !ok {
					t.Fatalf("missing field %q", fieldName)
				}
				if idx < 0 || idx >= len(stmt.Types) {
					t.Fatalf("field %q has invalid index %d", fieldName, idx)
				}
				if stmt.Types[idx] != expectedType {
					t.Fatalf("field %q expected type %s, got %s", fieldName, expectedType.String(), stmt.Types[idx].String())
				}
			}
		})
	}
}

func runTests(t *testing.T, tests map[string]InputOutput) {
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			timeout := time.After(1 * time.Second)
			done := make(chan bool)
			var p *Parser
			go func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("panicked: %v", r)
						done <- true
					}
				}()
				l := lexer.New(test.input)
				p = New(l)
				program := p.ParseProgram()
				if len(p.Errors) != 0 {
					t.Errorf("got parser errors: %v", p.Errors)
				}
				if program.String() != test.output {
					t.Errorf("wanted: %s, got: %s", test.output, program.String())
				}
				done <- true
			}()

			select {
			case <-timeout:
				var trimmedErrors []string
				if len(p.Errors) > 5 {
					trimmedErrors = p.Errors[:5]
				} else {
					trimmedErrors = p.Errors
				}
				t.Fatalf("timed out after 1 second with parser errors: %v", trimmedErrors)
			case <-done:
			}
		})
	}
}
