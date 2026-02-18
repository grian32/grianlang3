package parser

import (
	"fmt"
	"grianlang3/lexer"
	"strconv"
)

const (
	_ byte = iota
	LOWEST
	ASSIGN
	LOR
	LAND
	EQUALS      // ==
	LESSGREATER // > or <
	CAST
	SUM     // +
	PRODUCT // *
	PREFIX  // -X or !X
	CALL
	INDEX
)

var precedences = map[lexer.TokenType]byte{
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.ASTERISK: PRODUCT,
	lexer.SLASH:    PRODUCT,
	lexer.LPAREN:   CALL,
	lexer.DOT:      CALL, // same semantic as c
	lexer.COLON:    CALL, // same semantic as c
	lexer.ASSIGN:   ASSIGN,
	lexer.NOT:      PREFIX,
	lexer.LOR:      LOR,
	lexer.LAND:     LAND,
	lexer.EQ:       EQUALS,
	lexer.NOTEQ:    EQUALS,
	lexer.GT:       LESSGREATER,
	lexer.LT:       LESSGREATER,
	lexer.GTEQ:     LESSGREATER,
	lexer.LTEQ:     LESSGREATER,
	lexer.LBRACKET: INDEX,
	lexer.AS:       CAST,
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
	p.prefixParseFns[lexer.FLOAT] = p.parseFloatLiteral
	p.prefixParseFns[lexer.STRING] = p.parseStringLiteral
	p.prefixParseFns[lexer.MINUS] = p.parsePrefixExpression
	p.prefixParseFns[lexer.IDENTIFIER] = p.parseIdentifier
	p.prefixParseFns[lexer.LPAREN] = p.parseGroupedExpression
	p.prefixParseFns[lexer.AMPERSAND] = p.parseReference
	p.prefixParseFns[lexer.ASTERISK] = p.parseDereference
	p.prefixParseFns[lexer.TRUE] = p.parseBoolean
	p.prefixParseFns[lexer.FALSE] = p.parseBoolean
	p.prefixParseFns[lexer.NOT] = p.parsePrefixExpression
	p.prefixParseFns[lexer.SIZEOF] = p.parseSizeofExpression
	p.prefixParseFns[lexer.LBRACKET] = p.parseArrayLiteral
	p.prefixParseFns[lexer.CHAR] = p.parseCharLiteral

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.infixParseFns[lexer.PLUS] = p.parseInfixExpression
	p.infixParseFns[lexer.MINUS] = p.parseInfixExpression
	p.infixParseFns[lexer.SLASH] = p.parseInfixExpression
	p.infixParseFns[lexer.ASTERISK] = p.parseInfixExpression
	p.infixParseFns[lexer.LAND] = p.parseInfixExpression
	p.infixParseFns[lexer.LOR] = p.parseInfixExpression
	p.infixParseFns[lexer.EQ] = p.parseInfixExpression
	p.infixParseFns[lexer.LT] = p.parseInfixExpression
	p.infixParseFns[lexer.GT] = p.parseInfixExpression
	p.infixParseFns[lexer.LTEQ] = p.parseInfixExpression
	p.infixParseFns[lexer.GTEQ] = p.parseInfixExpression
	p.infixParseFns[lexer.NOTEQ] = p.parseInfixExpression
	p.infixParseFns[lexer.DOT] = p.parseInfixExpression
	p.infixParseFns[lexer.LPAREN] = p.parseCallExpression
	p.infixParseFns[lexer.ASSIGN] = p.parseAssignExpression
	p.infixParseFns[lexer.AS] = p.parseCastExpression
	p.infixParseFns[lexer.LBRACKET] = p.parseArrayIndexExpression
	p.infixParseFns[lexer.COLON] = p.parseStructInitialization

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

func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Literal,
	}

	p.NextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseIdentifier() Expression {
	expr := &IdentifierExpression{Token: p.currToken, Value: p.currToken.Literal}
	p.NextToken()
	return expr
}

func (p *Parser) parseGroupedExpression() Expression {
	p.NextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectCurr(lexer.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIntegerLiteral() Expression {
	vt := lexer.VarType{Base: lexer.Int, Pointer: 0}
	lit := &IntegerLiteral{Token: p.currToken, Type: vt}

	// could be more efficient by parsing based on type or the lack thereof but this makes for a decent chunk cleaner
	// code
	value, err := strconv.ParseInt(p.currToken.Literal, 0, 64)
	if err != nil {
		p.Errors = append(p.Errors, fmt.Sprintf("could not parse %q as integer", p.currToken.Literal))
	}
	uvalue, err := strconv.ParseUint(p.currToken.Literal, 0, 64)
	if err != nil {
		p.Errors = append(p.Errors, fmt.Sprintf("could not parse %q as unsigned integer", p.currToken.Literal))
	}

	lit.Value = value
	lit.UValue = uvalue

	p.NextToken()
	if p.currTokenIs(lexer.IDENTIFIER) {
		switch p.currToken.Literal {
		case "i32":
			lit.Type.Base = lexer.Int32
		case "i16":
			lit.Type.Base = lexer.Int16
		case "i8":
			lit.Type.Base = lexer.Int8
		case "u32":
			lit.Type.Base = lexer.Uint32
		case "u16":
			lit.Type.Base = lexer.Uint16
		case "u8":
			lit.Type.Base = lexer.Uint8
		case "u64":
			lit.Type.Base = lexer.Uint
		}
		p.NextToken()
	}

	return lit
}

func (p *Parser) parseStringLiteral() Expression {
	expr := &StringLiteral{Token: p.currToken, Value: p.currToken.Literal + "\000"}
	p.NextToken()
	return expr
}

func (p *Parser) parseSizeofExpression() Expression {
	expr := &SizeofExpression{Token: p.currToken}
	var vt lexer.VarType
	p.NextToken() // past sizeof
	if p.currTokenIs(lexer.TYPE) {
		vt = p.currToken.VarType
	} else if p.currTokenIs(lexer.IDENTIFIER) {
		vt = lexer.VarType{
			IsStructType: true,
			StructName:   p.currToken.Literal,
		}
	} else {
		return nil
	}
	p.NextToken() // past type/ident
	p.getPointers(&vt)

	expr.Type = vt
	p.NextToken()

	return expr
}

func (p *Parser) parseFloatLiteral() Expression {
	vt := lexer.VarType{Base: lexer.Float, Pointer: 0}
	lit := &FloatLiteral{Token: p.currToken, Type: vt}

	value, err := strconv.ParseFloat(p.currToken.Literal, 32)
	if err != nil {
		p.Errors = append(p.Errors, fmt.Sprintf("could not parse %q as float", p.currToken.Literal))
	}

	lit.Value = float32(value)

	return lit
}

func (p *Parser) parseCharLiteral() Expression {
	vt := lexer.VarType{Base: lexer.Int8, Pointer: 0}
	expr := &IntegerLiteral{Token: p.currToken, Value: int64(p.currToken.Literal[0]), Type: vt}
	p.NextToken()
	return expr
}

func (p *Parser) parseCastExpression(left Expression) Expression {
	expr := &CastExpression{Token: p.currToken}
	expr.Expr = left

	p.NextToken() // asvance past AS
	var castType lexer.VarType
	if p.currTokenIs(lexer.TYPE) {
		castType = p.currToken.VarType
	} else if p.currTokenIs(lexer.IDENTIFIER) {
		castType = lexer.VarType{
			IsStructType: true,
			StructName:   p.currToken.Literal,
		}
	} else {
		return nil
	}
	p.NextToken() // advance past type/ident
	p.getPointers(&castType)
	expr.Type = castType

	return expr
}

func (p *Parser) parseStructInitialization(left Expression) Expression {
	exp := &StructInitializationExpression{Token: p.currToken}
	if ident, ok := left.(*IdentifierExpression); ok {
		exp.Name = ident.Value
	} else {
		p.Errors = append(p.Errors, "expected identifier on lhs of struct init")
		return nil
	}
	p.NextToken() // skip past :
	if !p.expectCurr(lexer.LBRACE) {
		p.NextToken()
	}

	for !p.currTokenIs(lexer.RBRACE) {
		expr := p.parseExpression(LOWEST)
		exp.Values = append(exp.Values, expr)
		if p.currTokenIs(lexer.RBRACE) {
			p.NextToken()
			break
		} else if p.currTokenIs(lexer.COMMA) {
			p.NextToken()
			continue
		} else {
			return nil
		}
	}
	p.NextToken()

	return exp
}

func (p *Parser) parseBoolean() Expression {
	expr := &BooleanExpression{Token: p.currToken}

	if p.currTokenIs(lexer.TRUE) {
		expr.Value = true
	} else {
		expr.Value = false
	}
	p.NextToken()

	return expr
}

func (p *Parser) parseReference() Expression {
	expr := &ReferenceExpression{Token: p.currToken}
	p.NextToken()
	rhs := p.parseExpression(LOWEST)
	if ident, ok := rhs.(*IdentifierExpression); ok {
		expr.Var = ident
	}
	return expr
}

func (p *Parser) parseAssignExpression(left Expression) Expression {
	expr := &AssignmentExpression{Token: p.currToken}
	switch left.(type) {
	case *IdentifierExpression, *DereferenceExpression:
		expr.Left = left
	default:
		if infix, ok := left.(*InfixExpression); ok && infix.Operator == "." {
			expr.Left = left
		} else {
			p.Errors = append(p.Errors, fmt.Sprintf("got %T on lhs of assignment, expected ident or deref", left))
		}
	}
	p.NextToken()

	expr.Right = p.parseExpression(LOWEST)

	return expr
}

func (p *Parser) parseDereference() Expression {
	expr := &DereferenceExpression{Token: p.currToken}

	p.NextToken()
	if p.currTokenIs(lexer.IDENTIFIER) {
		expr.Var = &IdentifierExpression{Token: p.currToken, Value: p.currToken.Literal}
		p.NextToken()
	} else {
		expr.Var = p.parseExpression(PREFIX)
	}

	return expr
}

// parseArrayIndexExpression, this is rather dodgy as it basically attempts to be a sugar for deref + pointer arithmetic
// to keep the same semantics
func (p *Parser) parseArrayIndexExpression(left Expression) Expression {
	derefToken := p.currToken
	p.NextToken() // skip past [
	index := p.parseExpression(LOWEST)
	if !p.currTokenIs(lexer.RBRACKET) {
		return nil
	}
	p.NextToken()

	return &DereferenceExpression{
		Token: derefToken,
		Var:   &InfixExpression{Token: p.currToken, Left: left, Operator: "+", Right: index},
	}
}

func (p *Parser) parseArrayLiteral() Expression {
	lit := &ArrayLiteral{Token: p.currToken}
	// assumess curr = [
	p.NextToken()
	vt := p.currToken.VarType
	p.getPointers(&vt)
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}
	lit.Type = vt
	p.NextToken()

	lit.Items = []Expression{}

	for !p.currTokenIs(lexer.RBRACKET) {
		expr := p.parseExpression(LOWEST)
		lit.Items = append(lit.Items, expr)
		if p.currTokenIs(lexer.RBRACKET) {
			break
		} else if p.currTokenIs(lexer.COMMA) {
			p.NextToken()
			continue
		} else {
			return nil
		}
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
		// next token should be done by each individual parsing function as necessary, doing it this way
		// introduces rather strange bugs
		//p.NextToken()
		exp.Params = append(exp.Params, expr)
		if p.currTokenIs(lexer.RPAREN) {
			p.NextToken()
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
	switch p.currToken.Type {
	case lexer.DEF:
		return p.parseVarStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.FNC:
		return p.parseFunctionStatement()
	case lexer.IMPORT:
		return p.parseImportStatement()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.WHILE:
		return p.parseWhileStatement()
	case lexer.STRUCT:
		return p.parseStructStatement()
	}

	return p.parseExpressionStatement()
}

func (p *Parser) parseStructStatement() Statement {
	stmt := &StructStatement{Token: p.currToken}
	p.NextToken()
	if !p.currTokenIs(lexer.IDENTIFIER) {
		p.Errors = append(p.Errors, "expected identifier after struct keyword")
		return nil
	}
	stmt.Name = p.currToken.Literal
	p.NextToken()
	if !p.expectCurr(lexer.LBRACE) {
		return nil
	}
	stmt.Names = make(map[string]int)
	for !p.currTokenIs(lexer.RBRACE) {
		var vt lexer.VarType

		if p.currTokenIs(lexer.TYPE) {
			vt = p.currToken.VarType
			p.NextToken()
		} else if p.currTokenIs(lexer.IDENTIFIER) {
			vt = lexer.VarType{
				IsStructType: true,
				StructName:   p.currToken.Literal,
			}
			p.NextToken()
		} else {
			p.Errors = append(p.Errors, "expected type in struct definition")
			return nil
		}

		if !p.currTokenIs(lexer.IDENTIFIER) {
			p.Errors = append(p.Errors, "expected identifier after type in struct definition")
			return nil
		}
		stmt.Types = append(stmt.Types, vt)
		stmt.Names[p.currToken.Literal] = len(stmt.Types) - 1
		p.NextToken()
	}
	return stmt
}

func (p *Parser) parseWhileStatement() Statement {
	stmt := &WhileStatement{Token: p.currToken}
	p.NextToken()
	cond := p.parseExpression(LOWEST)
	stmt.Condition = cond
	if !p.expectCurr(lexer.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	if !p.expectCurr(lexer.RBRACE) {
		return nil
	}
	return stmt
}

func (p *Parser) parseIfStatement() Statement {
	stmt := &IfStatement{Token: p.currToken}
	p.NextToken() // past IF token
	cond := p.parseExpression(LOWEST)
	stmt.Condition = cond
	if !p.expectCurr(lexer.LBRACE) {
		return nil
	}
	stmt.Success = p.parseBlockStatement()
	if !p.expectCurr(lexer.RBRACE) {
		return nil
	}

	if !p.currTokenIs(lexer.ELSE) {
		return stmt
	}
	p.NextToken()
	if !p.expectCurr(lexer.LBRACE) {
		return nil
	}
	stmt.Fail = p.parseBlockStatement()
	if !p.expectCurr(lexer.RBRACE) {
		return nil
	}

	return stmt
}

func (p *Parser) parseImportStatement() Statement {
	stmt := &ImportStatement{Token: p.currToken}
	p.NextToken()
	if !p.currTokenIs(lexer.STRING) {
		return nil
	}
	stmt.Path = p.currToken.Literal

	return stmt
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
		if p.currTokenIs(lexer.SEMICOLON) {
			p.NextToken()
		}
	}

	return bs
}

func (p *Parser) getPointers(vt *lexer.VarType) {
	if !p.currTokenIs(lexer.ASTERISK) {
		return
	}

	for p.currTokenIs(lexer.ASTERISK) {
		vt.Pointer++
		p.NextToken()
	}
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
		var paramType lexer.VarType
		if p.peekTokenIs(lexer.TYPE) {
			p.NextToken()
			paramType = p.currToken.VarType
		} else if p.peekTokenIs(lexer.IDENTIFIER) {
			p.NextToken()
			paramType = lexer.VarType{
				IsStructType: true,
				StructName:   p.currToken.Literal,
			}
		} else {
			return nil
		}
		p.getPointers(&paramType)
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
	if p.peekTokenIs(lexer.TYPE) {
		p.NextToken()
		stmt.Type = p.currToken.VarType
	} else if p.peekTokenIs(lexer.IDENTIFIER) {
		p.NextToken()
		stmt.Type = lexer.VarType{
			IsStructType: true,
			StructName:   p.currToken.Literal,
		}
	} else {
		return nil
	}
	p.getPointers(&stmt.Type)

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

func (p *Parser) parseVarStatement() *DefStatement {
	stmt := &DefStatement{Token: p.currToken}
	if p.peekTokenIs(lexer.TYPE) {
		p.NextToken()
		stmt.Type = p.currToken.VarType
	} else if p.peekTokenIs(lexer.IDENTIFIER) {
		p.NextToken()
		stmt.Type = lexer.VarType{
			IsStructType: true,
			StructName:   p.currToken.Literal,
		}
	} else {
		return nil
	}
	p.getPointers(&stmt.Type)

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

	for !p.peekTokenIs(lexer.SEMICOLON) && precendence < p.currPrecedence() {
		infix := p.infixParseFns[p.currToken.Type]
		if infix == nil {
			return leftExp
		}

		//if p.currTokenIs(lexer.SEMICOLON) {
		//	p.NextToken()
		//}

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
	msg := fmt.Sprintf("no prefix parse function for %s, peek=%s found", t.Type.String(), p.peekToken.Type.String())
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

func (p *Parser) expectCurr(t lexer.TokenType) bool {
	if p.currTokenIs(t) {
		p.NextToken()
		return true
	}

	p.currError(t)
	return false
}

func (p *Parser) currError(t lexer.TokenType) {
	p.Errors = append(p.Errors, fmt.Sprintf("expected curr token to be %s, got %s instead", t, p.peekToken.Type))
}
