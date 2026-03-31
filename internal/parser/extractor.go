package parser

import "strings"

type Entry struct {
    Hanzi             string     `json:"hanzi"`
    Pinyin            string     `json:"pinyin"`
    PinyinNormalized  string     `json:"pinyin_normalized"`
    Meanings          []Meaning  `json:"meanings"`
}

type Meaning struct {
    Text        string   `json:"text"`
    PartOfSpeech string  `json:"part_of_speech,omitempty"`
    Refs        []string `json:"refs,omitempty"`
    Examples    []string `json:"examples,omitempty"`
    Order       int      `json:"order"`
}

func ExtractEntries(root *Node, limit int) []Entry {
    var entries []Entry

    var current *Entry

    for i := 0; i < len(root.Children); i++ {
    	// get limit for debug
    	if limit > 0 && len(entries) >= limit{
	    	break
	    }
					
        node := root.Children[i]

        // HEADER (hanzi + pinyin)
        if node.Type == NodeText {
            hanzi, pinyin := SplitHanziPinyin(node.Value)

            entry := Entry{
                Hanzi:            hanzi,
                Pinyin:           pinyin,
                PinyinNormalized: NormalizePinyin(pinyin),
            }

            entries = append(entries, entry)
            current = &entries[len(entries)-1]

            continue
        }

        // meanings
        if node.Type == NodeUnknown && current != nil {
            meanings := ExtractMeanings(node)
            current.Meanings = meanings
        }
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

    for _, child := range node.Children {

        switch child.Type {

        case NodeParagraph:
            // part of speech
            current.PartOfSpeech = ExtractText(child)

        case NodeText:
            text := ExtractText(child)
            if text != "" {
                current.Text += text + " "
            }

        case NodeRef:
            ref := ExtractText(child)
            if ref != "" {
                current.Refs = append(current.Refs, ref)
            }

        case NodeExample:
            ex := ExtractText(child)
            if ex != "" {
                current.Examples = append(current.Examples, ex)
            }
        }
    }

    current.Text = strings.TrimSpace(current.Text)

    if current.Text != "" {
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
        result += ExtractText(c) + " "
    }

    return strings.TrimSpace(result)
}