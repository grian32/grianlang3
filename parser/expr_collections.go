package parser

import "grianlang3/lexer"

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
	p.NextToken()

	return lit
}
