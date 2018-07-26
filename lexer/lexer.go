package lexer

import "monkey/token"

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	isNewLine := l.skipWhitespace()
	isNewLine = (isNewLine || l.position == 0)

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
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readData() string {
	for isSpace(l.ch) {
		l.readChar()
	}
	position := l.position // TODO: 数値か文字列("" は省略可)
	for l.ch != ':' && !isCRLF(l.ch) && l.ch != 0 { // TODO: , でつなげられる
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position
	for l.ch != '"' && !isCRLF(l.ch) && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '$' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readInteger() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return l.input[position:l.position]
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
