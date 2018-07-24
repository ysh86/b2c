package token

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	STRING  = "STRING"  // "foo"
	IDENT   = "IDENT"   // add, foobar, x, y, ...
	NUM     = "NUM"     // 1343456 or 0.987
	LINENO  = "LINENO"  // line number

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
	COMMA = ","
	COLON = ":"

	LPAREN = "("
	RPAREN = ")"

	// Keywords
	DIM      = "DIM"
	IF       = "IF"
	THEN     = "THEN"
	ELSE     = "ELSE"
	ON       = "ON"
	GOTO     = "GOTO"
	GOSUB    = "GOSUB"
	RETURN   = "RETURN"
	FOR      = "FOR"
	TO       = "TO"
	STEP     = "STEP"
	NEXT     = "NEXT"
	DATA     = "DATA"
	REM      = "REM"
	AND      = "AND"
	OR       = "OR"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
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
