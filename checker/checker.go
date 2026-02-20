package checker

import (
	"fmt"
	"grianlang3/emitter"
	"grianlang3/parser"
	"strings"
)

type Checker struct {
	importsFound map[string]struct{}
	builtinNames map[string]map[string]struct{}
	Errors       []string
}

func New() *Checker {
	return &Checker{builtinNames: emitter.GetBuiltinNames()}
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
		c.Check(node.Right)
	case *parser.PrefixExpression:
		c.Check(node.Right)
	case *parser.AssignmentExpression:
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
		c.appendError("stdlib function '%s' used without stdlib module '%s' imported, did you mean to do this?\n", node.Function.Value, moduleName)
	}
}

func (c *Checker) appendError(msg string, args ...any) {
	c.Errors = append(c.Errors, fmt.Sprintf(msg, args...))
}
