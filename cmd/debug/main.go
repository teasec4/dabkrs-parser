package main

import (
	"fmt"
	"parser/internal/parser"
)

func main() {
	total := 0
	for i := 1; i <= 3; i++ {
		entries, _ := parser.ParseFile(fmt.Sprintf("dabkrs/dabkrs_%d.dsl", i), 0)
		fmt.Printf("dabkrs_%d.dsl: %d entries\n", i, len(entries))
		total += len(entries)
	}
	fmt.Printf("Total: %d entries\n", total)
}
