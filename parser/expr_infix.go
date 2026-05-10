package parser

import (
	"grianlang3/lexer"
	"grianlang3/util"
)

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

func (p *Parser) parseCastExpression(left Expression) Expression {
	expr := &CastExpression{Token: p.currToken, position: util.Position{
		StartLine: p.currToken.Position.StartLine,
		StartCol:  p.currToken.Position.StartCol,
	}}
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
	expr.Position().CopyEnd(&p.currToken.Position)

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
			p.appendError(left.Position(), "got %T on lhs of assignment, expected ident or deref", left)
		}
	}
	p.NextToken()

	expr.Right = p.parseExpression(LOWEST)

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

func (p *Parser) parseCallExpression(left Expression) Expression {
	leftPos := left.Position()
	exp := &CallExpression{Token: p.currToken, position: util.Position{
		StartLine: leftPos.StartLine,
		StartCol:  leftPos.StartCol,
	}}
	if identExpr, ok := left.(*IdentifierExpression); ok {
		exp.Function = identExpr
	} else {
		return nil
	}
	if !p.expectCurr(lexer.LPAREN) {
		return nil
	}

	exp.Params = []Expression{}

	for !p.currTokenIs(lexer.RPAREN) {
		expr := p.parseExpression(LOWEST)
		// next token should be done by each individual parsing function as necessary, doing it this way
		// introduces rather strange bugs
		//p.NextToken()
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
	exp.Position().CopyEnd(&p.currToken.Position)
	p.NextToken()

	return exp
}

func (p *Parser) parseStructInitialization(left Expression) Expression {
	exp := &StructInitializationExpression{Token: p.currToken}
	if ident, ok := left.(*IdentifierExpression); ok {
		exp.Name = ident.Value
	} else {
		p.appendError(&p.currToken.Position, "expected identifier on lhs of struct init")
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
