package main

import (
	"fmt"
	"log"
	"parser/internal/parser"
	"parser/internal/storage"
)

func main() {
	dbPath := "/tmp/dabkrs.db"

	db, err := storage.NewDB(dbPath)
	if err != nil {
		log.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	// Parse DSL file using existing approach
	path := "./dabkrs/dabkrs_1.dsl"
	fmt.Printf("Parsing file: %s\n", path)

	// Read DSL file
	data, err := parser.ReadDSL(path)
	if err != nil {
		log.Fatalf("ReadDSL: %v", err)
	}

	// Tokenize
	tokens := parser.Lex(data)

	// Parse AST
	ast := parser.Parse(tokens)

	// Extract entries
	entries := parser.ExtractEntries(ast, 1000)

	fmt.Printf("Parsed %d entries\n", len(entries))
	for i, e := range entries {
		fmt.Printf("  %d: %s [%s] - %d meanings\n", i, e.Headword, e.Pinyin, len(e.Meanings))
		for j, m := range e.Meanings {
			if j >= 3 {
				fmt.Printf("    ... and %d more meanings\n", len(e.Meanings)-3)
				break
			}
			fmt.Printf("    Meaning %d (level %d): %s\n", j, m.Level, m.Text)
			if len(m.Tags) > 0 {
				fmt.Printf("      Tags: ")
				for k, tag := range m.Tags {
					if k > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s:%s", tag.Type, tag.Value)
				}
				fmt.Println()
			}
		}
	}

	inserted, err := db.InsertEntries(entries, 0)
	if err != nil {
		log.Fatalf("InsertEntries: %v", err)
	}
	fmt.Printf("\nInserted %d entries\n", inserted)

	count, err := db.Count()
	if err != nil {
		log.Fatalf("Count: %v", err)
	}
	fmt.Printf("Total entries in DB: %d\n", count)

	entry, err := db.GetEntryByHeadword("上海市")
	if err != nil {
		log.Fatalf("GetEntryByHeadword: %v", err)
	}
	if entry != nil {
		fmt.Printf("\nLookup '上海市':\n")
		fmt.Printf("  Headword: %s\n", entry.Headword)
		fmt.Printf("  Pinyin: %s\n", entry.Pinyin)
		for i, m := range entry.Meanings {
			fmt.Printf("  Meaning %d (level %d): %s\n", i, m.Level, m.Text)
			if len(m.Tags) > 0 {
				fmt.Printf("    Tags: ")
				for k, tag := range m.Tags {
					if k > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s:%s", tag.Type, tag.Value)
				}
				fmt.Println()
			}
		}
	}

	// Test search
	results, err := db.SearchByHeadword("上", 5)
	if err != nil {
		log.Fatalf("SearchByHeadword: %v", err)
	}
	fmt.Printf("\nSearch results for '上':\n")
	for i, e := range results {
		fmt.Printf("  %d: %s [%s]\n", i, e.Headword, e.Pinyin)
	}

	fmt.Println("\nDone!")
}
