package parser

import (
	"bytes"
	"grianlang3/lexer"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
		out.WriteString(";")
	}

	return out.String()
}

type ExpressionStatement struct {
	Token      lexer.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}

type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      { /* noop */ }
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type InfixExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      { /* noop */ }
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

type DefStatement struct {
	Token lexer.Token
	Name  string
	Type  lexer.VarType
	Right Expression
}

func (ds *DefStatement) statementNode()       { /* noop */ }
func (ds *DefStatement) TokenLiteral() string { return ds.Token.Literal }
func (ds *DefStatement) String() string {
	return "def " + ds.Type.String() + " " + ds.Name + " = " + ds.Right.String()
}

type AssignmentStatement struct {
	Token lexer.Token
	Name  string
	Right Expression
}

func (as *AssignmentStatement) statementNode()       { /* noop */ }
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignmentStatement) String() string {
	return as.Name + " = " + as.Right.String()
}

type RefExpression struct {
	Token lexer.Token
	Name  string
}

func (re *RefExpression) expressionNode()      { /* noop */ }
func (re *RefExpression) TokenLiteral() string { return re.Token.Literal }
func (re *RefExpression) String() string       { return re.Name }
