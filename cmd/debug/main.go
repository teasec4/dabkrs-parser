package main

import (
	"parser/internal/parser"
)

func main() {
	// Test с примером из реального DSL
	input := "三比西河\n sānbǐxīhé\n [m1]река Замбези ([i]Южная Африка[/i])[/m]"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)
	
	parser.PrintAST(ast, 0)
}
