package parser

import (
	"bytes"
	"fmt"
	"grianlang3/lexer"
	"strconv"
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
	Token  lexer.Token
	Type   lexer.VarType
	Value  int64
	UValue uint64
}

func (il *IntegerLiteral) expressionNode()      { /* noop */ }
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string {
	switch il.Type.Base {
	case lexer.Int8, lexer.Int16, lexer.Int32, lexer.Int:
		return strconv.Itoa(int(il.Value)) + "(" + il.Type.String() + ")"
	case lexer.Uint8, lexer.Uint16, lexer.Uint32, lexer.Uint:
		return strconv.FormatUint(il.UValue, 10) + "(" + il.Type.String() + ")"
	}
	return il.Token.Literal + "(" + il.Type.String() + ")"
}

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
	Name  *IdentifierExpression
	Type  lexer.VarType
	Right Expression
}

func (ds *DefStatement) statementNode()       { /* noop */ }
func (ds *DefStatement) TokenLiteral() string { return ds.Token.Literal }
func (ds *DefStatement) String() string {
	return "def " + ds.Type.String() + " " + ds.Name.String() + " = " + ds.Right.String()
}

type AssignmentExpression struct {
	Token lexer.Token
	Left  Expression
	Right Expression
}

func (ae *AssignmentExpression) expressionNode()      { /* noop */ }
func (ae *AssignmentExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AssignmentExpression) String() string {
	return ae.Left.String() + " = " + ae.Right.String()
}

type IdentifierExpression struct {
	Token lexer.Token
	Value string
}

func (ie *IdentifierExpression) expressionNode()      { /* noop */ }
func (ie *IdentifierExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IdentifierExpression) String() string       { return ie.Value }

type ReturnStatement struct {
	Token lexer.Token
	Expr  Expression
}

func (rs *ReturnStatement) statementNode()       { /* noop */ }
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string       { return "return " + rs.Expr.String() }

type BlockStatement struct {
	Token      lexer.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       { /* noop */ }
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for i, s := range bs.Statements {
		out.WriteString(s.String())
		if i != len(bs.Statements)-1 {
			out.WriteString(";")
		}
	}

	return out.String()
}

type FunctionParameter struct {
	Type lexer.VarType
	Name *IdentifierExpression
}

func (fp *FunctionParameter) String() string { return fp.Type.String() + " " + fp.Name.String() }

type FunctionStatement struct {
	Token  lexer.Token
	Name   *IdentifierExpression
	Type   lexer.VarType
	Params []FunctionParameter
	Body   *BlockStatement
}

func (fs *FunctionStatement) statementNode()       { /* noop */ }
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) String() string {
	var out bytes.Buffer

	out.WriteString("fnc " + fs.Name.String() + "(")

	for i, p := range fs.Params {
		out.WriteString(p.String())
		if i != len(fs.Params)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString(") -> " + fs.Type.String() + " { " + fs.Body.String() + " }")
	return out.String()
}

type CallExpression struct {
	Token    lexer.Token
	Function *IdentifierExpression
	Params   []Expression
}

func (ce *CallExpression) expressionNode()      { /* noop */ }
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ce.Function.String() + "(")

	for i, p := range ce.Params {
		out.WriteString(p.String())
		if i != len(ce.Params)-1 {
			out.WriteString(", ")
		}
	}

	out.WriteString(")")

	return out.String()
}

type ReferenceExpression struct {
	Token lexer.Token
	Var   *IdentifierExpression
}

func (re *ReferenceExpression) expressionNode()      { /* noop */ }
func (re *ReferenceExpression) TokenLiteral() string { return re.Token.Literal }
func (re *ReferenceExpression) String() string       { return "&" + re.Var.String() }

type DereferenceExpression struct {
	Token lexer.Token
	Var   Expression
}

func (de *DereferenceExpression) expressionNode()      { /* noop */ }
func (de *DereferenceExpression) TokenLiteral() string { return de.Token.Literal }
func (de *DereferenceExpression) String() string       { return "*" + de.Var.String() }

type BooleanExpression struct {
	Token lexer.Token
	Value bool
}

func (be *BooleanExpression) expressionNode()      { /* noop */ }
func (be *BooleanExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BooleanExpression) String() string       { return fmt.Sprintf("%t", be.Value) }

type PrefixExpression struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      { /* noop */ }
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string       { return "(" + pe.Operator + pe.Right.String() + ")" }

type CastExpression struct {
	Token lexer.Token
	Expr  Expression
	Type  lexer.VarType
}

func (ce *CastExpression) expressionNode()      { /* noop */ }
func (ce *CastExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CastExpression) String() string       { return ce.Expr.String() + " as " + ce.Type.String() }

type FloatLiteral struct {
	Token lexer.Token
	Value float32
	Type  lexer.VarType // opts : Float only
}

func (fl *FloatLiteral) expressionNode()      { /* noop */ }
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal + "(" + fl.Type.String() + ")" }

type SizeofExpression struct {
	Token lexer.Token
	Type  lexer.VarType
}

func (se *SizeofExpression) expressionNode()      { /* noop */ }
func (se *SizeofExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SizeofExpression) String() string       { return "sizeof " + se.Type.String() }

type ArrayLiteral struct {
	Token lexer.Token
	Type  lexer.VarType
	Items []Expression
}

func (al *ArrayLiteral) expressionNode()      { /* noop */ }
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("[")
	out.WriteString(al.Type.String())
	out.WriteString(";")

	for i, e := range al.Items {
		out.WriteString(e.String())
		if i != len(al.Items)-1 {
			out.WriteString(",")
		}
	}
	out.WriteString("]")

	return out.String()
}

type ImportStatement struct {
	Token lexer.Token
	Path  string
}

func (is *ImportStatement) statementNode()       { /* noop */ }
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string       { return "import \"" + is.Path + "\"" }

type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      { /* noop */ }
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

type IfStatement struct {
	Token     lexer.Token
	Condition Expression
	Success   *BlockStatement
	Fail      *BlockStatement
}

func (is *IfStatement) statementNode()       { /* noop */ }
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString("if ")
	out.WriteString(is.Condition.String())
	out.WriteString(" { ")
	out.WriteString(is.Success.String())
	out.WriteString(" }")
	if is.Fail != nil {
		out.WriteString(" else { ")
		out.WriteString(is.Fail.String())
		out.WriteString(" }")
	}

	return out.String()
}

type WhileStatement struct {
	Token     lexer.Token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       { /* noop */ }
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while ")
	out.WriteString(ws.Condition.String())
	out.WriteString(" { ")
	out.WriteString(ws.Body.String())
	out.WriteString(" }")

	return out.String()
}

type StructStatement struct {
	Token lexer.Token
	Name  string
	// this is done intentionally as in llvm terms theyre just indexes, so first component just maps to
	// 0, so map[compname]idx, types[idx]
	Types []lexer.VarType
	Names map[string]int
}

func (ss *StructStatement) statementNode()       { /* noop */ }
func (ss *StructStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *StructStatement) String() string {
	var out bytes.Buffer
	out.WriteString("struct ")
	out.WriteString(ss.Name)
	out.WriteString("{")
	for name, idx := range ss.Names {
		out.WriteString(ss.Types[idx].String())
		out.WriteString(" ")
		out.WriteString(name)
		out.WriteString(";")
	}
	out.WriteString("}")
	return out.String()
}

type StructInitializationExpression struct {
	Token  lexer.Token
	Name   string
	Values []Expression
}

func (sie *StructInitializationExpression) expressionNode()      { /* noop */ }
func (sie *StructInitializationExpression) TokenLiteral() string { return sie.Token.Literal }
func (sie *StructInitializationExpression) String() string {
	var out bytes.Buffer
	out.WriteString(sie.Name)
	out.WriteString(":{")
	for i, e := range sie.Values {
		out.WriteString(e.String())
		if i != len(sie.Values)-1 {
			out.WriteString(",")
		}
	}
	out.WriteString("}")
	return out.String()
}
