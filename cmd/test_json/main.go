package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"

	"os"
	"parser/internal/parser"

)

func main() {
	// path to first part of Dictionray
	path := "./dabkrs/dabkrs_1.dsl"
	fmt.Printf("Start parsing, path: %s \n", path)

	file, err := os.Create("test.json")
	if err != nil {
		fmt.Errorf(err.Error())
	}
	defer file.Close()

	const batchSize = 1000
	batch := make([]parser.Entry, 0, batchSize)

	r, err := parser.OpenDSL(path)
	if err != nil {
		log.Printf("OpenDSL %s: %v", file, err)
		return
	}
	defer r.Close()
	
	w := bufio.NewWriter(file)
	defer w.Flush()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	done := false
	w.WriteString("[\n")
	first := true
	parser.ParseFSMStream(r, func(re parser.RawEntry) {
		if done{
			return
		}
		e := convertSingleEntry(re)

		if e.Headword != "" {
			batch = append(batch, e)
			
			if len(batch) >= batchSize{
				for _, e := range batch{
					if !first {
						w.WriteString(",\n")
					}
					first = false
					
					err := encoder.Encode(e)
					if err != nil{
						return
					}
				}
				batch = batch[:0]
				done = true
				return
			}
		}
	})
	
	w.WriteString("\n]")

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
