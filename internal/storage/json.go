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
			// проверяем что это реальный entry
			if i+1 >= len(root.Children) || root.Children[i+1].Type != parser.NodeUnknown {
				continue
			}
			
			// теперь закрываем предыдущий
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

			hanzi, pinyin := parser.SplitHanziPinyin(node.Value)

			entry := parser.Entry{
				Hanzi:            hanzi,
				Pinyin:           pinyin,
				PinyinNormalized: parser.NormalizePinyin(pinyin),
			}

			current = &entry
			continue
		}

		if node.Type == parser.NodeUnknown && current != nil {
			current.Meanings = append(current.Meanings, parser.ExtractMeanings(node)...)
			if !first {
				w.Write([]byte(",\n"))
			}

			if err := encoder.Encode(current); err != nil {
				return err
			}

			first = false
			count++
		}
	}

	w.Write([]byte("]\n"))

	return nil
}
