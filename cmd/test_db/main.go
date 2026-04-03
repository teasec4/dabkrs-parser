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

	entries, err := parser.ParseFile("dabkrs/dabkrs_1.dsl", 1000)
	if err != nil {
		log.Fatalf("ParseFile: %v", err)
	}

	fmt.Printf("Parsed %d entries\n", len(entries))
	for i, e := range entries {
		fmt.Printf("  %d: %s [%s] - %d meanings\n", i, e.Hanzi, e.Pinyin, len(e.Meanings))
		for j, m := range e.Meanings {
			if j >= 3 {
				fmt.Printf("    ... and %d more meanings\n", len(e.Meanings)-3)
				break
			}
			fmt.Printf("    Meaning %d: %s\n", j, m.Text)
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

	// Resolve refs
	if err := db.ResolveRefs(); err != nil {
		log.Fatalf("ResolveRefs: %v", err)
	}
	fmt.Println("Refs resolved")

	entry, err := db.GetEntryByHanzi("上海市")
	if err != nil {
		log.Fatalf("GetEntryByHanzi: %v", err)
	}
	if entry != nil {
		fmt.Printf("\nLookup '北京':\n")
		fmt.Printf("  Hanzi: %s\n", entry.Hanzi)
		fmt.Printf("  Pinyin: %s\n", entry.Pinyin)
		for i, m := range entry.Meanings {
			fmt.Printf("  Meaning %d: %s\n", i, m.Text)
		}
	}

	fmt.Println("\nDone!")
}
