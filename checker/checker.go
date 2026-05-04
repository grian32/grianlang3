package checker

import (
	"fmt"
	"grianlang3/emitter"
	"grianlang3/lexer"
	"grianlang3/parser"
	"grianlang3/util"
	"regexp"
	"strings"
)

type Checker struct {
	importsFound map[string]struct{}
	builtinNames map[string]map[string]struct{}
	constVars    map[string]struct{}
	varTypes     map[string]lexer.VarType
	Errors       []util.PositionError
}

func New() *Checker {
	return &Checker{
		builtinNames: emitter.GetBuiltinNames(),
		importsFound: make(map[string]struct{}),
		constVars:    make(map[string]struct{}),
		varTypes:     make(map[string]lexer.VarType),
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
			fmt.Printf("%v\n", node)
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
	}
}

func (c *Checker) appendError(pos *util.Position, msg string, args ...any) {
	c.Errors = append(c.Errors, util.PositionError{
		Position: pos,
		Msg:      fmt.Sprintf(msg, args...),
	})
}

func (c *Checker) appendErrorRaw(pos *util.Position, msg string) {
	c.Errors = append(c.Errors, util.PositionError{
		Position: pos,
		Msg:      msg,
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

func (c *Checker) getIdentType(expr parser.Expression) (lexer.VarType, bool) {
	switch e := expr.(type) {
	case *parser.IdentifierExpression:
		vt, ok := c.varTypes[e.Value]
		return vt, ok
	case *parser.DereferenceExpression:
		if vi, ok := e.Var.(*parser.IdentifierExpression); ok {
			vt, ok := c.varTypes[vi.Value]
			return vt, ok
		} else {
			return c.getIdentType(e.Var)
		}
	case *parser.InfixExpression:
		if e.Operator == "." {
			return c.getIdentNameAssign(e.Left, error)
		}
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
		c.appendError(node.Position(), "print/ln function should have as many specifiers as arguments given")
		return
	}

	for i := range len(found) {
		specifier := found[i]
		arg := node.Params[i+1]
		switch specifier {
		case "%b":
			if _, ok := arg.(*parser.BooleanExpression); !ok {
				identName := c.getIdentNameAssign(arg, false)
				vt, ok := c.varTypes[identName]
				// TODO: maybe some better way to handle this?
				var errMsg string = fmt.Sprintf("printf arg %d with specifier %%b isnt't bool, has type: %T", i, arg)
				if ok && (vt.Base != lexer.Bool || vt.Pointer > 0) {
					errMsg = fmt.Sprintf("printf arg %d with specifier %%b isnt't variable with type bool, has type: %s", i, vt.String())
				}
				c.appendErrorRaw(node.Position(), errMsg)
			}
		}
	}
}
