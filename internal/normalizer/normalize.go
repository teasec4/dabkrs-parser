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

