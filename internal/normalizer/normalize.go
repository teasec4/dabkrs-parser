package normalizer

import (
	"regexp"
	"strings"
)

func CleanText(s string) string {
    s = strings.TrimSpace(s)

    // убрать лишние пробелы
    s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")

    // убрать пустые скобки
    s = strings.ReplaceAll(s, "( )", "")
    s = strings.ReplaceAll(s, "()", "")

    return strings.TrimSpace(s)
}

func UniqueStrings(input []string) []string {
    seen := map[string]bool{}
    var result []string

    for _, v := range input {
        v = strings.TrimSpace(v)
        if v == "" {
            continue
        }

        if !seen[v] {
            seen[v] = true
            result = append(result, v)
        }
    }

    return result
}

func isDSLMetadata(text string) bool {
	// Check if text contains DSL metadata directives
	// DSL metadata starts with # (e.g., #NAME, #INDEX_LANGUAGE, #CONTENTS_LANGUAGE, #INCLUDE)
	return strings.Contains(text, "#NAME") ||
		strings.Contains(text, "#INDEX_LANGUAGE") ||
		strings.Contains(text, "#CONTENTS_LANGUAGE") ||
		strings.Contains(text, "#INCLUDE")
}

