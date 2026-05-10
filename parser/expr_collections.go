package parser

import (
	"grianlang3/lexer"
	"grianlang3/util"
)

func (p *Parser) parseArrayLiteral() Expression {
	lit := &ArrayLiteral{Token: p.currToken, position: util.Position{
		StartLine: p.currToken.Position.StartLine,
		StartCol:  p.currToken.Position.StartCol,
	}}
	// assumess curr = [
	p.NextToken()
	vt, ok := p.parseType()
	if !ok {
		lit.Position().CopyEnd(&p.currToken.Position)
		p.appendError(&lit.Token.Position, "expected type after [ in array literal expr")
		return nil
	}
	if !p.expectCurr(lexer.SEMICOLON) {
		lit.Position().CopyEnd(&p.currToken.Position)
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
			lit.Position().CopyEnd(&p.currToken.Position)
			return nil
		}
	}
	p.NextToken()

	lit.Position().CopyEnd(&p.currToken.Position)
	return lit
}
