package parser

type TokenType int

const(
	TokenText TokenType = iota
	TokenOpenTag // m1
	TokenCloseTag // m2
)

type Token struct{
	Type TokenType
	Value string
}