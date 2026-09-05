package parser

import (
	"fmt"
	"gl3/lexer"
	"strings"
	"testing"
)

// These tests characterize the current grammar, including the parentheses needed
// to cast only one operand. They are not a proposal for a new precedence order.
func TestCastPrecedence(t *testing.T) {
	tests := []InputOutput{
		{"a + b as float", "cast(Float, +(a, b))"},
		{"a + (b as float)", "+(a, cast(Float, b))"},
		{"a as float + b", "+(cast(Float, a), b)"},
		{"a as float + b as float", "cast(Float, +(cast(Float, a), b))"},
		{"a as float + (b as float)", "+(cast(Float, a), cast(Float, b))"},
		{"(a + b) as float", "cast(Float, +(a, b))"},
		{"a * b as float", "cast(Float, *(a, b))"},
		{"a * (b as float)", "*(a, cast(Float, b))"},
		{"(a as float) * b", "*(cast(Float, a), b)"},
		{"a as float / b", "/(cast(Float, a), b)"},
		{"a / b as float", "cast(Float, /(a, b))"},
		{"a < b as int", "<(a, cast(Int, b))"},
		{"a as int < b", "<(cast(Int, a), b)"},
		{"a == b as int", "==(a, cast(Int, b))"},
		{"a as int == b", "==(cast(Int, a), b)"},
		{"a && b as bool", "&&(a, cast(Bool, b))"},
		{"a as bool || b", "||(cast(Bool, a), b)"},
		{"-a as float", "cast(Float, prefix(-, a))"},
		{"-(a as float)", "prefix(-, cast(Float, a))"},
		{"!a as bool", "cast(Bool, prefix(!, a))"},
		{"!(a as bool)", "prefix(!, cast(Bool, a))"},
		{"*p as int", "cast(Int, deref(p))"},
		{"*p as int*", "cast(Int*, deref(p))"},
		{"*(p as int*)", "deref(cast(Int*, p))"},
		{"(&a) as int", "cast(Int, ref(a))"},
		{"a as int32 as float", "cast(Float, cast(Int32, a))"},
		{"p as Node**", "cast(Node**, p)"},
		{"f(a) as int", "cast(Int, call(f, a))"},
		{"f(a as int, b)", "call(f, cast(Int, a), b)"},
		{"items[i] as int", "cast(Int, deref(+(items, i)))"},
		{"(p as int*)[i]", "deref(+(cast(Int*, p), i))"},
		{"p as int*[i]", "deref(+(cast(Int*, p), i))"},
		{"p.field as int", "cast(Int, .(p, field))"},
		{"(p as Node*).field", ".(cast(Node*, p), field)"},
		{"p as Node*.field", ".(cast(Node*, p), field)"},
		{"f(a)[i] as int", "cast(Int, deref(+(call(f, a), i)))"},
		{"p = malloc(sizeof Node) as Node*", "assign(p, cast(Node*, call(malloc, sizeof(Node))))"},
		{"a = b as int + c", "assign(a, +(cast(Int, b), c))"},
	}
	runExpressionShapeTests(t, tests)
}

func TestSizeofPrecedence(t *testing.T) {
	tests := []InputOutput{
		{"sizeof Node", "sizeof(Node)"},
		{"sizeof Node**", "sizeof(Node**)"},
		{"sizeof int32**", "sizeof(Int32**)"},
		// Whitespace does not turn pointer stars into multiplication operators.
		{"sizeof int32 *", "sizeof(Int32*)"},
		{"sizeof int32 * *", "sizeof(Int32**)"},
		{"(sizeof int32) * n", "*(sizeof(Int32), n)"},
		{"(sizeof int32*) * n", "*(sizeof(Int32*), n)"},
		{"n * sizeof int32", "*(n, sizeof(Int32))"},
		{"sizeof int32 / n", "/(sizeof(Int32), n)"},
		{"sizeof int32 + n", "+(sizeof(Int32), n)"},
		{"n + sizeof int32", "+(n, sizeof(Int32))"},
		{"sizeof int32 < n", "<(sizeof(Int32), n)"},
		{"-sizeof int32", "prefix(-, sizeof(Int32))"},
		{"sizeof int32 as uint", "cast(Uint, sizeof(Int32))"},
		{"sizeof int32 + n as uint", "cast(Uint, +(sizeof(Int32), n))"},
		{"malloc((sizeof Node) * n) as Node*", "cast(Node*, call(malloc, *(sizeof(Node), n)))"},
		{"f(sizeof Node*, n as int, sizeof char)", "call(f, sizeof(Node*), cast(Int, n), sizeof(Char))"},
	}
	runExpressionShapeTests(t, tests)
}

func TestTypeStarsConsumeFollowingAsterisks(t *testing.T) {
	// A star following a type is a pointer suffix, even with spaces. These
	// inputs currently leave n as a separate statement; they do not multiply.
	// This characterizes token consumption, not a recommended way to write code.
	tests := []InputOutput{
		{"sizeof int32 * n", "sizeof(Int32*)"},
		{"sizeof int32* * n", "sizeof(Int32**)"},
		{"p as int32 * n", "cast(Int32*, p)"},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			p := New(lexer.New(test.input))
			program := p.ParseProgram()
			if len(p.Errors) != 0 {
				t.Fatalf("parser errors: %v", p.Errors)
			}
			if len(program.Statements) != 2 {
				t.Fatalf("expected type expression and trailing identifier, got %d statements", len(program.Statements))
			}
			if got := statementExpressionShape(t, program.Statements[0]); got != test.output {
				t.Fatalf("want: %s; got: %s", test.output, got)
			}
			if got := statementExpressionShape(t, program.Statements[1]); got != "n" {
				t.Fatalf("expected trailing n, got %s", got)
			}
		})
	}
}

func TestExpressionAssociativityAndPostfix(t *testing.T) {
	tests := []InputOutput{
		{"a - b - c", "-(-(a, b), c)"},
		{"a / b * c", "*(/(a, b), c)"},
		{"a = b = c", "assign(a, assign(b, c))"},
		{"a || b && c", "||(a, &&(b, c))"},
		{"a && b || c", "||(&&(a, b), c)"},
		{"-f(a)[i]", "prefix(-, deref(+(call(f, a), i)))"},
		{"items[i][j]", "deref(+(deref(+(items, i)), j))"},
		{"items[i].field", ".(deref(+(items, i)), field)"},
	}
	runExpressionShapeTests(t, tests)
}

func runExpressionShapeTests(t *testing.T, tests []InputOutput) {
	t.Helper()
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			// Exercise EOF and block/semicolon boundaries, including the next
			// statement so a parser cannot silently swallow trailing tokens.
			for _, inFunction := range []bool{false, true} {
				name, input := "standalone", test.input
				if inFunction {
					name = "function_body"
					input = "fnc probe() -> none { " + input + "; sentinel() }"
				}
				t.Run(name, func(t *testing.T) {
					p := New(lexer.New(input))
					program := p.ParseProgram()
					if len(p.Errors) != 0 {
						t.Fatalf("parser errors: %v", p.Errors)
					}
					if len(program.Statements) != 1 {
						t.Fatalf("expected one top-level statement, got %d", len(program.Statements))
					}
					stmt := program.Statements[0]
					if inFunction {
						fn, ok := stmt.(*FunctionStatement)
						if !ok {
							t.Fatalf("expected function, got %T", stmt)
						}
						if len(fn.Body.Statements) != 2 {
							t.Fatalf("expected expression and sentinel, got %d statements", len(fn.Body.Statements))
						}
						if got := statementExpressionShape(t, fn.Body.Statements[1]); got != "call(sentinel)" {
							t.Fatalf("following statement changed: %s", got)
						}
						stmt = fn.Body.Statements[0]
					}
					if got := statementExpressionShape(t, stmt); got != test.output {
						t.Fatalf("input: %s\nwant: %s\n got: %s", input, test.output, got)
					}
				})
			}
		})
	}
}

func statementExpressionShape(t *testing.T, stmt Statement) string {
	t.Helper()
	expr, ok := stmt.(*ExpressionStatement)
	if !ok {
		t.Fatalf("expected expression statement, got %T", stmt)
	}
	return expressionShape(t, expr.Expression)
}

// Unlike Expression.String(), this representation distinguishes a cast of a
// dereference from a dereference of a cast, and preserves assignment nesting.
func expressionShape(t *testing.T, expr Expression) string {
	t.Helper()
	shape := func(e Expression) string { return expressionShape(t, e) }
	switch e := expr.(type) {
	case *IdentifierExpression:
		return e.Value
	case *InfixExpression:
		return fmt.Sprintf("%s(%s, %s)", e.Operator, shape(e.Left), shape(e.Right))
	case *CastExpression:
		return fmt.Sprintf("cast(%s, %s)", e.Type.String(), shape(e.Expr))
	case *SizeofExpression:
		return "sizeof(" + e.Type.String() + ")"
	case *PrefixExpression:
		return fmt.Sprintf("prefix(%s, %s)", e.Operator, shape(e.Right))
	case *DereferenceExpression:
		return "deref(" + shape(e.Var) + ")"
	case *ReferenceExpression:
		if e.Var == nil {
			t.Fatal("reference has no operand")
		}
		return "ref(" + shape(e.Var) + ")"
	case *AssignmentExpression:
		return fmt.Sprintf("assign(%s, %s)", shape(e.Left), shape(e.Right))
	case *CallExpression:
		parts := []string{shape(e.Function)}
		for _, param := range e.Params {
			parts = append(parts, shape(param))
		}
		return "call(" + strings.Join(parts, ", ") + ")"
	default:
		t.Fatalf("unexpected expression node %T", expr)
		return ""
	}
}
