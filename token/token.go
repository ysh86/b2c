package token

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

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	STRING = "STRING" // "foo"
	IDENT  = "IDENT"  // add, foobar, x, y, ...
	NUM    = "NUM"    // 1343456 or 0.987
	LINENO = "LINENO" // line number

	// Operators
	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"

	LT     = "<"
	GT     = ">"
	EQ     = "="
	NOT_EQ = "<>"

	// Delimiters
	SEMICOLON = ";"
	COMMA     = ","
	COLON     = ":"

	LPAREN = "("
	RPAREN = ")"

	// Special keywords
	LET  = "LET"
	CALL = "CALL"
	// Prefix
	LEN   = "LEN"
	ASC   = "ASC"
	CHR_D = "CHR$"
	// Keywords
	DIM    = "DIM"
	IF     = "IF"
	THEN   = "THEN"
	ELSE   = "ELSE"
	ON     = "ON"
	GOTO   = "GOTO"
	GOSUB  = "GOSUB"
	RETURN = "RETURN"
	FOR    = "FOR"
	TO     = "TO"
	STEP   = "STEP"
	NEXT   = "NEXT"
	DATA   = "DATA"
	REM    = "REM"
	AND    = "AND"
	OR     = "OR"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"LEN":    LEN,
	"ASC":    ASC,
	"CHR$":   CHR_D,
	"DIM":    DIM,
	"IF":     IF,
	"THEN":   THEN,
	"ELSE":   ELSE,
	"ON":     ON,
	"GOTO":   GOTO,
	"GOSUB":  GOSUB,
	"RETURN": RETURN,
	"FOR":    FOR,
	"TO":     TO,
	"STEP":   STEP,
	"NEXT":   NEXT,
	"DATA":   DATA,
	"REM":    REM,
	"AND":    AND,
	"OR":     OR,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
