package parser

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"unicode"
)

type RawEntry struct {
	Headword string
	Pinyin   string
	Meanings []RawMeaning
}

type RawMeaning struct {
	Level int
	Text  string
	Tags  []Tag
}

type ParseState int

const (
	StateExpectHeadword ParseState = iota
	StateExpectPinyin
	StateExpectMeaning
)

func ParseFSMStream(r io.Reader, onEntry func(RawEntry)) {
	var current *RawEntry
	state := StateExpectHeadword
	var meaningBuffer strings.Builder

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimRight(line, "\r")

		if shouldSkipLine(line) {
			continue
		}

		switch state {
		case StateExpectHeadword:
			line = strings.TrimSpace(line)
			if containsChinese(line) {
				if current != nil && current.Headword != "" {
					onEntry(*current)
				}
				current = &RawEntry{
					Headword: line,
					Pinyin:   "",
					Meanings: nil,
				}
				state = StateExpectPinyin
			}

		case StateExpectPinyin:
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "[m") {
				state = StateExpectMeaning
				meaningBuffer.Reset()
				meaningBuffer.WriteString(line)
				if current != nil {
					meanings := parseMeaningBlock(&meaningBuffer, scanner)
					current.Meanings = append(current.Meanings, meanings...)
				}
			} else if !containsChinese(line) && !strings.HasPrefix(line, "[") {
				if current != nil {
					current.Pinyin = line
				}
				state = StateExpectMeaning
			}

		case StateExpectMeaning:
			rawLine := line
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "[m") {
				// Normal meaning line
				meaningBuffer.Reset()
				meaningBuffer.WriteString(line)
				if current != nil {
					meanings := parseMeaningBlock(&meaningBuffer, scanner)
					current.Meanings = append(current.Meanings, meanings...)
				}
			} else if containsChinese(line) && !startsWithWhitespace(rawLine) {
				// Chinese at column 0 = new headword (NOT a meaning line with embedded Chinese)
				if current != nil {
					onEntry(*current)
				}

				current = &RawEntry{
					Headword: line,
					Pinyin:   "",
					Meanings: nil,
				}
				state = StateExpectPinyin
			} else if strings.HasPrefix(line, "#") {
				state = StateExpectHeadword
			} else if startsWithWhitespace(rawLine) {
				// Malformed meaning: indented line that's not [m]-tagged — treat as implicit meaning
				text := extractMeaningText(line)
				tags := extractTags(line)
				if text != "" && len([]rune(text)) >= 2 {
					if current != nil {
						current.Meanings = append(current.Meanings, RawMeaning{
							Level: 1,
							Text:  text,
							Tags:  tags,
						})
					}
				}
			}
		}
	}

	if current != nil && current.Headword != "" {
		onEntry(*current)
	}

}

func shouldSkipLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return true
	}
	if strings.HasPrefix(line, "#") {
		return true
	}
	return false
}

func containsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func startsWithWhitespace(s string) bool {
	if len(s) == 0 {
		return false
	}
	return s[0] == ' ' || s[0] == '\t'
}

func parseMeaningBlock(buf *strings.Builder, scanner *bufio.Scanner) []RawMeaning {
	// Check if the first line already contains the full meaning block
	raw := buf.String()

	// Handle multiple meanings on same line: [m1]...[/m][m2]...
	allMeanings := make([]RawMeaning, 0)
	for {
		startIdx := strings.Index(raw, "[m")
		if startIdx == -1 {
			break
		}

		// Find level
		level := 1
		rest := raw[startIdx+2:]
		for i := 0; i < len(rest) && i < 2; i++ {
			if rest[i] >= '1' && rest[i] <= '9' {
				level = int(rest[i] - '0')
				break
			}
		}

		// Find closing [/m]
		closeIdx := strings.Index(raw[startIdx:], "[/m]")
		if closeIdx == -1 {
			// Need to read more lines to find closing tag
			if scanner.Scan() {
				nextLine := scanner.Text()
				nextLine = strings.TrimRight(nextLine, "\r")
				buf.WriteString("\n")
				buf.WriteString(nextLine)
				raw = buf.String()
				continue
			}
			break
		}
		closeIdx += startIdx

		meaningBlock := raw[startIdx:closeIdx]
		text := extractMeaningText(meaningBlock)
		tags := extractTags(meaningBlock)

		allMeanings = append(allMeanings, RawMeaning{
			Level: level,
			Text:  text,
			Tags:  tags,
		})

		if closeIdx+4 >= len(raw) {
			break
		}
		raw = raw[closeIdx+4:]
	}

	if len(allMeanings) == 0 {
		return []RawMeaning{{Level: 1, Text: "", Tags: nil}}
	}

	return allMeanings
}

func extractLevel(s string) int {
	for i := 1; i <= 9; i++ {
		if strings.Contains(s, "[m"+string(rune('0'+i))) {
			return i
		}
	}
	return 1
}

func extractMeaningText(s string) string {
	// Remove parameterized color tags: [c red], [c blue], [c brown], [c green], [c violet], etc.
	s = regexp.MustCompile(`\[c\s+\w+\]`).ReplaceAllString(s, "")

	replacer := strings.NewReplacer(
		// Теги уровней
		"[m1]", "", "[m2]", "", "[m3]", "", "[m4]", "", "[m5]", "",
		"[m6]", "", "[m7]", "", "[m8]", "", "[m9]", "", "[m]", "",
		"[/m]", "",
		// Теги форматирования
		"[i]", "", "[/i]", "",
		"[c]", "", "[/c]", "",
		"[p]", "", "[/p]", "",
		"[b]", "", "[/b]", "",
		"[*]", "", "[/*]", "",
		// Теги ссылок и примеров
		"[ref]", "", "[/ref]", "", "[ex]", "", "[/ex]", "",
		// Цифры со скобками
		"1)", "", "2)", "", "3)", "", "4)", "", "5)", "",
		"6)", "", "7)", "", "8)", "", "9)", "", "0)", "",
		"(1)", "", "(2)", "", "(3)", "", "(4)", "", "(5)", "",
		"(6)", "", "(7)", "", "(8)", "", "(9)", "", "(0)", "",
	)
	s = replacer.Replace(s)
	s = strings.TrimSpace(s)
	return s
}

func extractTags(s string) []Tag {
	tags := make([]Tag, 0)

	for {
		idx := strings.Index(s, "[ref]")
		if idx == -1 {
			break
		}
		s = s[idx+5:]
		closeIdx := strings.Index(s, "[/ref]")
		if closeIdx == -1 {
			break
		}
		value := strings.TrimSpace(s[:closeIdx])
		tags = append(tags, Tag{Type: "ref", Value: value})
		s = s[closeIdx+6:]
	}

	for {
		idx := strings.Index(s, "[ex]")
		if idx == -1 {
			break
		}
		s = s[idx+4:]
		closeIdx := strings.Index(s, "[/ex]")
		if closeIdx == -1 {
			break
		}
		value := strings.TrimSpace(s[:closeIdx])
		tags = append(tags, Tag{Type: "ex", Value: value})
		s = s[closeIdx+5:]
	}

	return tags
}

func NormalizePinyin(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "’", "")
	s = strings.ReplaceAll(s, "”", "")
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSpace(s)

	replacer := strings.NewReplacer(
		"ā", "a", "á", "a", "ǎ", "a", "à", "a",
		"ē", "e", "é", "e", "ě", "e", "è", "e",
		"ī", "i", "í", "i", "ǐ", "i", "ì", "i",
		"ō", "o", "ó", "o", "ǒ", "o", "ò", "o",
		"ū", "u", "ú", "u", "ǔ", "u", "ù", "u",
		"ǖ", "v", "ǘ", "v", "ǚ", "v", "ǜ", "v",
	)
	return replacer.Replace(s)
}
