package parser

import (
	"grianlang3/lexer"
	"grianlang3/util"
)

func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Literal,
	}

	p.NextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
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

func (p *Parser) parseSizeofExpression() Expression {
	expr := &SizeofExpression{Token: p.currToken, position: util.Position{
		StartLine: p.currToken.Position.StartLine,
		StartCol:  p.currToken.Position.StartCol,
	}}
	p.NextToken() // past sizeof
	vt, ok := p.parseType()
	if !ok {
		expr.Position().CopyEnd(&p.currToken.Position)
		p.appendError(&expr.position, "expected type after sizeof keyword")
		return nil
	}

	expr.Type = vt
	expr.Position().CopyEnd(&p.currToken.Position)

	return expr
}
