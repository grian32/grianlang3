package checker

import (
	"fmt"
	"gl3/emitter"
	"gl3/lexer"
	"gl3/parser"
	"gl3/util"
	"regexp"
	"strings"
)

type Checker struct {
	importsFound     map[string]struct{}
	builtinNames     map[string]map[string]struct{}
	constVars        map[string]struct{}
	varTypes         map[string]lexer.VarType
	structFieldTypes map[string]map[string]lexer.VarType
	Errors           []util.PositionError
}

func New() *Checker {
	return &Checker{
		builtinNames:     emitter.GetBuiltinNames(),
		importsFound:     make(map[string]struct{}),
		constVars:        make(map[string]struct{}),
		varTypes:         make(map[string]lexer.VarType),
		structFieldTypes: make(map[string]map[string]lexer.VarType),
	}
}

func (c *Checker) Check(node parser.Node) {
	switch node := node.(type) {
	case *parser.Program:
		for _, s := range node.Statements {
			c.Check(s)
		}
	case *parser.FunctionStatement:
		for _, s := range node.Body.Statements {
			c.Check(s)
		}
	case *parser.WhileStatement:
		c.Check(node.Condition)
		for _, s := range node.Body.Statements {
			c.Check(s)
		}
	case *parser.IfStatement:
		c.Check(node.Condition)
		for _, s := range node.Success.Statements {
			c.Check(s)
		}
		if node.Fail != nil {
			for _, s := range node.Fail.Statements {
				c.Check(s)
			}
		}
	case *parser.ExpressionStatement:
		c.Check(node.Expression)
	case *parser.ReturnStatement:
		c.Check(node.Expr)
	case *parser.CastExpression:
		c.Check(node.Expr)
	case *parser.DereferenceExpression:
		c.Check(node.Var)
	case *parser.ImportStatement:
		if !strings.HasSuffix(node.Path, ".gl3") {
			c.importsFound[node.Path] = struct{}{}
		}
	case *parser.InfixExpression:
		c.Check(node.Left)
		c.Check(node.Right)
	case *parser.DefStatement:
		if node.Global && node.Constant {
			c.constVars[node.Name.Value] = struct{}{}
		}
		c.varTypes[node.Name.Value] = node.Type
		c.Check(node.Right)
	case *parser.PrefixExpression:
		c.Check(node.Right)
	case *parser.AssignmentExpression:
		name := c.getIdentNameAssign(node.Left, true)
		if _, ok := c.constVars[name]; ok {
			c.appendError(node.Position(), "cannot assign to constant variable '%s'\n", name)
		}
		c.Check(node.Left)
		c.Check(node.Right)
	case *parser.StructInitializationExpression:
		for _, e := range node.Values {
			c.Check(e)
		}
	case *parser.ArrayLiteral:
		for _, e := range node.Items {
			c.Check(e)
		}
	case *parser.CallExpression:
		for _, arg := range node.Params {
			c.Check(arg)
		}

		if _, ok := c.importsFound["io"]; ok && (node.Function.Value == "print" || node.Function.Value == "println") {
			c.checkPrintArgs(node)
		}

		var moduleName string
		moduleFound := false
		for name, module := range c.builtinNames {
			if _, ok := module[node.Function.Value]; ok {
				moduleName = name
				moduleFound = true
				break
			}
		}
		if _, ok := c.importsFound[moduleName]; (ok && moduleFound) || !moduleFound {
			return
		}
		c.appendError(node.Position(), "stdlib function '%s' used without stdlib module '%s' imported, did you mean to do this?\n", node.Function.Value, moduleName)
	case *parser.StructStatement:
		c.structFieldTypes[node.Name] = make(map[string]lexer.VarType)
		for name, idx := range node.Names {
			c.structFieldTypes[node.Name][name] = node.Types[idx]
		}
	}
}

func (c *Checker) appendError(pos *util.Position, msg string, args ...any) {
	c.Errors = append(c.Errors, util.PositionError{
		Position: pos,
		Msg:      fmt.Sprintf(msg, args...),
	})
}

func (c *Checker) getIdentNameAssign(expr parser.Expression, error bool) string {
	switch e := expr.(type) {
	case *parser.IdentifierExpression:
		return e.Value
	case *parser.DereferenceExpression:
		if vi, ok := e.Var.(*parser.IdentifierExpression); ok {
			return vi.Value
		} else if di, ok := e.Var.(*parser.DereferenceExpression); ok {
			return c.getIdentNameAssign(di, error)
		}
	case *parser.InfixExpression:
		if e.Operator == "." {
			return c.getIdentNameAssign(e.Left, error)
		}
	default:
		if error {
			c.appendError(expr.Position(), "unknown node %T on lhs of assignment", expr)
			return ""
		}
	}
	return ""
}

func (c *Checker) getVarType(expr parser.Expression) (lexer.VarType, bool) {
	switch e := expr.(type) {
	case *parser.IdentifierExpression:
		vt, ok := c.varTypes[e.Value]
		return vt, ok
	case *parser.DereferenceExpression:
		if vi, ok := e.Var.(*parser.IdentifierExpression); ok {
			vt, ok := c.varTypes[vi.Value]
			return vt, ok
		} else {
			return c.getVarType(e.Var)
		}
	case *parser.InfixExpression:
		if e.Operator == "." {
			// NOTE: we assume this is correct, i haven't implemented checking at the time of writing but ehh, the X:{4i32}.a usecase is kinda dogshit and i'm not sure that's something i'd like to support.. it's not like structs have constructors or anything to make this something you'd want to do ..
			vt, ok := c.varTypes[e.Left.(*parser.IdentifierExpression).Value]
			if !ok || !vt.IsStructType {
				return lexer.VarType{}, false
			}
			fieldTypes, ok := c.structFieldTypes[vt.StructName]
			if !ok {
				return lexer.VarType{}, false
			}
			ft, ok := fieldTypes[e.Right.(*parser.IdentifierExpression).Value]
			return ft, ok
		}
	case *parser.BooleanExpression:
		return lexer.VarType{Base: lexer.Bool}, true
	case *parser.IntegerLiteral:
		return e.Type, true
	case *parser.StringLiteral:
		return lexer.VarType{Base: lexer.Char, Pointer: 1}, true
	case *parser.FloatLiteral:
		return e.Type, true
	}
	return lexer.VarType{}, false
}

func (c *Checker) checkPrintArgs(node *parser.CallExpression) {
	var fmtStr string
	if s, ok := node.Params[0].(*parser.StringLiteral); ok {
		fmtStr = s.Value
	} else {
		c.appendError(node.Position(), "first argument of print/ln function should be string literal")
		return
	}
	r, _ := regexp.Compile("%(?:(?:b|c|s)|(?:(?:f|u|fu|)(?:y|w|d|l)))")
	found := r.FindAllString(fmtStr, -1)
	if len(found) != len(node.Params)-1 {
		c.appendError(node.Position(), "print/ln function should have as many specifiers as arguments given, %d specifiers, %d argument", len(found), len(node.Params)-1)
		return
	}

	for i := range len(found) {
		specifier := found[i]
		arg := node.Params[i+1]
		switch specifier {
		case "%b":
			c.checkSpecifier(arg, lexer.Bool, 0, i, specifier, node.Position())
		case "%c":
			c.checkSpecifier(arg, lexer.Char, 0, i, specifier, node.Position())
		case "%s":
			c.checkSpecifier(arg, lexer.Char, 1, i, specifier, node.Position())
		case "%y", "%fy":
			c.checkSpecifier(arg, lexer.Int8, 0, i, specifier, node.Position())
		case "%uy", "%fuy":
			c.checkSpecifier(arg, lexer.Uint8, 0, i, specifier, node.Position())
		case "%w", "%fw":
			c.checkSpecifier(arg, lexer.Int16, 0, i, specifier, node.Position())
		case "%uw", "%fuw":
			c.checkSpecifier(arg, lexer.Uint16, 0, i, specifier, node.Position())
		case "%d", "%fd":
			c.checkSpecifier(arg, lexer.Int32, 0, i, specifier, node.Position())
		case "%ud", "%fud":
			c.checkSpecifier(arg, lexer.Uint32, 0, i, specifier, node.Position())
		case "%l", "%fl":
			c.checkSpecifier(arg, lexer.Int, 0, i, specifier, node.Position())
		case "%ul", "%ful":
			c.checkSpecifier(arg, lexer.Uint, 0, i, specifier, node.Position())
		}
	}
}

func (c *Checker) checkSpecifier(arg parser.Expression, expectedType lexer.BaseVarType, expectedPointer uint8, idx int, specifier string, position *util.Position) {
	typ, ok := c.getVarType(arg)
	if !ok || typ.Base != expectedType || typ.Pointer != expectedPointer {
		expectedVt := lexer.VarType{Base: expectedType, Pointer: expectedPointer}
		c.appendError(position, "print/ln arg %d with specifier %s isnt't %s, has type: %s", idx, specifier, expectedVt, typ.String())
	}
}
