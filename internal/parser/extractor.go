package parser

import (
	"strings"
)

// ExtractEntries extracts dictionary entries from AST
func ExtractEntries(root *Node, limit int) []Entry {
	var entries []Entry
	var currentEntry *Entry

	processEntry := func() {
		if currentEntry != nil && len(currentEntry.Meanings) > 0 {
			entries = append(entries, *currentEntry)
		}
		currentEntry = nil
	}

	processMeaning := func(node *Node) {
		if currentEntry != nil {
			meaning := extractMeaning(node, len(currentEntry.Meanings))
			if meaning.Text != "" || len(meaning.Tags) > 0 {
				currentEntry.Meanings = append(currentEntry.Meanings, meaning)
			}
		}
	}

	for i := 0; i < len(root.Children); i++ {
		child := root.Children[i]

		switch child.Type {
		case NodeText:
			text := strings.TrimSpace(child.Value)
			if text == "" || strings.HasPrefix(text, "#") {
				continue
			}

			lines := strings.Split(text, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				if HasChinese(line) {
					headword, pinyin := SplitHanziPinyin(line)
					if pinyin == "" {
						headword = line
					}
					processEntry()
					currentEntry = &Entry{
						Headword: headword,
						Pinyin:   pinyin,
						Meanings: []Meaning{},
					}
				} else if currentEntry != nil {
					line = strings.TrimSpace(line)
					if line == "_" {
						currentEntry.Pinyin = ""
					} else if IsPinyin(line) {
						currentEntry.Pinyin = line
					}
				}
			}

		case NodeMeaning:
			processMeaning(child)
		}
	}

	processEntry()

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries
}

// extractMeaning extracts a single meaning from AST node
func extractMeaning(node *Node, order int) Meaning {
	m := Meaning{
		Level: getLevel(node.Value),
		Text:  "",
		Tags:  []Tag{},
		Order: order,
	}

	var textParts []string

	for _, child := range node.Children {
		switch child.Type {
		case NodeText:
			text := strings.TrimSpace(child.Value)
			if text != "" {
				textParts = append(textParts, text)
			}

		case NodeParagraph:
			text := extractTextContent(child)
			if text != "" {
				m.Tags = append(m.Tags, Tag{Type: "p", Value: text})
			}

		case NodeItalic:
			text := extractTextContent(child)
			if text != "" {
				m.Tags = append(m.Tags, Tag{Type: "i", Value: text})
			}

		case NodeRef:
			ref := extractTextContent(child)
			if ref != "" && HasChinese(ref) {
				m.Tags = append(m.Tags, Tag{Type: "ref", Value: ref})
			}

		case NodeExample:
			ex := extractTextContent(child)
			if ex != "" {
				m.Tags = append(m.Tags, Tag{Type: "ex", Value: ex})
			}

		case NodeStar:
			ex := extractTextContent(child)
			if ex != "" {
				m.Tags = append(m.Tags, Tag{Type: "*", Value: ex})
			}
		}
	}

	m.Text = strings.Join(textParts, " ")
	return m
}

// extractTextContent extracts text content from a node and its children
func extractTextContent(node *Node) string {
	if node.Type == NodeText {
		return strings.TrimSpace(node.Value)
	}

	var parts []string
	for _, child := range node.Children {
		text := extractTextContent(child)
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

// getLevel extracts level number from tag (e.g., "m1" -> 1, "m2" -> 2)
func getLevel(tag string) int {
	level := 0
	for _, c := range tag {
		if c >= '1' && c <= '9' {
			level = level*10 + int(c-'0')
		}
	}
	return level
}

// HasChinese checks if string contains Chinese characters
func HasChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

// SplitHanziPinyin splits Chinese characters and pinyin
func SplitHanziPinyin(s string) (string, string) {
	parts := strings.Fields(s)
	if len(parts) < 2 {
		return s, ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

// NormalizePinyin normalizes pinyin by removing tones
func NormalizePinyin(p string) string {
	p = strings.ToLower(p)
	replacer := strings.NewReplacer(
		"ā", "a", "á", "a", "ǎ", "a", "à", "a",
		"ē", "e", "é", "e", "ě", "e", "è", "e",
		"ī", "i", "í", "i", "ǐ", "i", "ì", "i",
		"ō", "o", "ó", "o", "ǒ", "o", "ò", "o",
		"ū", "u", "ú", "u", "ǔ", "u", "ù", "u",
		"ǖ", "v", "ǘ", "v", "ǚ", "v", "ǜ", "v",
	)
	return replacer.Replace(p)
}

// IsPinyin checks if string looks like pinyin
func IsPinyin(s string) bool {
	if len(s) == 0 {
		return false
	}
	hasLetter := false
	toneChars := "āáǎàēéěèīíǐìōóǒòūúǔùǖǘǚǜ"
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			hasLetter = true
			continue
		}
		if r == '\'' || r == 0x2019 || r == ' ' || r == ',' || r == '.' || r == '-' {
			continue
		}
		if r >= 0x00C0 && r <= 0x024F {
			continue
		}
		if strings.ContainsRune(toneChars, r) {
			continue
		}
		return false
	}
	return hasLetter
}
