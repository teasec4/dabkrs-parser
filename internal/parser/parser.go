// Package parser содержит функции для парсинга DSL словарей
package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// ParseDSLFile парсит DSL файл и возвращает структурированные записи
func ParseDSLFile(config Config) ([]Entry, error) {
	// Открываем файл
	file, err := os.Open(config.InputFile)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close()

	// Декодер для UTF-16 LE (формат DSL файлов)
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewDecoder()
	reader := transform.NewReader(file, decoder)

	var entries []Entry
	var currentEntry *Entry
	scanner := bufio.NewScanner(reader)
	lineCount := 0

	// Читаем файл построчно
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		// Останавливаемся если достигли максимального количества строк
		if config.MaxLines > 0 && lineCount > config.MaxLines {
			break
		}

		// Пропускаем пустые строки и метаданные
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Пропускаем ссылки "см." если настроено
		if config.SkipSeeAlso && (strings.Contains(line, "см.") || strings.Contains(line, "см.[/p]")) {
			continue
		}

		// Проверяем, начинается ли строка с новой записи (китайские иероглифы без пробелов в начале)
		if isChineseEntryStart(line) && !strings.HasPrefix(line, " ") {
			// Сохраняем предыдущую запись если она существует
			if currentEntry != nil && currentEntry.Chinese != "" {
				entries = append(entries, *currentEntry)
			}

			// Начинаем новую запись
			currentEntry = &Entry{}

			// Парсим первую строку записи (может содержать иероглифы и пиньинь)
			parts := parseEntryFirstLine(line)
			if len(parts) >= 2 {
				currentEntry.Chinese = parts[0]
				currentEntry.Pinyin = parts[1]
			} else if len(parts) == 1 {
				currentEntry.Chinese = parts[0]
			}
		} else if currentEntry != nil {
			// Это продолжение записи (пиньинь или значение)
			if currentEntry.Pinyin == "" && isPinyinLine(line) {
				currentEntry.Pinyin = strings.TrimSpace(line)
			} else {
				// Извлекаем значения из строки
				meanings := extractMeanings(line)
				currentEntry.Meanings = append(currentEntry.Meanings, meanings...)
			}
		}
	}

	// Добавляем последнюю запись
	if currentEntry != nil && currentEntry.Chinese != "" {
		entries = append(entries, *currentEntry)
	}

	if err := scanner.Err(); err != nil {
		return entries, fmt.Errorf("ошибка при чтении файла: %w", err)
	}

	fmt.Printf("Обработано %d строк из DSL файла\n", lineCount)
	return entries, nil
}

// isChineseEntryStart проверяет, начинается ли строка с китайских иероглифов
func isChineseEntryStart(line string) bool {
	// Проверяем наличие китайских иероглифов в строке
	for _, r := range line {
		if (r >= 0x4E00 && r <= 0x9FFF) || // Основные CJK иероглифы
			(r >= 0x3400 && r <= 0x4DBF) || // Расширение A
			(r >= 0x20000 && r <= 0x2A6DF) { // Расширение B
			return true
		}
	}
	return false
}

// isPinyinLine проверяет, является ли строка пиньинем
func isPinyinLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	// Строки с пиньинем обычно не начинаются с квадратных скобок
	if strings.HasPrefix(line, "[") {
		return false
	}

	// Проверяем, содержит ли строка китайские иероглифы (если да, это не пиньинь)
	for _, r := range line {
		if r >= 0x4E00 && r <= 0x9FFF {
			return false
		}
	}

	// Проверяем наличие латинских букв и тоновых знаков пиньиня
	hasLatin := false
	hasPinyinChar := false
	for _, r := range line {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLatin = true
		}
		// Проверяем тоновые знаки и специальные символы пиньиня
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

// parseEntryFirstLine парсит первую строку записи, которая может содержать иероглифы и пиньинь
func parseEntryFirstLine(line string) []string {
	line = strings.TrimSpace(line)
	var parts []string

	// Разделяем по пробелам
	words := strings.Fields(line)
	if len(words) == 0 {
		return parts
	}

	// Первая часть - китайские иероглифы
	parts = append(parts, words[0])

	// Проверяем, есть ли пиньинь в той же строке
	if len(words) > 1 {
		// Объединяем оставшиеся части как потенциальный пиньинь
		potentialPinyin := strings.Join(words[1:], " ")
		if isPinyinLine(potentialPinyin) {
			parts = append(parts, potentialPinyin)
		}
	}

	return parts
}

// extractMeanings извлекает очищенные значения из строки с DSL разметкой
func extractMeanings(line string) []string {
	var meanings []string

	// Регулярное выражение для поиска паттернов [mX]...[/m]
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

	// Если паттерны [mX] не найдены, проверяем всю строку
	if len(meanings) == 0 && strings.Contains(line, "[") {
		// Пытаемся извлечь любой текст в скобках
		cleaned := cleanDSL(line)
		if cleaned != "" {
			meanings = append(meanings, cleaned)
		}
	}

	return meanings
}

// cleanDSL удаляет DSL теги форматирования из текста, сохраняя содержимое
func cleanDSL(text string) string {
	// Сначала сохраняем содержимое внутри тегов перед их удалением
	patterns := []struct {
		pattern string
		replace string
	}{
		{`\[i\](.*?)\[/i\]`, "$1"},   // Сохраняем курсивное содержимое
		{`\[c\](.*?)\[/c\]`, "$1"},   // Сохраняем комментарии
		{`\[p\](.*?)\[/p\]`, "$1"},   // Сохраняем части речи
		{`\[ref\](.*?)\[/ref\]`, ""}, // Удаляем ссылки полностью
		{`\[m\d+\]`, ""},             // Удаляем открывающие теги значений
		{`\[/m\]`, ""},               // Удаляем закрывающие теги значений
	}

	result := text
	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		result = re.ReplaceAllString(result, p.replace)
	}

	// Удаляем оставшиеся теги
	re := regexp.MustCompile(`\[.*?\]`)
	result = re.ReplaceAllString(result, "")

	// Очищаем лишние пробелы и пунктуацию
	result = strings.TrimSpace(result)

	// Удаляем пустые скобки и лишние запятые
	result = strings.ReplaceAll(result, "()", "")
	result = strings.ReplaceAll(result, "( )", "")
	result = strings.ReplaceAll(result, ", ,", ",")
	result = strings.ReplaceAll(result, " ,", ",")
	result = strings.ReplaceAll(result, ", ", ",")

	// Финальная обрезка
	result = strings.Trim(result, ",.;: ")

	// Убираем множественные пробелы
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}

	return result
}
