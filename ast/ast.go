package ast

import (
	"bytes"
	"monkey/token"
	"strconv"
	"strings"
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
	/*
		params := []string{}
		for _, p := range fl.Parameters {
			params = append(params, p.String())
		}

		out.WriteString(fl.TokenLiteral())
		out.WriteString("(")
		out.WriteString(strings.Join(params, ", "))
		out.WriteString(") ")
		out.WriteString(fl.Body.String())*/

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

type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// Expressions
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. !
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

	out.WriteString("(")
	out.WriteString(oe.Left.String())
	out.WriteString(" " + oe.Operator + " ")
	out.WriteString(oe.Right.String())
	out.WriteString(")")

	return out.String()
}

type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
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
