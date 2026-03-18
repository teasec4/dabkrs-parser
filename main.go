package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Entry represents a parsed dictionary entry
type Entry struct {
	Chinese  string   // Chinese characters
	Pinyin   string   // Pinyin transcription
	Meanings []string // Cleaned meanings/translations
	RawLine  string   // Original line for reference
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

	// Save cleaned results to a new file
	err = saveCleanedResults(entries, "./cleaned_results.txt")
	if err != nil {
		fmt.Printf("Error saving cleaned results: %v\n", err)
		return
	}

	fmt.Println("Cleaned results saved to cleaned_results.txt")
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

		// Check if this line starts a new entry (Chinese characters without leading spaces)
		if isChineseEntryStart(line) && !strings.HasPrefix(line, " ") {
			// Save previous entry if exists
			if currentEntry != nil && currentEntry.Chinese != "" {
				entries = append(entries, *currentEntry)
			}

			// Start new entry
			currentEntry = &Entry{RawLine: line}

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

	// Pinyin lines usually don't start with brackets and contain Latin letters
	if strings.HasPrefix(line, "[") || strings.HasPrefix(line, " ") {
		return false
	}

	// Check for Latin letters with possible tone numbers
	hasLatin := false
	for _, r := range line {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == ' ' || r == '\'' {
			hasLatin = true
		} else if r > 127 && !(r >= 0x4E00 && r <= 0x9FFF) {
			// Non-ASCII but not Chinese - might be pinyin with tone marks
			hasLatin = true
		}
	}

	return hasLatin && !strings.Contains(line, "[m")
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

// cleanDSL removes DSL formatting tags from text
func cleanDSL(text string) string {
	// Remove various DSL tags
	patterns := []string{
		`\[m\d+\]`,           // Opening meaning tag
		`\[/m\]`,             // Closing meaning tag
		`\[p\].*?\[/p\]`,     // Part of speech tags
		`\[c\].*?\[/c\]`,     // Comment tags
		`\[i\].*?\[/i\]`,     // Italic tags
		`\[ref\].*?\[/ref\]`, // Reference tags
		`\[.*?\]`,            // Any other tags
	}

	result := text
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, "")
	}

	// Clean up extra spaces and punctuation
	result = strings.TrimSpace(result)
	result = strings.Trim(result, ",.;:")

	return result
}

// saveCleanedResults saves parsed entries to a file
func saveCleanedResults(entries []Entry, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, entry := range entries {
		// Write Chinese and pinyin
		line := fmt.Sprintf("%s\t%s", entry.Chinese, entry.Pinyin)

		// Add meanings
		if len(entry.Meanings) > 0 {
			line += "\t" + strings.Join(entry.Meanings, " | ")
		}

		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return nil
}
