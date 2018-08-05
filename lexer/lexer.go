package lexer

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
	"io"
	"strings"

	"github.com/ysh86/b2c/token"
)

type Lexer struct {
	reader  io.Reader
	isFirst bool // Is it the first token?
	ch      byte // current char under examination
	peekCh  byte // char after current char
}

func New(r io.Reader) *Lexer {
	l := &Lexer{reader: r, isFirst: true}
	l.readChar()
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	isNewLine := l.skipWhitespace()
	isNewLine = (isNewLine || l.isFirst)
	l.isFirst = false

	switch l.ch {
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '=':
		tok = newToken(token.EQ, l.ch)
	case '<':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		tok = newToken(token.GT, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '\'':
		l.readChar()
		tok.Type = token.REM
		tok.Literal = l.readData()
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if l.ch == '"' {
			l.readChar()
			tok.Type = token.STRING
			tok.Literal = l.readString()
			if l.ch != '"' {
				tok = newToken(token.ILLEGAL, l.ch)
			}
			l.readChar()
			return tok
		} else if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			if tok.Type == token.DATA || tok.Type == token.REM {
				tok.Literal = l.readData()
			}
			return tok
		} else if isDigit(l.ch) {
			if isNewLine {
				tok.Type = token.LINENO
				tok.Literal = l.readInteger()
			} else {
				tok.Type = token.NUM
				tok.Literal = l.readNumber()
			}
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() (isNewLine bool) {
	isNewLine = false
	for isSpace(l.ch) || isCRLF(l.ch) {
		if isCRLF(l.ch) {
			isNewLine = true
		} else {
			isNewLine = false
		}
		l.readChar()
	}
	return
}

func (l *Lexer) readChar() {
	var p [1]byte

	_, err := l.reader.Read(p[:])
	if err != nil {
		l.ch = l.peekCh
		l.peekCh = 0
		return
	}

	l.ch = l.peekCh
	l.peekCh = p[0]
}

func (l *Lexer) peekChar() byte {
	return l.peekCh
}

func (l *Lexer) readData() string {
	for isSpace(l.ch) {
		l.readChar()
	}
	// TODO: 数値か文字列("" は省略可)
	var out strings.Builder
	for l.ch != ':' && !isCRLF(l.ch) && l.ch != 0 { // TODO: , でつなげられる
		out.WriteByte(l.ch)
		l.readChar()
	}
	return out.String()
}

func (l *Lexer) readString() string {
	var out strings.Builder
	for l.ch != '"' && !isCRLF(l.ch) && l.ch != 0 {
		out.WriteByte(l.ch)
		l.readChar()
	}
	return out.String()
}

func (l *Lexer) readIdentifier() string {
	var out strings.Builder
	for isLetter(l.ch) || isDigit(l.ch) {
		out.WriteByte(l.ch)
		l.readChar()
	}
	if l.ch == '$' {
		out.WriteByte(l.ch)
		l.readChar()
	}
	return out.String()
}

func (l *Lexer) readInteger() string {
	var out strings.Builder
	for isDigit(l.ch) {
		out.WriteByte(l.ch)
		l.readChar()
	}
	return out.String()
}

func (l *Lexer) readNumber() string {
	var out strings.Builder
	for isDigit(l.ch) || l.ch == '.' {
		out.WriteByte(l.ch)
		l.readChar()
	}
	return out.String()
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t'
}

func isCRLF(ch byte) bool {
	return ch == '\n' || ch == '\r'
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
