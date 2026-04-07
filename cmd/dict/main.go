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
        
		// Use FSM parser stream
		parser.ParseFSMStream(r, func(entry parser.RawEntry) {
			e := convertSingleEntry(entry)
			
			if e.Headword != ""{
				// feature create a butch
				// 
				batch = append(batch, e)
				fileEntryCount ++
				
				if len(batch) >= batchSize{
					_, err := db.InsertEntries(batch, batchSize)
					if err != nil {
						log.Printf("InsertEntries batch: %v", err)
					}
					batch = batch[:0] // очистить
				}
				
			}
			
			fileEntryCount++
			if fileEntryCount % 10000 == 0 {
                fmt.Printf("Processed %d entries...\n", total)
            }
            
            
		})
		
		if len(batch) > 0{
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
    entry := parser.Entry{
        Headword:         raw.Headword,
        Pinyin:           raw.Pinyin,
        PinyinNormalized: parser.NormalizePinyin(raw.Pinyin),
        Meanings:         make([]parser.Meaning, 0),
    }
    for _, rm := range raw.Meanings {
        meaning := parser.Meaning{
            Level: rm.Level,
            Text:  rm.Text,
            Tags:  rm.Tags,
            Order: len(entry.Meanings),
        }
        entry.Meanings = append(entry.Meanings, meaning)
    }
    return entry
}

func convertRawEntries(raw []parser.RawEntry, limit int) []parser.Entry {
	entries := make([]parser.Entry, 0)
	for i, re := range raw {
		if limit > 0 && i >= limit {
			break
		}
		entry := parser.Entry{
			Headword:         re.Headword,
			Pinyin:           re.Pinyin,
			PinyinNormalized: parser.NormalizePinyin(re.Pinyin),
			Meanings:         make([]parser.Meaning, 0),
		}
		for _, rm := range re.Meanings {
			meaning := parser.Meaning{
				Level: rm.Level,
				Text:  rm.Text,
				Tags:  rm.Tags,
				Order: len(entry.Meanings),
			}
			entry.Meanings = append(entry.Meanings, meaning)
		}
		if entry.Headword != "" {
			entries = append(entries, entry)
		}
	}
	return entries
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
