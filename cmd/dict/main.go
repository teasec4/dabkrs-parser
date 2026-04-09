package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"parser/internal/parser"
	"parser/internal/storage"
	"strings"
	"unicode"
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

	const batchSize = 1000
	batch := make([]parser.Entry, 0, batchSize)

	for _, file := range fileList {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", file)
			continue
		}

		fmt.Printf("Parsing %s...\n", file)
		// Read DSL file
		// by Stream
		r, err := parser.OpenDSL(file)
		if err != nil {
			log.Printf("OpenDSL %s: %v", file, err)
			continue
		}
		defer r.Close()

		fileEntryCount := 0

		// Use FSM parser stream with callback - no memory accumulation
		parser.ParseFSMStream(r, func(entry parser.RawEntry) {
			e := convertSingleEntry(entry)
			// if e is has a problem with Validation its return empty object
			if e.Headword == "" {
				return
			}

			batch = append(batch, e)
			fileEntryCount++

			if len(batch) >= batchSize {
				_, err := db.InsertEntries(batch, batchSize)
				if err != nil {
					log.Printf("InsertEntries batch: %v", err)
				}
				batch = batch[:0]
			}

			if fileEntryCount%10000 == 0 {
				fmt.Printf("Processed %d entries...\n", fileEntryCount)
			}
		})

		if len(batch) > 0 {
			_, err := db.InsertEntries(batch, batchSize)
			if err != nil {
				log.Printf("InsertEntries final batch: %v", err)
			}
			batch = batch[:0]
		}

		total += fileEntryCount
		fmt.Printf("Inserted %d entries from %s\n", total, file)
	}

	count, _ := db.Count()
	fmt.Printf("\nTotal: %d entries in database\n", count)
}

func convertSingleEntry(raw parser.RawEntry) parser.Entry {
	// check for Hanzi (Headword)
	if raw.Headword == "" {
		return parser.Entry{}
	}
	entry := parser.Entry{
		Headword:         raw.Headword,
		Pinyin:           strings.ReplaceAll(raw.Pinyin, " ", ""),
		PinyinNormalized: parser.NormalizePinyin(raw.Pinyin),
		Meanings:         make([]parser.Meaning, 0),
	}
	for _, rm := range raw.Meanings {
		// here add a some Filter method
		if shouldKeepMeaning(rm) {
			meaning := parser.Meaning{
				Level: rm.Level,
				Text:  rm.Text,
				Tags:  rm.Tags,
				Order: len(entry.Meanings),
			}
			entry.Meanings = append(entry.Meanings, meaning)
		}

	}
	return entry
}

// filtering out 垃圾 meanings
func shouldKeepMeaning(m parser.RawMeaning) bool {
	// if empty
	if strings.TrimSpace(m.Text) == "" {
		return false
	}

	if len(m.Text) < 2 {
		return false
	}

	return true
}

func isCommonChinese(s string) bool {
	for _, r := range s {
		if !isCommonChineseChar(r) {
			return false
		}
	}

	return true
}

func isCommonChineseChar(r rune) bool {
	if !unicode.Is(unicode.Han, r) {
		return false
	}
	if isRadical(r) {
		return false
	}
	// Диалектные цифры (〢, 〥, 〨 и т.д.)
	if r >= 0x3000 && r <= 0x303F {
		return false
	}
	// CJK Compatibility (декоративные варианты)
	if r >= 0xF900 && r <= 0xFAFF {
		return false
	}
	// Расширение A — редкие иероглифы
	if r >= 0x3400 && r <= 0x4DBF {
		return false
	}
	// Расширения B-F (очень редкие иероглифы)
	if r >= 0x20000 && r <= 0x2A6DF {
		return false
	}
	if r >= 0x2A700 && r <= 0x2B73F {
		return false
	}
	if r >= 0x2B740 && r <= 0x2B81F {
		return false
	}
	if r >= 0x2B820 && r <= 0x2CEAF {
		return false
	}
	if r >= 0x2CEB0 && r <= 0x2EBEF {
		return false
	}
	return true
}

func isRadical(r rune) bool {
	return (r >= 0x2E80 && r <= 0x2EFF) ||
		(r >= 0x2F00 && r <= 0x2FDF) ||
		(r >= 0x2FF0 && r <= 0x2FFF)
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
