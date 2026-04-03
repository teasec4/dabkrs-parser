package main

import (
	"fmt"
	"parser/internal/parser"
	"strings"
)

func isPinyin(s string) bool {
	if len(s) == 0 { return false }
	hasLetter := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' { hasLetter = true; continue }
		if r == '\'' || r == ' ' || r == 'ō' || r == 'ó' || r == 'ě' || r == 'è' ||
			r == 'ā' || r == 'á' || r == 'ǎ' || r == 'à' ||
			r == 'ē' || r == 'é' || r == 'ě' || r == 'è' ||
			r == 'ī' || r == 'í' || r == 'ǐ' || r == 'ì' ||
			r == 'ū' || r == 'ú' || r == 'ǔ' || r == 'ù' ||
			r == 'ǖ' || r == 'ǘ' || r == 'ǚ' || r == 'ǜ' { continue }
		return false
	}
	return hasLetter
}

func main() {
	data, _ := parser.ReadDSL("./dabkrs/dabkrs_1.dsl")
	tokens := parser.Lex(data)
	ast := parser.Parse(tokens)
	
	node := ast.Children[12]
	lines := strings.Split(node.Value, "\n")
	
	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed == "" { continue }
		
		// Check each char
		for i, r := range trimmed {
			valid := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || 
			         r == '\'' || r == ' ' ||
			         r == 'ō' || r == 'ó' || r == 'ě' || r == 'è' ||
			         r == 'ā' || r == 'á' || r == 'ǎ' || r == 'à' ||
			         r == 'ē' || r == 'é' || r == 'ě' || r == 'è' ||
			         r == 'ī' || r == 'í' || r == 'ǐ' || r == 'ì' ||
			         r == 'ū' || r == 'ú' || r == 'ǔ' || r == 'ù' ||
			         r == 'ǖ' || r == 'ǘ' || r == 'ǚ' || r == 'ǜ'
			if !valid {
				fmt.Printf("Line %q: Char %d: %c (U+%04X) - INVALID\n", trimmed, i, r, r)
			}
		}
		fmt.Printf("Line %q: isPinyin=%v\n", trimmed, isPinyin(trimmed))
	}
}
