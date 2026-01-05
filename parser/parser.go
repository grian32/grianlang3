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
	lexer.PLUS:   SUM,
	lexer.LPAREN: CALL,
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
	p.prefixParseFns[lexer.IDENTIFIER] = p.parseIdentifier
	//p.prefixParseFns[lexer.LPAREN] = p.parseGroupedExpression

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.infixParseFns[lexer.PLUS] = p.parseInfixExpression
	p.infixParseFns[lexer.LPAREN] = p.parseCallExpression

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

func (p *Parser) parseIdentifier() Expression {
	return &IdentifierExpression{Token: p.currToken, Value: p.currToken.Literal}
}

func (p *Parser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{Token: p.currToken, Type: lexer.Int}

	value, err := strconv.ParseInt(p.currToken.Literal, 0, 64)
	if err != nil {
		p.Errors = append(p.Errors, fmt.Sprintf("could not parse %q as integer", p.currToken.Literal))
	}

	lit.Value = value
	if p.peekTokenIs(lexer.IDENTIFIER) && p.peekToken.Literal == "i32" {
		p.NextToken()
		lit.Type = lexer.Int32
	}

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

func (p *Parser) parseCallExpression(left Expression) Expression {
	exp := &CallExpression{Token: p.currToken}
	if identExpr, ok := left.(*IdentifierExpression); ok {
		exp.Function = identExpr
	} else {
		return nil
	}
	p.NextToken()

	exp.Params = []Expression{}

	for !p.currTokenIs(lexer.RPAREN) {
		expr := p.parseExpression(LOWEST)
		p.NextToken()
		exp.Params = append(exp.Params, expr)
		if p.currTokenIs(lexer.RPAREN) {
			break
		} else if p.currTokenIs(lexer.COMMA) {
			p.NextToken()
			continue
		} else {
			return nil
		}
	}

	return exp
}

func (p *Parser) parseStatement() Statement {
	if p.currTokenIs(lexer.DEF) {
		return p.parseVarStatement()
	} else if p.currTokenIs(lexer.IDENTIFIER) && p.peekTokenIs(lexer.ASSIGN) {
		return p.parseAssignStatement()
	} else if p.currTokenIs(lexer.RETURN) {
		return p.parseReturnStatement()
	} else if p.currTokenIs(lexer.FNC) {
		return p.parseFunctionStatement()
	}

	return p.parseExpressionStatement()
}

func (p *Parser) parseReturnStatement() Statement {
	stmt := &ReturnStatement{Token: p.currToken}
	p.NextToken()
	expr := p.parseExpression(LOWEST)
	stmt.Expr = expr
	return stmt
}

func (p *Parser) parseBlockStatement() *BlockStatement {
	bs := &BlockStatement{Token: p.currToken}
	bs.Statements = []Statement{}

	for !p.currTokenIs(lexer.RBRACE) {
		stmt := p.parseStatement()
		if stmt != nil {
			bs.Statements = append(bs.Statements, stmt)
		}
		p.NextToken()
	}

	return bs
}

func (p *Parser) parseFunctionStatement() Statement {
	stmt := &FunctionStatement{Token: p.currToken}

	if !p.expectPeek(lexer.IDENTIFIER) {
		return nil
	}
	stmt.Name = &IdentifierExpression{Token: p.currToken, Value: p.currToken.Literal}
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	stmt.Params = []FunctionParameter{}

	// for empty arg list if it is rparen then it just stops immediately since we curr are on lparen
	for !p.peekTokenIs(lexer.RPAREN) {
		if !p.expectPeek(lexer.TYPE) {
			return nil
		}
		paramType := p.currToken.VarType
		if !p.expectPeek(lexer.IDENTIFIER) {
			return nil
		}
		ident := &IdentifierExpression{Token: p.currToken, Value: p.currToken.Literal}
		p.NextToken()

		param := FunctionParameter{
			Type: paramType,
			Name: ident,
		}

		stmt.Params = append(stmt.Params, param)
		if p.currTokenIs(lexer.RPAREN) {
			break
		} else if p.currTokenIs(lexer.COMMA) {
			continue
		} else {
			return nil
		}
	}

	if len(stmt.Params) == 0 {
		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	} else {
		if !p.currTokenIs(lexer.RPAREN) {
			return nil
		}
	}

	if !p.expectPeek(lexer.ARROW) {
		return nil
	}
	if !p.expectPeek(lexer.TYPE) {
		return nil
	}
	stmt.Type = p.currToken.VarType

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}
	p.NextToken()
	stmt.Body = p.parseBlockStatement()
	if !p.currTokenIs(lexer.RBRACE) {
		return nil
	}

	return stmt
}

func (p *Parser) parseAssignStatement() *AssignmentStatement {
	stmt := &AssignmentStatement{Token: p.currToken}
	stmt.Name = &IdentifierExpression{Token: p.currToken, Value: p.currToken.Literal}

	p.NextToken() // skip ident
	p.NextToken() // skip assign, prechecked by parseStatement

	stmt.Right = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseVarStatement() *DefStatement {
	stmt := &DefStatement{Token: p.currToken}

	if !p.expectPeek(lexer.TYPE) {
		return nil
	}

	stmt.Type = p.currToken.VarType

	if !p.expectPeek(lexer.IDENTIFIER) {
		return nil
	}

	stmt.Name = &IdentifierExpression{Token: p.currToken, Value: p.currToken.Literal}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.NextToken()

	stmt.Right = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.NextToken()
	}

	return stmt
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

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.NextToken()
		return true
	}

	p.peekError(t)
	return false
}
func (p *Parser) peekError(t lexer.TokenType) {
	p.Errors = append(p.Errors, fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type))
}
