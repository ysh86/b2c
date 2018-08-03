package ast

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
	"strconv"
	"strings"

	"github.com/ysh86/b2c/token"
)

// The base Node interface
type Node interface {
	TokenLiteral() string
	String() string
}

// All statement nodes implement this
type Statement interface {
	Node
	statementNode()
}

// All expression nodes implement this
type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
		out.WriteString("\n")
	}

	return out.String()
}

// Statements
type LineNoStatement struct {
	Token token.Token // the token.LINENO token
	Name  *Identifier // copy from Token.Literal
	Data  *DataStatement
}

func (lns *LineNoStatement) statementNode()       {}
func (lns *LineNoStatement) TokenLiteral() string { return lns.Token.Literal }
func (lns *LineNoStatement) String() string {
	var out bytes.Buffer

	if lns.Data == nil {
		out.WriteString("_" + lns.Name.String() + ":;")
	} else {
		out.WriteString(lns.Data.String())
	}

	return out.String()
}

type LabelStatement struct {
	Token token.Token // the '*' token
	Name  *Identifier
}

func (ls *LabelStatement) statementNode()       {}
func (ls *LabelStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LabelStatement) String() string {
	var out bytes.Buffer

	out.WriteString("\n")
	out.WriteString("// -----------------------------------\n")
	out.WriteString(ls.Name.String() + ":;")

	return out.String()
}

type DimStatement struct {
	Token  token.Token // the token.DIM token
	Names  []*Identifier
	Values [][]*IntegerLiteral
}

func (ds *DimStatement) statementNode()       {}
func (ds *DimStatement) TokenLiteral() string { return ds.Token.Literal }
func (ds *DimStatement) String() string {
	var out bytes.Buffer

	l := len(ds.Names)

	if l > 0 {
		out.WriteString("int ") // TODO: char がいる
		out.WriteString(ds.Names[0].String())
		ll := len(ds.Values[0])
		for j := 0; j < ll; j++ { // TODO: x,y が逆かも
			out.WriteString("[")
			out.WriteString(ds.Values[0][j].String())
			out.WriteString("]")
		}
	}

	for i := 1; i < l; i++ {
		out.WriteString(", ")
		out.WriteString(ds.Names[i].String())
		ll := len(ds.Values[i])
		for j := 0; j < ll; j++ { // TODO: x,y が逆かも
			out.WriteString("[")
			out.WriteString(ds.Values[i][j].String())
			out.WriteString("]")
		}
	}

	out.WriteString(";")

	return out.String()
}

type IfStatement struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence []Statement
	Alternative []Statement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out bytes.Buffer

	out.WriteString("if (")
	out.WriteString(is.Condition.String())
	out.WriteString(") {\n")
	for _, s := range is.Consequence {
		out.WriteString("    ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	if len(is.Alternative) != 0 {
		out.WriteString("} else {\n")
		for _, s := range is.Alternative {
			out.WriteString("    ")
			out.WriteString(s.String())
			out.WriteString("\n")
		}
	}
	out.WriteString("}")

	return out.String()
}

type OnStatement struct {
	Token       token.Token // the token.ON token
	Value       Expression
	Instruction token.Token // 'GOTO' or 'GUSUB'
	Names       []*Identifier
}

func (os *OnStatement) statementNode()       {}
func (os *OnStatement) TokenLiteral() string { return os.Token.Literal }
func (os *OnStatement) String() string {
	var out bytes.Buffer

	out.WriteString("switch (")
	out.WriteString(os.Value.String())
	out.WriteString(") {\n")

	for i, n := range os.Names {
		out.WriteString("case ")
		out.WriteString(strconv.Itoa(i + 1)) // 1 origin
		out.WriteString(":\n")

		var c byte
		if len(n.String()) > 0 {
			c = n.String()[0]
		}

		if os.Instruction.Type == token.GOTO {
			out.WriteString("    goto ")
			if '0' <= c && c <= '9' {
				out.WriteString("_")
			}
			out.WriteString(n.String())
			out.WriteString(";\n")
		} else {
			out.WriteString("    if (setjmp(env) == 0) {\n")
			out.WriteString("        goto ")
			if '0' <= c && c <= '9' {
				out.WriteString("_")
			}
			out.WriteString(n.String())
			out.WriteString(";\n")
			out.WriteString("    }\n")
			out.WriteString("    // return from longjmp()\n")
		}
		out.WriteString("    break;\n")
	}

	out.WriteString("default:\n")
	out.WriteString("    // nothing to do\n")
	out.WriteString("    break;\n")
	out.WriteString("}")

	return out.String()
}

type GotoStatement struct {
	Token token.Token // the token.GOTO token
	Name  *Identifier
}

func (gs *GotoStatement) statementNode()       {}
func (gs *GotoStatement) TokenLiteral() string { return gs.Token.Literal }
func (gs *GotoStatement) String() string {
	var out bytes.Buffer

	var c byte
	if len(gs.Name.String()) > 0 {
		c = gs.Name.String()[0]
	}

	out.WriteString("goto ") // TODO: 飛び先に RETURN があると死ぬ
	if '0' <= c && c <= '9' {
		out.WriteString("_")
	}
	out.WriteString(gs.Name.String())
	out.WriteString(";")

	return out.String()
}

type GosubStatement struct {
	Token token.Token // the token.GOSUB token
	Name  *Identifier
}

func (gss *GosubStatement) statementNode()       {}
func (gss *GosubStatement) TokenLiteral() string { return gss.Token.Literal }
func (gss *GosubStatement) String() string {
	var out bytes.Buffer

	var c byte
	if len(gss.Name.String()) > 0 {
		c = gss.Name.String()[0]
	}

	out.WriteString("if (setjmp(env) == 0) {\n")
	out.WriteString("    goto ")
	if '0' <= c && c <= '9' {
		out.WriteString("_")
	}
	out.WriteString(gss.Name.String())
	out.WriteString(";\n")
	out.WriteString("}\n")
	out.WriteString("// return from longjmp()")

	return out.String()
}

type ReturnStatement struct {
	Token token.Token // the 'return' token
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString("longjmp(env, 1);\n")
	out.WriteString("// -----------------------------------")

	return out.String()
}

type ForStatement struct {
	Token      token.Token // the token.FOR token
	Name       *Identifier
	Begin      Expression
	End        Expression
	Step       Expression
	Statements []Statement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out bytes.Buffer

	out.WriteString("for (")
	out.WriteString("int ") // TODO: float
	out.WriteString(fs.Name.String())
	out.WriteString(" = ")
	out.WriteString(fs.Begin.String())
	out.WriteString("; ")
	out.WriteString(fs.Name.String())
	out.WriteString(" != ")
	out.WriteString(fs.End.String())
	out.WriteString("; ")
	out.WriteString(fs.Name.String())
	out.WriteString(" += ")
	out.WriteString(fs.Step.String())
	out.WriteString(") {\n")
	for _, s := range fs.Statements {
		out.WriteString("    ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	out.WriteString("}")

	return out.String()
}

type DataStatement struct {
	Token token.Token // the token.DATA token
	Name  *Identifier
	Value string
}

func (das *DataStatement) statementNode()       {}
func (das *DataStatement) TokenLiteral() string { return das.Token.Literal }
func (das *DataStatement) String() string {
	var out bytes.Buffer

	// TODO: すでに "0123456789" の時がある
	out.WriteString("char *_" + das.Name.String() + " = \"" + das.Value + "\";")

	return out.String()
}

type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

type CallStatement struct {
	Token      token.Token // the token.CALL token
	Expression *CallExpression
}

func (cs *CallStatement) statementNode()       {}
func (cs *CallStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *CallStatement) String() string {
	var out bytes.Buffer

	if cs.Expression != nil {
		out.WriteString(cs.Expression.String())
		out.WriteString(";")
	}

	return out.String()
}

// Expressions
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
	Indices []Expression
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string {
	var out bytes.Buffer

	out.WriteString(i.Value) // TODO: $ を置換しないと
	l := len(i.Indices)
	for j := 0; j < l; j++ { // TODO: x,y が逆かも
		out.WriteString("[")
		out.WriteString(i.Indices[j].String())
		out.WriteString("]")
	}

	return out.String()
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. -
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (oe *InfixExpression) expressionNode()      {}
func (oe *InfixExpression) TokenLiteral() string { return oe.Token.Literal }
func (oe *InfixExpression) String() string {
	var out bytes.Buffer

	var op string
	switch oe.Operator {
	case token.OR:
		op = "||"
	case token.AND:
		op = "&&"
	case token.EQ:
		op = "=="
	case token.NOT_EQ:
		op = "!="
	default:
		op = oe.Operator
	}

	out.WriteString("(")
	out.WriteString(oe.Left.String())
	out.WriteString(" " + op + " ")
	out.WriteString(oe.Right.String())
	out.WriteString(")")

	return out.String()
}

type CallExpression struct {
	Token     token.Token // The token.CALL token
	Function  *Identifier
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}
