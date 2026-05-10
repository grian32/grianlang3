package parser

import (
	"grianlang3/lexer"
	"grianlang3/util"
)

func (p *Parser) parseExternStatement() Statement {
	stmt := &ExternStructStatement{Token: p.currToken, position: util.Position{
		StartLine: p.currToken.Position.StartLine,
		StartCol:  p.currToken.Position.StartCol,
	}}
	p.NextToken()
	if !p.expectCurr(lexer.STRUCT) {
		return nil
	}

	if !p.currTokenIs(lexer.IDENTIFIER) {
		return nil
	}
	stmt.position.CopyEnd(&p.currToken.Position)
	stmt.Name = p.currToken.Literal
	p.NextToken()

	return stmt
}

func (p *Parser) parseStructStatement() Statement {
	stmt := &StructStatement{Token: p.currToken, position: util.Position{
		StartLine: p.currToken.Position.StartLine,
		StartCol:  p.currToken.Position.StartCol,
	}}
	p.NextToken()
	if !p.currTokenIs(lexer.IDENTIFIER) {
		p.appendError(&p.currToken.Position, "expected identifier after struct keyword")
		return nil
	}
	stmt.Name = p.currToken.Literal
	p.NextToken()
	if !p.expectCurr(lexer.LBRACE) {
		return nil
	}
	stmt.Names = make(map[string]int)
	for !p.currTokenIs(lexer.RBRACE) {
		vt, ident, typeOk, identOk := p.parseTypedIdentifier()
		if !typeOk || !identOk {
			pos := stmt.Position()
			pos.CopyEnd(&p.currToken.Position)
			if !typeOk {
				p.appendError(pos, "expected type in struct definition")
			}
			if !identOk {
				p.appendError(pos, "expected identifer after typein struct definition")
			}
			return nil
		}
		stmt.Types = append(stmt.Types, vt)
		stmt.Names[ident.Value] = len(stmt.Types) - 1
	}
	stmt.Position().CopyEnd(&p.currToken.Position)
	p.NextToken()

	return stmt
}
func (p *Parser) parseImportStatement() Statement {
	stmt := &ImportStatement{Token: p.currToken, position: util.Position{
		StartLine: p.currToken.Position.StartLine,
		StartCol:  p.currToken.Position.StartCol,
	}}
	p.NextToken()
	if !p.currTokenIs(lexer.STRING) {
		return nil
	}
	stmt.Path = p.currToken.Literal
	stmt.position.EndLine = p.currToken.Position.EndLine
	stmt.position.EndCol = p.currToken.Position.EndCol
	p.NextToken()

	return stmt
}

func (p *Parser) parseFunctionStatement() Statement {
	stmt := &FunctionStatement{Token: p.currToken, position: util.Position{
		StartLine: p.currToken.Position.StartLine,
		StartCol:  p.currToken.Position.EndCol,
	}}
	p.NextToken()
	if !p.currTokenIs(lexer.IDENTIFIER) {
		p.appendError(&p.currToken.Position, "expected identifier after fnc keyword")
		return nil
	}
	stmt.Name = &IdentifierExpression{Token: p.currToken, Value: p.currToken.Literal}
	p.NextToken()
	if !p.expectCurr(lexer.LPAREN) {
		return nil
	}

	stmt.Params = []FunctionParameter{}

	// for empty arg list if it is rparen then it just stops immediately since we curr are on lparen
	for !p.currTokenIs(lexer.RPAREN) {
		paramType, paramName, typeOk, identOk := p.parseTypedIdentifier()
		if !typeOk || !identOk {
			stmt.Position().CopyEnd(&p.currToken.Position)
			if !typeOk {
				p.appendError(&p.currToken.Position, "expected type in function definition paramaters")
			}
			if !identOk {
				p.appendError(&p.currToken.Position, "expected identifier after type in function definition paramaters")
			}
			return nil
		}
		param := FunctionParameter{
			Type: paramType,
			Name: paramName,
		}

		stmt.Params = append(stmt.Params, param)
		if p.currTokenIs(lexer.RPAREN) {
			break
		} else if p.currTokenIs(lexer.COMMA) {
			p.NextToken()
			continue
		} else {
			return nil
		}
	}
	p.NextToken()

	if !p.expectCurr(lexer.ARROW) {
		return nil
	}
	retType, ok := p.parseType()
	if !ok {
		stmt.Position().CopyEnd(&p.currToken.Position)
		p.appendError(stmt.Position(), "expected return type after arrow in function decl")
		return nil
	}
	stmt.Type = retType

	if !p.expectCurr(lexer.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	currPos := p.currToken.Position
	if !p.expectCurr(lexer.RBRACE) {
		return nil
	}
	stmt.Position().CopyEnd(&currPos)

	return stmt
}

func (p *Parser) parseVarStatement() *DefStatement {
	stmt := &DefStatement{Token: p.currToken}
	if p.currTokenIs(lexer.GLOBAL) {
		stmt.Global = true
	}
	p.NextToken()
	if p.currTokenIs(lexer.CONST) {
		if stmt.Global {
			stmt.Constant = true
			p.NextToken()
		} else {
			p.appendError(&p.currToken.Position, "const keyword can only be used with global variables")
			return nil
		}
	}
	vt, name, typeOk, identOk := p.parseTypedIdentifier()
	if !typeOk {
		p.appendError(&p.currToken.Position, "expected type in def statement")
		return nil
	}
	if !identOk {
		p.appendError(&p.currToken.Position, "expected identifier after type in def statement")
		return nil
	}
	stmt.Type = vt
	stmt.Name = name

	if !p.expectCurr(lexer.ASSIGN) {
		return nil
	}

	stmt.Right = p.parseExpression(LOWEST)

	return stmt
}
