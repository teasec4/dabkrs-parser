package main

import (
	"fmt"
	"strings"
	"parser/internal/parser"
)

func main() {
	data, _ := parser.ReadDSL("dabkrs/dabkrs_1.dsl")
	
	// Split by lines and find 北京 as a standalone header
	lines := strings.Split(data, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "北京" {
			fmt.Printf("=== 北京 at line %d ===\n", i)
			for j := i; j < i+10 && j < len(lines); j++ {
				fmt.Printf("  %d: %q\n", j, lines[j])
			}
			fmt.Println()
		}
	}
}
