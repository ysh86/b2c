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
	"bytes"
	"testing"

	"github.com/ysh86/b2c/token"
)

func TestNextToken(t *testing.T) {
	input := `0
 ! 'this is illegal
REM this is test
 1 STR="stringggg" 'this is illegal
2 FIVE = 5:ten = 10
3 STR2="strin
4
10 CLEAR : RANDOMIZE :DIM ADD(10,15),A$(3)
15 LOCATE 2,3
20 result = ADD(five, ten)
30 +-/*5
40 5 < 10 > 5.2
60 10 = 10
70 10 <> 9
80 PRINT "foo";A$
90 *GOGO:A AND D:O OR R:RETURN :
100 IF 5<10 THEN 30 ELSE GOSUB *GOGO
200 ON N GOTO 10
300 FOR I=10 TO 1 STEP -1: NEXT
1000 DATA ThisIsData:' this is data
1100 DATA "001122334455"
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LINENO, "0"},
		{token.ILLEGAL, "!"},
		{token.REM, "this is illegal"},
		{token.REM, "this is test"},
		{token.NUM, "1"},
		{token.IDENT, "STR"},
		{token.EQ, "="},
		{token.STRING, "stringggg"},
		{token.REM, "this is illegal"},
		{token.LINENO, "2"},
		{token.IDENT, "FIVE"},
		{token.EQ, "="},
		{token.NUM, "5"},
		{token.COLON, ":"},
		{token.IDENT, "ten"},
		{token.EQ, "="},
		{token.NUM, "10"},
		{token.LINENO, "3"},
		{token.IDENT, "STR2"},
		{token.EQ, "="},
		{token.ILLEGAL, "\n"},
		{token.NUM, "4"},
		{token.LINENO, "10"},
		{token.IDENT, "CLEAR"},
		{token.COLON, ":"},
		{token.IDENT, "RANDOMIZE"},
		{token.COLON, ":"},
		{token.DIM, "DIM"},
		{token.IDENT, "ADD"},
		{token.LPAREN, "("},
		{token.NUM, "10"},
		{token.COMMA, ","},
		{token.NUM, "15"},
		{token.RPAREN, ")"},
		{token.COMMA, ","},
		{token.IDENT, "A$"},
		{token.LPAREN, "("},
		{token.NUM, "3"},
		{token.RPAREN, ")"},
		{token.LINENO, "15"},
		{token.IDENT, "LOCATE"},
		{token.NUM, "2"},
		{token.COMMA, ","},
		{token.NUM, "3"},
		{token.LINENO, "20"},
		{token.IDENT, "result"},
		{token.EQ, "="},
		{token.IDENT, "ADD"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.LINENO, "30"},
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.NUM, "5"},
		{token.LINENO, "40"},
		{token.NUM, "5"},
		{token.LT, "<"},
		{token.NUM, "10"},
		{token.GT, ">"},
		{token.NUM, "5.2"},
		{token.LINENO, "60"},
		{token.NUM, "10"},
		{token.EQ, "="},
		{token.NUM, "10"},
		{token.LINENO, "70"},
		{token.NUM, "10"},
		{token.NOT_EQ, "<>"},
		{token.NUM, "9"},
		{token.LINENO, "80"},
		{token.IDENT, "PRINT"},
		{token.STRING, "foo"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "A$"},
		{token.LINENO, "90"},
		{token.ASTERISK, "*"},
		{token.IDENT, "GOGO"},
		{token.COLON, ":"},
		{token.IDENT, "A"},
		{token.AND, "AND"},
		{token.IDENT, "D"},
		{token.COLON, ":"},
		{token.IDENT, "O"},
		{token.OR, "OR"},
		{token.IDENT, "R"},
		{token.COLON, ":"},
		{token.RETURN, "RETURN"},
		{token.COLON, ":"},
		{token.LINENO, "100"},
		{token.IF, "IF"},
		{token.NUM, "5"},
		{token.LT, "<"},
		{token.NUM, "10"},
		{token.THEN, "THEN"},
		{token.NUM, "30"},
		{token.ELSE, "ELSE"},
		{token.GOSUB, "GOSUB"},
		{token.ASTERISK, "*"},
		{token.IDENT, "GOGO"},
		{token.LINENO, "200"},
		{token.ON, "ON"},
		{token.IDENT, "N"},
		{token.GOTO, "GOTO"},
		{token.NUM, "10"},
		{token.LINENO, "300"},
		{token.FOR, "FOR"},
		{token.IDENT, "I"},
		{token.EQ, "="},
		{token.NUM, "10"},
		{token.TO, "TO"},
		{token.NUM, "1"},
		{token.STEP, "STEP"},
		{token.MINUS, "-"},
		{token.NUM, "1"},
		{token.COLON, ":"},
		{token.NEXT, "NEXT"},
		{token.LINENO, "1000"},
		{token.DATA, "ThisIsData"},
		{token.COLON, ":"},
		{token.REM, "this is data"},
		{token.LINENO, "1100"},
		{token.DATA, "\"001122334455\""},
		{token.EOF, ""},
	}

	l := New(bytes.NewBufferString(input))

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
