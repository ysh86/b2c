package parser

/*
It is based on the code from the "Writing An Interpreter In Go" book.

Copyright (c) 2016-2017 Thorsten Ball

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

import (
	"fmt"
	"io"
	"strconv"

	"github.com/ysh86/b2c/ast"
	"github.com/ysh86/b2c/lexer"
	"github.com/ysh86/b2c/token"
)

const (
	_ int = iota
	LOWEST
	LOGICOR     // OR
	LOGICAND    // AND
	EQUALS      // = (NOT assignment)
	LESSGREATER // > or <
	SUM         // + or -
	PRODUCT     // / or *
	PREFIX      // -X, LEN etc.
	CALL        // myFunction(X) or (group)
)

var precedences = map[token.TokenType]int{
	token.OR:       LOGICOR,
	token.AND:      LOGICAND,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn

	dimVars map[string]*token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.NUM, p.parseIntegerLiteral) // TODO: float
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LEN, p.parsePrefixExpression)
	p.registerPrefix(token.ASC, p.parsePrefixExpression)
	p.registerPrefix(token.CHR_D, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)

	p.registerInfix(token.LPAREN, p.parseCallExpression)

	p.dimVars = make(map[string]*token.Token)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) PrintErrors(w io.Writer) {
	io.WriteString(w, "// ERR: ========== parser ==========\n")
	for _, msg := range p.errors {
		io.WriteString(w, "//  "+msg+"\n")
	}
	io.WriteString(w, "\n")
	p.errors = nil // clear messages
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// ------------------------------------------------------------
// Program
// ------------------------------------------------------------

func (p *Parser) ParseProgram(callback func(s string, isErrors bool)) *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		isErrors := len(p.errors) > 0
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
			callback(stmt.String(), isErrors)
		} else {
			callback("", isErrors)
		}
		p.nextToken()
	}

	return program
}

// ------------------------------------------------------------
// Statements
// ------------------------------------------------------------

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.REM:
		if p.peekTokenIs(token.COLON) {
			p.nextToken()
		}
		return nil

	case token.LINENO:
		if s := p.parseLineNoStatement(); s != nil {
			return s
		}
		return nil
	case token.ASTERISK:
		if s := p.parseLabelStatement(); s != nil {
			return s
		}
		return nil
	case token.DIM:
		if s := p.parseDimStatement(); s != nil {
			return s
		}
		return nil
	case token.IF:
		if s := p.parseIfStatement(); s != nil {
			return s
		}
		return nil
	case token.ON:
		if s := p.parseOnStatement(); s != nil {
			return s
		}
		return nil
	case token.GOTO:
		if s := p.parseGotoStatement(); s != nil {
			return s
		}
		return nil
	case token.GOSUB:
		if s := p.parseGosubStatement(); s != nil {
			return s
		}
		return nil
	case token.RETURN:
		if s := p.parseReturnStatement(); s != nil {
			return s
		}
		return nil
	case token.FOR:
		if s := p.parseForStatement(); s != nil {
			return s
		}
		return nil
	case token.NEXT:
		msg := fmt.Sprintf("invalid statement: %s", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		if p.peekTokenIs(token.COLON) {
			p.nextToken()
		}
		return nil
	case token.IDENT:
		if p.peekToken.Type == token.EQ {
			if s := p.parseLetStatement(); s != nil {
				return s
			}
		} else if _, ok := p.dimVars[p.curToken.Literal]; ok {
			if s := p.parseLetArrayStatement(); s != nil {
				return s
			}
		} else {
			if s := p.parseCallStatement(); s != nil {
				return s
			}
		}
		return nil
	default:
		msg := fmt.Sprintf("got invalid token: %s", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
}

func (p *Parser) parseLineNoStatement() *ast.LineNoStatement {
	stmt := &ast.LineNoStatement{Token: p.curToken}

	l := p.curToken.Literal
	t := token.Token{Type: token.IDENT, Literal: l}
	stmt.Name = &ast.Identifier{Token: t, Value: l}

	if p.peekTokenIs(token.DATA) {
		p.nextToken()

		if s := p.parseDataStatement(); s != nil {
			s.Name = &ast.Identifier{Token: t, Value: l}
			stmt.Data = s
		} else {
			return nil
		}

		if p.peekTokenIs(token.COLON) {
			p.nextToken()
		}
	}

	return stmt
}

func (p *Parser) parseLabelStatement() *ast.LabelStatement {
	stmt := &ast.LabelStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseDimStatement() *ast.DimStatement {
	stmt := &ast.DimStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	params := p.parseDimParameters()
	if params == nil {
		return nil
	}

	stmt.Names = append(stmt.Names, ident)
	stmt.Values = append(stmt.Values, params)
	p.dimVars[ident.Value] = &ident.Token

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()

		if !p.expectPeek(token.IDENT) {
			return nil
		}

		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(token.LPAREN) {
			return nil
		}

		params := p.parseDimParameters()
		if params == nil {
			return nil
		}

		stmt.Names = append(stmt.Names, ident)
		stmt.Values = append(stmt.Values, params)
		p.dimVars[ident.Value] = &ident.Token
	}

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseDimParameters() []*ast.IntegerLiteral {
	integers := []*ast.IntegerLiteral{}

	if !p.expectPeek(token.NUM) {
		return nil
	}

	if i, ok := p.parseIntegerLiteral().(*ast.IntegerLiteral); ok {
		integers = append(integers, i)
	} else {
		return nil
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()

		if !p.expectPeek(token.NUM) {
			return nil
		}

		if i, ok := p.parseIntegerLiteral().(*ast.IntegerLiteral); ok {
			integers = append(integers, i)
		} else {
			// TODO: error message
			return nil
		}
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return integers
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	p.nextToken()

	cond := p.parseExpression(LOWEST)
	if cond == nil {
		return nil
	}

	stmt.Condition = cond

	if !p.expectPeek(token.THEN) {
		return nil
	}

	if p.peekTokenIs(token.ASTERISK) || p.peekTokenIs(token.NUM) {
		// overwrite the 'THEN' token
		p.curToken.Type = token.GOTO
		p.curToken.Literal = token.GOTO
	} else {
		p.nextToken()
	}

	stmts := p.parseStatements(token.ELSE, true)
	if stmts == nil {
		// TODO: error message
		return nil
	}

	stmt.Consequence = stmts

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if p.peekTokenIs(token.ASTERISK) || p.peekTokenIs(token.NUM) {
			// overwrite the 'ELSE' token
			p.curToken.Type = token.GOTO
			p.curToken.Literal = token.GOTO
		} else {
			p.nextToken()
		}

		stmts := p.parseStatements(token.LINENO, true)
		if stmts == nil {
			// TODO: error message
			return nil
		}

		stmt.Alternative = stmts
	}

	return stmt
}

func (p *Parser) parseOnStatement() *ast.OnStatement {
	stmt := &ast.OnStatement{Token: p.curToken}

	p.nextToken()

	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}

	stmt.Value = value // TODO: 整数限定

	if p.peekTokenIs(token.GOTO) {
		if !p.expectPeek(token.GOTO) {
			return nil
		}
	} else {
		if !p.expectPeek(token.GOSUB) {
			return nil
		}
	}

	stmt.Instruction = p.curToken

	idents := p.parseGotoIdentifiers()
	if idents == nil {
		return nil
	}

	stmt.Names = idents

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseGotoIdentifiers() []*ast.Identifier {
	idents := []*ast.Identifier{}

	i := p.parseGotoIdentifier()
	if i == nil {
		return nil
	}

	idents = append(idents, i)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()

		i := p.parseGotoIdentifier()
		if i == nil {
			return nil
		}

		idents = append(idents, i)
	}

	return idents
}

func (p *Parser) parseGotoStatement() *ast.GotoStatement {
	stmt := &ast.GotoStatement{Token: p.curToken}

	i := p.parseGotoIdentifier()
	if i == nil {
		return nil
	}
	stmt.Name = i

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseGosubStatement() *ast.GosubStatement {
	stmt := &ast.GosubStatement{Token: p.curToken}

	i := p.parseGotoIdentifier()
	if i == nil {
		return nil
	}
	stmt.Name = i

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseGotoIdentifier() *ast.Identifier {
	if p.peekTokenIs(token.NUM) {
		p.nextToken()

		// TODO: 整数限定
		t := token.Token{Type: token.IDENT, Literal: p.curToken.Literal}
		return &ast.Identifier{Token: t, Value: p.curToken.Literal}
	} else if p.peekTokenIs(token.ASTERISK) {
		p.nextToken()

		if !p.expectPeek(token.IDENT) {
			return nil
		}

		return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	p.peekError(token.NUM)
	return nil
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.EQ) {
		return nil
	}

	p.nextToken()

	b := p.parseExpression(LOWEST)
	if b == nil {
		return nil
	}

	stmt.Begin = b

	if !p.expectPeek(token.TO) {
		return nil
	}

	p.nextToken()

	e := p.parseExpression(LOWEST)
	if e == nil {
		return nil
	}

	stmt.End = e

	if p.peekTokenIs(token.STEP) {
		p.nextToken()
		p.nextToken()

		s := p.parseExpression(LOWEST)
		if s == nil {
			return nil
		}

		stmt.Step = s
	} else {
		t := token.Token{Type: token.NUM, Literal: "1"}
		stmt.Step = &ast.IntegerLiteral{Token: t, Value: int64(1)}
	}

	if !p.peekTokenIs(token.COLON) && !p.peekTokenIs(token.LINENO) {
		// TODO: error message
		return nil
	} else if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	p.nextToken()

	stmts := p.parseStatements(token.NEXT, false)
	if stmts == nil {
		// TODO: error message
		return nil
	}

	stmt.Statements = stmts

	if !p.expectPeek(token.NEXT) {
		return nil
	}

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseDataStatement() *ast.DataStatement {
	stmt := &ast.DataStatement{Token: p.curToken, Value: p.curToken.Literal}

	return stmt
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	t := token.Token{Type: token.LET, Literal: token.LET}
	stmt := &ast.LetStatement{Token: t}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.EQ) {
		return nil
	}

	p.nextToken()

	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}

	stmt.Value = value

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseLetArrayStatement() *ast.LetStatement {
	t := token.Token{Type: token.LET, Literal: token.LET}
	stmt := &ast.LetStatement{Token: t}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	indices := p.parseIndices()
	if indices == nil {
		return nil
	}

	stmt.Name.Indices = indices

	if !p.expectPeek(token.EQ) {
		return nil
	}

	p.nextToken()

	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}

	stmt.Value = value

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseIndices() []ast.Expression {
	indices := []ast.Expression{}

	p.nextToken()

	i := p.parseExpression(LOWEST)
	if i == nil {
		return nil
	}

	indices = append(indices, i)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		i := p.parseExpression(LOWEST)
		if i == nil {
			return nil
		}

		indices = append(indices, i)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return indices
}

func (p *Parser) parseStatements(stopToken token.TokenType, stopByLine bool) []ast.Statement {
	statements := []ast.Statement{}

	if p.curTokenIs(stopToken) || (stopByLine && p.curTokenIs(token.LINENO)) || p.curTokenIs(token.EOF) {
		// TODO: error message
		return statements
	}

	stmt := p.parseStatement()
	if stmt != nil {
		statements = append(statements, stmt)
	} else {
		// ignore errors because the 'REM' statement returns nil
	}

	for !p.peekTokenIs(stopToken) && !(stopByLine && p.peekTokenIs(token.LINENO)) && !p.peekTokenIs(token.EOF) {
		p.nextToken()

		stmt := p.parseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		} else {
			// ignore errors because the 'REM' statement returns nil
		}
	}

	return statements
}

func (p *Parser) parseCallStatement() *ast.CallStatement {
	t := token.Token{Type: token.CALL, Literal: token.CALL}
	stmt := &ast.CallStatement{Token: t}

	f := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	var exp *ast.CallExpression
	if p.peekTokenIs(token.COLON) || p.peekTokenIs(token.LINENO) || p.peekTokenIs(token.EOF) {
		// no args
		exp = &ast.CallExpression{Token: t, Function: f}
	} else {
		p.nextToken()

		e := p.parseCallExpression(f)
		if ce, ok := e.(*ast.CallExpression); !ok || ce == nil {
			return nil
		} else {
			exp = ce
		}
	}

	stmt.Expression = exp

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

// ------------------------------------------------------------
// Expressions
// ------------------------------------------------------------

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.COLON) && !p.peekTokenIs(token.LINENO) && !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) parseIdentifier() ast.Expression {
	var ident ast.Expression

	if _, ok := p.dimVars[p.curToken.Literal]; ok {
		name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if p.peekTokenIs(token.LPAREN) { // TODO: () なしはスカラになる
			if !p.expectPeek(token.LPAREN) {
				return nil
			}

			indices := p.parseIndices()
			if indices == nil {
				return nil
			}

			name.Indices = indices
		}

		ident = name
	} else if p.peekTokenIs(token.IDENT) || p.peekTokenIs(token.NUM) /*|| p.peekTokenIs(token.MINUS)*/ {
		f := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		p.nextToken()

		exp := p.parseCallExpression(f)
		if exp == nil {
			return nil
		}

		ident = exp
	} else {
		ident = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	return ident
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	f, ok := function.(*ast.Identifier)
	if !ok {
		msg := fmt.Sprintf("could not parse %q as identifier", function.TokenLiteral())
		p.errors = append(p.errors, msg)
		return nil
	}

	t := token.Token{Type: token.CALL, Literal: token.CALL}
	exp := &ast.CallExpression{Token: t, Function: f}

	args := p.parseCallArguments()
	if args == nil {
		return nil
	}
	exp.Arguments = args

	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	requireR := false
	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		requireR = true
	}

	arg := p.parseExpression(LOWEST)
	if arg == nil {
		return nil
	}
	args = append(args, arg)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		arg := p.parseExpression(LOWEST)
		if arg == nil {
			return nil
		}
		args = append(args, arg)
	}

	if requireR {
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	return args
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
