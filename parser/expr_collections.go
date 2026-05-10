package parser

import "grianlang3/lexer"

func (p *Parser) parseArrayLiteral() Expression {
	lit := &ArrayLiteral{Token: p.currToken}
	// assumess curr = [
	p.NextToken()
	vt, err := p.parseType()
	if err != nil {
		p.appendError(lit.Position(), "expected type after [ in array literal expr")
	}
	if !p.expectCurr(lexer.SEMICOLON) {
		return nil
	}
	lit.Type = vt
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
