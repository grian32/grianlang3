package parser

import (
	"gl3/lexer"
	"gl3/util"
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
		p.appendError(&lit.Token.Position, "expected ; after type in array literal expr")
		lit.Position().CopyEnd(&p.currToken.Position)
		return nil
	}
	lit.Type = vt

	items, ok := p.parseExpressionList(lexer.RBRACKET)
	if !ok {
		p.appendError(&lit.Token.Position, "expected ] or , in array literal expr")
		lit.Position().CopyEnd(&p.currToken.Position)
		return nil
	}
	lit.Items = items

	p.NextToken()

	lit.Position().CopyEnd(&p.currToken.Position)
	return lit
}
