package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"parser/internal/parser"
	"parser/internal/storage"
)

var (
	dbPath    string
	dslfiles  string
	limit     int
	searchStr string
	byPinyin  bool
)

func main() {
	flag.StringVar(&dbPath, "db", "dictionary.db", "path to SQLite database")
	flag.StringVar(&dslfiles, "import", "", "comma-separated list of DSL files to import")
	flag.IntVar(&limit, "limit", 0, "limit number of entries to import (0 = all)")
	flag.StringVar(&searchStr, "search", "", "search by hanzi or pinyin prefix")
	flag.BoolVar(&byPinyin, "pinyin", false, "search by pinyin instead of hanzi")
	flag.Parse()

	db, err := storage.NewDB(dbPath)
	if err != nil {
		log.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	if dslfiles != "" {
		importFiles(db, dslfiles)
	}

	if searchStr != "" {
		search(db, searchStr)
	}

	if dslfiles == "" && searchStr == "" {
		stats(db)
	}
}

func importFiles(db *storage.DB, files string) {
	fileList := splitComma(files)
	total := 0

	for _, file := range fileList {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", file)
			continue
		}

		fmt.Printf("Parsing %s...\n", file)
		entries, err := parser.ParseFile(file, limit)
		if err != nil {
			log.Printf("ParseFile %s: %v", file, err)
			continue
		}

		fmt.Printf("Inserting %d entries...\n", len(entries))
		inserted, err := db.InsertEntries(entries, 1000)
		if err != nil {
			log.Printf("InsertEntries %s: %v", file, err)
			continue
		}
		fmt.Printf("Inserted %d entries from %s\n", inserted, file)
		total += inserted
	}

	fmt.Printf("\nResolving refs...\n")
	if err := db.ResolveRefs(); err != nil {
		log.Printf("ResolveRefs: %v", err)
	}

	count, _ := db.Count()
	fmt.Printf("\nTotal: %d entries in database\n", count)
}

func search(db *storage.DB, query string) {
	var entries []parser.Entry
	var err error

	if byPinyin {
		entries, err = db.SearchByPinyin(query, 20)
	} else {
		entries, err = db.Search(query, 20)
	}

	if err != nil {
		log.Fatalf("Search: %v", err)
	}

	if len(entries) == 0 {
		fmt.Println("No results found")
		return
	}

	fmt.Printf("Found %d results:\n\n", len(entries))
	for _, e := range entries {
		fmt.Printf("%s [%s]\n", e.Hanzi, e.Pinyin)
	}
}

func stats(db *storage.DB) {
	count, err := db.Count()
	if err != nil {
		log.Fatalf("Count: %v", err)
	}
	fmt.Printf("Database: %s\n", dbPath)
	fmt.Printf("Total entries: %d\n", count)
}

func splitComma(s string) []string {
	var result []string
	var current string
	for _, c := range s {
		if c == ',' {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
