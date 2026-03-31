package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"parser/internal/parser"
)

func StreamEntiresToJSON(root *parser.Node, filename string, limit int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	encoder := json.NewEncoder(w)

	// Start JSON array
	w.Write([]byte("[\n"))

	count := 0
	first := true

	var current *parser.Entry

	for i := 0; i < len(root.Children); i++ {
		if limit > 0 && count >= limit {
			break
		}

		node := root.Children[i]

		if node.Type == parser.NodeText {
			// Check if this is a real entry (next element should be NodeUnknown)
			if i+1 >= len(root.Children) || root.Children[i+1].Type != parser.NodeUnknown {
				continue
			}

			// Write previous entry if it exists
			if current != nil {
				if !first {
					w.Write([]byte(",\n"))
				}

				if err := encoder.Encode(current); err != nil {
					return err
				}

				first = false
				count++
			}

			// Create new entry
			hanzi, pinyin := parser.SplitHanziPinyin(node.Value)

			entry := parser.Entry{
				Hanzi:            hanzi,
				Pinyin:           pinyin,
				PinyinNormalized: parser.NormalizePinyin(pinyin),
				Meanings:         []parser.Meaning{}, // Initialize empty array
			}

			current = &entry
			continue
		}

		if node.Type == parser.NodeUnknown && current != nil {
			// Extract meanings and add to current entry
			meanings := parser.ExtractMeanings(node)

			// Set correct order for each meaning
			for j := range meanings {
				meanings[j].Order = len(current.Meanings) + j
			}

			current.Meanings = append(current.Meanings, meanings...)
		}
	}

	// Write the last entry if it exists
	if current != nil {
		if !first {
			w.Write([]byte(",\n"))
		}
		if err := encoder.Encode(current); err != nil {
			return err
		}
	}

	// End JSON array
	w.Write([]byte("\n]\n"))

	return nil
}
