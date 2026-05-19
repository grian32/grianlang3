package parser

import (
	"grianlang3/lexer"
	"grianlang3/util"
)

func (p *Parser) parseExternStatement(private bool) Statement {
	currTkn := p.currToken
	p.NextToken()
	if p.currTokenIs(lexer.STRUCT) {
		stmt := &ExternStructStatement{Token: currTkn, position: util.Position{
			StartLine: p.currToken.Position.StartLine,
			StartCol:  p.currToken.Position.StartCol,
		}, Private: private}
		if !p.currTokenIs(lexer.IDENTIFIER) {
			return nil
		}
		stmt.position.CopyEnd(&p.currToken.Position)
		stmt.Name = p.currToken.Literal
		p.NextToken()

		return stmt
	} else if p.currTokenIs(lexer.FNC) {
		stmt := &ExternFunctionStatement{Token: currTkn, position: util.Position{
			StartLine: p.currToken.Position.StartLine,
			StartCol:  p.currToken.Position.StartCol,
		}, Private: private}
		name, params, retType := p.parseFunctionHeader()
		if name == nil || params == nil {
			stmt.Position().CopyEnd(&p.currToken.Position)
			return nil
		}
		stmt.Name = name.Value
		stmt.Params = params
		stmt.ReturnType = retType

		return stmt
	}

	return nil
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
				p.appendError(pos, "expected identifer after type in struct definition")
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

// returns: name ident, []params, return type
func (p *Parser) parseFunctionHeader() (*IdentifierExpression, []FunctionParameter, lexer.VarType) {
	p.NextToken()
	if !p.currTokenIs(lexer.IDENTIFIER) {
		p.appendError(&p.currToken.Position, "expected identifier after fnc keyword")
		return nil, nil, lexer.VarType{}
	}
	name := &IdentifierExpression{Token: p.currToken, Value: p.currToken.Literal}
	p.NextToken()
	if !p.expectCurr(lexer.LPAREN) {
		return nil, nil, lexer.VarType{}
	}

	params := []FunctionParameter{}
	// for empty arg list if it is rparen then it just stops immediately since we curr are on lparen
	for !p.currTokenIs(lexer.RPAREN) {
		paramType, paramName, typeOk, identOk := p.parseTypedIdentifier()
		if !typeOk || !identOk {
			if !typeOk {
				p.appendError(&p.currToken.Position, "expected type in function definition params")
			}
			if !identOk {
				p.appendError(&p.currToken.Position, "expected identifier after type in function definition params")
			}
			return nil, nil, lexer.VarType{}
		}
		param := FunctionParameter{
			Type: paramType,
			Name: paramName,
		}

		params = append(params, param)
		if p.currTokenIs(lexer.RPAREN) {
			break
		} else if p.currTokenIs(lexer.COMMA) {
			p.NextToken()
			continue
		} else {
			return nil, nil, lexer.VarType{}
		}
	}
	p.NextToken()

	if !p.expectCurr(lexer.ARROW) {
		return nil, nil, lexer.VarType{}
	}
	retType, ok := p.parseType()
	if !ok {
		p.appendError(&p.currToken.Position, "expected return type after arrow in function decl")
		return nil, nil, lexer.VarType{}
	}

	return name, params, retType
}

func (p *Parser) parseFunctionStatement(private bool) Statement {
	stmt := &FunctionStatement{Token: p.currToken, position: util.Position{
		StartLine: p.currToken.Position.StartLine,
		StartCol:  p.currToken.Position.EndCol,
	}, Private: private}
	name, params, retType := p.parseFunctionHeader()
	if name == nil || params == nil {
		stmt.Position().CopyEnd(&p.currToken.Position)
		return nil
	}
	stmt.Name = name
	stmt.Params = params
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
