package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"parser/internal/parser"
	"parser/internal/storage"
	"strings"
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
	flag.StringVar(&searchStr, "search", "", "search by headword or pinyin prefix")
	flag.BoolVar(&byPinyin, "pinyin", false, "search by pinyin instead of headword")
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
		// Read DSL file
		data, err := parser.ReadDSL(file)
		if err != nil {
			log.Printf("ReadDSL %s: %v", file, err)
			continue
		}

		// Tokenize
		tokens := parser.Lex(data)

		// Parse AST
		ast := parser.Parse(tokens)

		// Extract entries
		entries := parser.ExtractEntries(ast, limit)

		fmt.Printf("Inserting %d entries...\n", len(entries))
		inserted, err := db.InsertEntries(entries, 1000)
		if err != nil {
			log.Printf("InsertEntries %s: %v", file, err)
			continue
		}
		fmt.Printf("Inserted %d entries from %s\n", inserted, file)
		total += inserted
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
		fmt.Printf("%s [%s]\n", e.Headword, e.Pinyin)
		if len(e.Meanings) > 0 {
			for i, m := range e.Meanings {
				if i >= 2 {
					fmt.Printf("  ... and %d more meanings\n", len(e.Meanings)-2)
					break
				}
				fmt.Printf("  %d. %s", i+1, m.Text)
				if m.Level > 0 {
					fmt.Printf(" (level %d)", m.Level)
				}
				fmt.Println()
			}
		}
		fmt.Println()
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
	parts := strings.Split(s, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
