package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Entry represents a parsed dictionary entry
type Entry struct {
	Chinese          string   `json:"chinese"`           // Chinese characters
	Pinyin           string   `json:"pinyin"`            // Pinyin transcription
	Meanings         []string `json:"meanings"`          // Cleaned meanings/translations as array
	MeaningsCombined string   `json:"meanings_combined"` // All meanings combined as single string
}

func main() {
	// First, let's parse the existing dump.txt file
	entries, err := parseDumpFile("./dump.txt")
	if err != nil {
		fmt.Printf("Error parsing dump.txt: %v\n", err)
		return
	}

	// Print parsed entries
	fmt.Printf("Parsed %d entries:\n", len(entries))
	for i, entry := range entries {
		if i >= 10 { // Show only first 10 entries
			fmt.Printf("... and %d more entries\n", len(entries)-10)
			break
		}
		fmt.Printf("%d. %s [%s]\n", i+1, entry.Chinese, entry.Pinyin)
		for j, meaning := range entry.Meanings {
			fmt.Printf("   %d) %s\n", j+1, meaning)
		}
		fmt.Println()
	}

	// Save cleaned results to a text file
	err = saveCleanedResults(entries, "./cleaned_results.txt")
	if err != nil {
		fmt.Printf("Error saving cleaned results: %v\n", err)
		return
	}

	fmt.Println("Cleaned results saved to cleaned_results.txt")

	// Save as JSON for SQL import
	err = saveAsJSON(entries, "./dictionary.json")
	if err != nil {
		fmt.Printf("Error saving JSON: %v\n", err)
		return
	}

	fmt.Println("JSON data saved to dictionary.json")
}

// parseDumpFile parses the dump.txt file and returns structured entries
func parseDumpFile(filename string) ([]Entry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var entries []Entry
	var currentEntry *Entry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and metadata lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip lines that are just references (contain "см." or "см.[/p]")
		if strings.Contains(line, "см.") || strings.Contains(line, "см.[/p]") {
			continue
		}

		// Check if this line starts a new entry (Chinese characters without leading spaces)
		if isChineseEntryStart(line) && !strings.HasPrefix(line, " ") {
			// Save previous entry if exists
			if currentEntry != nil && currentEntry.Chinese != "" {
				entries = append(entries, *currentEntry)
			}

			// Start new entry
			currentEntry = &Entry{
				Meanings: []string{},
			}

			// Try to parse the first line of the entry
			parts := parseEntryFirstLine(line)
			if len(parts) >= 2 {
				currentEntry.Chinese = parts[0]
				currentEntry.Pinyin = parts[1]
			} else if len(parts) == 1 {
				currentEntry.Chinese = parts[0]
			}
		} else if currentEntry != nil {
			// This is a continuation line (pinyin or meaning)
			if currentEntry.Pinyin == "" && isPinyinLine(line) {
				currentEntry.Pinyin = strings.TrimSpace(line)
			} else {
				// Extract meanings from the line
				meanings := extractMeanings(line)
				currentEntry.Meanings = append(currentEntry.Meanings, meanings...)
			}
		}
	}

	// Add the last entry
	if currentEntry != nil && currentEntry.Chinese != "" {
		entries = append(entries, *currentEntry)
	}

	if err := scanner.Err(); err != nil {
		return entries, fmt.Errorf("scanner error: %w", err)
	}

	return entries, nil
}

// isChineseEntryStart checks if a line looks like the start of a Chinese entry
func isChineseEntryStart(line string) bool {
	// Check if line contains Chinese characters (basic check)
	for _, r := range line {
		if (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
			(r >= 0x3400 && r <= 0x4DBF) || // CJK Unified Ideographs Extension A
			(r >= 0x20000 && r <= 0x2A6DF) { // CJK Unified Ideographs Extension B
			return true
		}
	}
	return false
}

// isPinyinLine checks if a line looks like pinyin (contains Latin letters with tone marks)
func isPinyinLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	// Pinyin lines usually don't start with brackets
	if strings.HasPrefix(line, "[") {
		return false
	}

	// Check if line contains Chinese characters (if yes, it's not pinyin)
	for _, r := range line {
		if r >= 0x4E00 && r <= 0x9FFF {
			return false
		}
	}

	// Check for Latin letters, tone marks, and pinyin-specific characters
	hasLatin := false
	hasPinyinChar := false
	for _, r := range line {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLatin = true
		}
		// Check for pinyin tone marks and special characters
		if r == 'ā' || r == 'á' || r == 'ǎ' || r == 'à' ||
			r == 'ē' || r == 'é' || r == 'ě' || r == 'è' ||
			r == 'ī' || r == 'í' || r == 'ǐ' || r == 'ì' ||
			r == 'ō' || r == 'ó' || r == 'ǒ' || r == 'ò' ||
			r == 'ū' || r == 'ú' || r == 'ǔ' || r == 'ù' ||
			r == 'ǖ' || r == 'ǘ' || r == 'ǚ' || r == 'ǜ' ||
			r == ' ' || r == '\'' || r == '’' {
			hasPinyinChar = true
		}
	}

	return hasLatin && (hasPinyinChar || !strings.ContainsAny(line, "[]"))
}

// parseEntryFirstLine parses the first line of an entry which may contain Chinese and pinyin
func parseEntryFirstLine(line string) []string {
	line = strings.TrimSpace(line)
	var parts []string

	// Split by spaces
	words := strings.Fields(line)
	if len(words) == 0 {
		return parts
	}

	// First part is Chinese
	parts = append(parts, words[0])

	// Check if there's pinyin in the same line
	if len(words) > 1 {
		// Join remaining parts as potential pinyin
		potentialPinyin := strings.Join(words[1:], " ")
		if isPinyinLine(potentialPinyin) {
			parts = append(parts, potentialPinyin)
		}
	}

	return parts
}

// extractMeanings extracts cleaned meanings from a DSL formatted line
func extractMeanings(line string) []string {
	var meanings []string

	// Regex to find [mX]...[/m] patterns
	re := regexp.MustCompile(`\[m\d+\](.*?)\[/m\]`)
	matches := re.FindAllStringSubmatch(line, -1)

	for _, match := range matches {
		if len(match) > 1 {
			cleaned := cleanDSL(match[1])
			if cleaned != "" {
				meanings = append(meanings, cleaned)
			}
		}
	}

	// If no [mX] patterns found, check if the whole line is a meaning
	if len(meanings) == 0 && strings.Contains(line, "[") {
		// Try to extract any text within brackets
		cleaned := cleanDSL(line)
		if cleaned != "" {
			meanings = append(meanings, cleaned)
		}
	}

	return meanings
}

// cleanDSL removes DSL formatting tags from text while preserving content
func cleanDSL(text string) string {
	// First, preserve content inside tags before removing the tags
	// Replace specific tags with their content
	patterns := []struct {
		pattern string
		replace string
	}{
		{`\[i\](.*?)\[/i\]`, "$1"},   // Keep italic content
		{`\[c\](.*?)\[/c\]`, "$1"},   // Keep comment content
		{`\[p\](.*?)\[/p\]`, "$1"},   // Keep part of speech content
		{`\[ref\](.*?)\[/ref\]`, ""}, // Remove references completely
		{`\[m\d+\]`, ""},             // Remove opening meaning tag
		{`\[/m\]`, ""},               // Remove closing meaning tag
	}

	result := text
	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		result = re.ReplaceAllString(result, p.replace)
	}

	// Remove any remaining tags
	re := regexp.MustCompile(`\[.*?\]`)
	result = re.ReplaceAllString(result, "")

	// Remove numbered prefixes like "1) ", "2) ", etc.
	re = regexp.MustCompile(`^\d+\)\s*`)
	result = re.ReplaceAllString(result, "")

	// Clean up extra spaces and punctuation
	result = strings.TrimSpace(result)

	// Remove empty parentheses and extra commas
	result = strings.ReplaceAll(result, "()", "")
	result = strings.ReplaceAll(result, "( )", "")
	result = strings.ReplaceAll(result, ", ,", ",")
	result = strings.ReplaceAll(result, " ,", ",")
	result = strings.ReplaceAll(result, ", ", ",")

	// Final trim
	result = strings.Trim(result, ",.;: ")

	// Collapse multiple spaces
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}

	return result
}

// saveCleanedResults saves parsed entries to a text file
func saveCleanedResults(entries []Entry, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Filter out invalid entries (those without Chinese or with empty meanings)
	var validEntries []Entry
	for _, entry := range entries {
		// Only include entries that have Chinese characters and at least one meaning
		if entry.Chinese != "" && len(entry.Meanings) > 0 {
			validEntries = append(validEntries, entry)
		}
	}

	for _, entry := range validEntries {
		// Write Chinese and pinyin
		line := fmt.Sprintf("%s\t%s", entry.Chinese, entry.Pinyin)

		// Add meanings
		if len(entry.Meanings) > 0 {
			// Join meanings and clean up
			meaningsStr := strings.Join(entry.Meanings, " | ")
			// Remove any remaining empty parentheses
			meaningsStr = strings.ReplaceAll(meaningsStr, "()", "")
			meaningsStr = strings.ReplaceAll(meaningsStr, "( )", "")
			meaningsStr = strings.TrimSpace(meaningsStr)

			if meaningsStr != "" {
				line += "\t" + meaningsStr
			}
		}

		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	fmt.Printf("Saved %d valid entries (filtered from %d total)\n", len(validEntries), len(entries))
	return nil
}

// saveAsJSON saves parsed entries as JSON for SQL import
func saveAsJSON(entries []Entry, filename string) error {
	// Filter out invalid entries (those without Chinese or with empty meanings)
	var validEntries []Entry
	for _, entry := range entries {
		// Only include entries that have Chinese characters and at least one meaning
		if entry.Chinese != "" && len(entry.Meanings) > 0 {
			// Create combined meanings string
			combined := strings.Join(entry.Meanings, " | ")
			entry.MeaningsCombined = combined
			validEntries = append(validEntries, entry)
		}
	}

	// Create JSON data
	jsonData, err := json.MarshalIndent(validEntries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Save to file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	fmt.Printf("Saved %d valid entries as JSON\n", len(validEntries))
	return nil
}
