package parser

import (
	"strings"
)

type Entry struct {
	Hanzi            string    `json:"hanzi"`
	Pinyin           string    `json:"pinyin"`
	PinyinNormalized string    `json:"pinyin_normalized"`
	Meanings         []Meaning `json:"meanings"`
}

type Meaning struct {
	Text         string   `json:"text"`
	PartOfSpeech string   `json:"part_of_speech,omitempty"`
	Refs         []string `json:"refs,omitempty"`
	Examples     []string `json:"examples,omitempty"`
	Order        int      `json:"order"`
}

func ExtractEntries(root *Node, limit int) []Entry {
	var entries []Entry

	var current *Entry

	for i := 0; i < len(root.Children); i++ {
		node := root.Children[i]

		if node.Type == NodeText {
			value := node.Value

			lines := strings.Split(value, "\n")
			var pendingLine string

			for _, line := range lines {
				line = strings.TrimRight(line, "\r")
				line = strings.TrimSpace(line)

				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				if isPinyin(line) {
					if pendingLine != "" {
						entry := Entry{
							Hanzi:            pendingLine,
							Pinyin:           line,
							PinyinNormalized: NormalizePinyin(line),
							Meanings:         []Meaning{},
						}
						entries = append(entries, entry)
						current = &entries[len(entries)-1]
						pendingLine = ""
					}
					continue
				}

				if hasChinese(line) {
					hanzi, pinyin := SplitHanziPinyin(line)
					if pinyin != "" {
						entry := Entry{
							Hanzi:            hanzi,
							Pinyin:           pinyin,
							PinyinNormalized: NormalizePinyin(pinyin),
							Meanings:         []Meaning{},
						}
						entries = append(entries, entry)
						current = &entries[len(entries)-1]
					} else {
						pendingLine = hanzi
					}
					continue
				}

				hanzi, pinyin := SplitHanziPinyin(line)
				if hanzi != "" {
					entry := Entry{
						Hanzi:            hanzi,
						Pinyin:           pinyin,
						PinyinNormalized: NormalizePinyin(pinyin),
						Meanings:         []Meaning{},
					}
					entries = append(entries, entry)
					current = &entries[len(entries)-1]
				}
			}
			continue
		}

		if node.Type == NodeUnknown && current != nil {
			meanings := ExtractMeanings(node)
			for j := range meanings {
				meanings[j].Order = len(current.Meanings) + j
			}
			current.Meanings = append(current.Meanings, meanings...)
		}
	}

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries
}

func SplitHanziPinyin(s string) (string, string) {
	parts := strings.Fields(s)

	if len(parts) < 2 {
		return s, ""
	}

	// heuristic:
	// китайские иероглифы обычно первый блок
	hanzi := parts[0]
	pinyin := strings.Join(parts[1:], " ")

	return hanzi, pinyin
}

func isPinyin(s string) bool {
	if len(s) == 0 {
		return false
	}
	hasLetter := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			hasLetter = true
			continue
		}
		if r == '\'' || r == 0x2019 || r == ' ' || r == 'ō' || r == 'ó' || r == 'ě' || r == 'è' ||
			r == 'ā' || r == 'á' || r == 'ǎ' || r == 'à' ||
			r == 'ē' || r == 'é' || r == 'ě' || r == 'è' ||
			r == 'ī' || r == 'í' || r == 'ǐ' || r == 'ì' ||
			r == 'ū' || r == 'ú' || r == 'ǔ' || r == 'ù' ||
			r == 'ǖ' || r == 'ǘ' || r == 'ǚ' || r == 'ǜ' {
			continue
		}
		return false
	}
	return hasLetter
}

func hasChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

func NormalizePinyin(p string) string {
	p = strings.ToLower(p)

	replacer := strings.NewReplacer(
		"ā", "a", "á", "a", "ǎ", "a", "à", "a",
		"ē", "e", "é", "e", "ě", "e", "è", "e",
		"ī", "i", "í", "i", "ǐ", "i", "ì", "i",
		"ō", "o", "ó", "o", "ǒ", "o", "ò", "o",
		"ū", "u", "ú", "u", "ǔ", "u", "ù", "u",
		"ǖ", "u", "ǘ", "u", "ǚ", "u", "ǜ", "u",
	)

	return replacer.Replace(p)
}

func ExtractMeanings(node *Node) []Meaning {
	var meanings []Meaning

	var current Meaning

	var collectExamples func(n *Node)
	collectExamples = func(n *Node) {
		for _, c := range n.Children {
			if c.Type == NodeExample {
				ex := strings.TrimSpace(ExtractText(c))
				if ex != "" {
					current.Examples = append(current.Examples, ex)
				}
			}
			collectExamples(c)
		}
	}

	for _, child := range node.Children {
		switch child.Type {
		case NodeParagraph:
			text := strings.TrimSpace(ExtractText(child))
			if text != "" {
				current.PartOfSpeech = text
			}

		case NodeText:
			text := strings.TrimSpace(ExtractText(child))
			if text != "" {
				skipSpace := text == "(" || text == ")" || strings.HasPrefix(text, "(")
				if current.Text != "" && !skipSpace {
					current.Text += " "
				}
				current.Text += text
			}

		case NodeRef:
			ref := strings.TrimSpace(ExtractText(child))
			if ref != "" {
				current.Refs = append(current.Refs, ref)
				if current.Text != "" && !strings.HasSuffix(current.Text, " ") {
					current.Text += " "
				}
				current.Text += "→" + ref
			}

		case NodeItalic:
			text := strings.TrimSpace(ExtractText(child))
			if text != "" {
				if current.Text != "" && !strings.HasSuffix(current.Text, "(") {
					current.Text += " "
				}
				current.Text += text
			}

		case NodeContainer:
			text := strings.TrimSpace(ExtractText(child))
			if text != "" {
				if current.Text != "" && !strings.HasSuffix(current.Text, " ") && !strings.HasSuffix(current.Text, ")") {
					current.Text += " "
				}
				current.Text += text
			}

		case NodeExample:
			ex := strings.TrimSpace(ExtractText(child))
			if ex != "" {
				current.Examples = append(current.Examples, ex)
			}

		case NodeStar:
			collectExamples(child)
		}
	}

	current.Text = strings.TrimSpace(current.Text)

	if current.Text != "" || len(current.Examples) > 0 {
		current.Order = 0
		meanings = append(meanings, current)
	}

	return meanings
}

func ExtractText(n *Node) string {
	if n.Type == NodeText {
		return n.Value
	}

	var result string
	for _, c := range n.Children {
		text := ExtractText(c)
		if text == "" {
			continue
		}

		switch c.Type {
		case NodeItalic, NodeContainer:
			result += "(" + text + ")"
		case NodeRef:
			result += text
		default:
			result += text + " "
		}
	}

	return strings.TrimSpace(result)
}
