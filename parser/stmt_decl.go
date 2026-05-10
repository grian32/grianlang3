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
			pos := stmt.Position()
			pos.CopyEnd(&p.currToken.Position)
			p.appendError(pos, "expected type in struct definition")
			return nil
		}
		p.getPointers(&vt)

		if !p.currTokenIs(lexer.IDENTIFIER) {
			pos := stmt.Position()
			pos.CopyEnd(&p.currToken.Position)
			p.appendError(pos, "expected identifier after type in struct definition")
			return nil
		}
		stmt.Types = append(stmt.Types, vt)
		stmt.Names[p.currToken.Literal] = len(stmt.Types) - 1
		p.NextToken()
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
		var paramType lexer.VarType
		if p.currTokenIs(lexer.TYPE) {
			paramType = p.currToken.VarType
		} else if p.currTokenIs(lexer.IDENTIFIER) {
			paramType = lexer.VarType{
				IsStructType: true,
				StructName:   p.currToken.Literal,
			}
		} else {
			return nil
		}
		p.NextToken()
		p.getPointers(&paramType)
		if !p.currTokenIs(lexer.IDENTIFIER) {
			p.appendError(&p.currToken.Position, "expected identifier after type in function definition")
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
	if p.currTokenIs(lexer.TYPE) {
		stmt.Type = p.currToken.VarType
	} else if p.currTokenIs(lexer.IDENTIFIER) {
		stmt.Type = lexer.VarType{
			IsStructType: true,
			StructName:   p.currToken.Literal,
		}
	} else {
		return nil
	}
	p.NextToken()
	p.getPointers(&stmt.Type)

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
	if p.currTokenIs(lexer.TYPE) {
		stmt.Type = p.currToken.VarType
	} else if p.currTokenIs(lexer.IDENTIFIER) {
		stmt.Type = lexer.VarType{
			IsStructType: true,
			StructName:   p.currToken.Literal,
		}
	} else {
		return nil
	}
	p.NextToken()
	p.getPointers(&stmt.Type)

	if !p.currTokenIs(lexer.IDENTIFIER) {
		p.appendError(&p.currToken.Position, "expected identifier after type in def stmt")
		return nil
	}

	stmt.Name = &IdentifierExpression{Token: p.currToken, Value: p.currToken.Literal}
	p.NextToken()

	if !p.expectCurr(lexer.ASSIGN) {
		return nil
	}

	stmt.Right = p.parseExpression(LOWEST)

	return stmt
}
