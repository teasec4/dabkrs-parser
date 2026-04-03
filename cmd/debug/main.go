package main

import (
	"fmt"
	"parser/internal/parser"
)

func main() {
	dsl := "#NAME \"Test\"\n北京 beijing\n[m1]столица\n上海 shanghai\n[m1]город\n天津 tianjin\n[m1]город\n"
	
	tokens := parser.Lex(dsl)
	fmt.Println("=== Tokens ===")
	for i, t := range tokens {
		fmt.Printf("%d: %s %q\n", i, t.Type, t.Value)
	}
	
	fmt.Println("\n=== AST ===")
	ast := parser.Parse(tokens)
	parser.PrintAST(ast, 0)
	
	fmt.Println("\n=== Entries ===")
	entries, _ := parser.ParseDSL(dsl)
	fmt.Printf("Count: %d\n", len(entries))
	for i, e := range entries {
		fmt.Printf("%d: %s [%s] - Meanings: %d\n", i, e.Hanzi, e.Pinyin, len(e.Meanings))
		for j, m := range e.Meanings {
			fmt.Printf("   Meaning %d: %q\n", j, m.Text)
		}
	}
}
