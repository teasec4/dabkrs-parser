package parser

import (
	"fmt"
	"os"
	"testing"
)

func TestShanghai(t *testing.T) {
	wd, _ := os.Getwd()
	entries, _ := ParseFile(wd+"/../../dabkrs/dabkrs_1.dsl", 0)
	
	count := 0
	for _, e := range entries {
		if e.Hanzi == "上海市" {
			count++
			fmt.Printf("#%d: %s [%s]\n", count, e.Hanzi, e.Pinyin)
			fmt.Printf("  Meanings: %d\n", len(e.Meanings))
			for i, m := range e.Meanings {
				fmt.Printf("  [%d] %q refs=%v\n", i, m.Text, m.Refs)
			}
		}
	}
	fmt.Printf("\nTotal 上海市 entries: %d\n", count)
}
