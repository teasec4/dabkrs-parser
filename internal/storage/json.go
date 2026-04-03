package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"parser/internal/parser"
)

func StreamEntiresToJSON(root *parser.Node, filename string, limit int) error {
	entries := parser.ExtractEntries(root, limit)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	w.Write([]byte("[\n"))

	for i, entry := range entries {
		if i > 0 {
			w.Write([]byte(",\n"))
		}
		if err := encoder.Encode(entry); err != nil {
			return err
		}
	}

	w.Write([]byte("\n]\n"))

	return nil
}
