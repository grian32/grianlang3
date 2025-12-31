package parser

import (
	"fmt"
	"grianlang3/lexer"
	"strconv"
)

const (
	_ byte = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL
	INDEX
)

var precedences = map[lexer.TokenType]byte{
	lexer.PLUS: SUM,
}

type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

// Parser heavily based on the grpgscript parser https://github.com/grian32/grpg/grpgscript
type Parser struct {
	lexer *lexer.Lexer

	Errors []string

	currToken lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{lexer: l}

	p.NextToken()
	p.NextToken()

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.prefixParseFns[lexer.INT] = p.parseIntegerLiteral

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.infixParseFns[lexer.PLUS] = p.parseInfixExpression

	return p
}

func (p *Parser) NextToken() {
	p.currToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Literal,
		Left:     left,
	}

	precendece := p.currPrecedence()
	p.NextToken()
	expression.Right = p.parseExpression(precendece)

	return expression
}

func (p *Parser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{Token: p.currToken}

	value, err := strconv.ParseInt(p.currToken.Literal, 0, 64)
	if err != nil {
		p.Errors = append(p.Errors, fmt.Sprintf("could not parse %q as integer", p.currToken.Literal))
	}

	lit.Value = value

	return lit
}
func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = []Statement{}

	for !p.currTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.NextToken()
	}

	return program
}

func (p *Parser) parseStatement() Statement {
	return p.parseExpressionStatement()
}

func (p *Parser) peekTokenIs(tt lexer.TokenType) bool {
	return p.peekToken.Type == tt
}

func (p *Parser) currTokenIs(tt lexer.TokenType) bool {
	return p.currToken.Type == tt
}

func (p *Parser) parseExpressionStatement() Statement {
	stmt := &ExpressionStatement{Token: p.currToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.NextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precendence byte) Expression {
	prefix := p.prefixParseFns[p.currToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currToken)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.SEMICOLON) && precendence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.NextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) peekPrecedence() byte {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) currPrecedence() byte {
	if p, ok := precedences[p.currToken.Type]; ok {
		return p
	}

	return LOWEST
}
func (p *Parser) noPrefixParseFnError(t lexer.Token) {
	msg := fmt.Sprintf("no prefix parse function for %d found", t.Type)
	p.Errors = append(p.Errors, msg)
}
