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

// normalizePinyin удаляет тоны из пиньиня и приводит к нижнему регистру
func normalizePinyin(pinyin string) string {
	if pinyin == "" {
		return ""
	}

	// Создаем карту для замены тоновых символов
	tonesMap := map[rune]rune{
		'ā': 'a', 'á': 'a', 'ǎ': 'a', 'à': 'a',
		'ē': 'e', 'é': 'e', 'ě': 'e', 'è': 'e',
		'ī': 'i', 'í': 'i', 'ǐ': 'i', 'ì': 'i',
		'ō': 'o', 'ó': 'o', 'ǒ': 'o', 'ò': 'o',
		'ū': 'u', 'ú': 'u', 'ǔ': 'u', 'ù': 'u',
		'ǖ': 'ü', 'ǘ': 'ü', 'ǚ': 'ü', 'ǜ': 'ü',
		'Ā': 'A', 'Á': 'A', 'Ǎ': 'A', 'À': 'A',
		'Ē': 'E', 'É': 'E', 'Ě': 'E', 'È': 'E',
		'Ī': 'I', 'Í': 'I', 'Ǐ': 'I', 'Ì': 'I',
		'Ō': 'O', 'Ó': 'O', 'Ǒ': 'O', 'Ò': 'O',
		'Ū': 'U', 'Ú': 'U', 'Ǔ': 'U', 'Ù': 'U',
		'Ǖ': 'Ü', 'Ǘ': 'Ü', 'Ǚ': 'Ü', 'Ǜ': 'Ü',
	}

	// Заменяем тоновые символы
	var result strings.Builder
	for _, r := range pinyin {
		if replacement, ok := tonesMap[r]; ok {
			result.WriteRune(replacement)
		} else {
			result.WriteRune(r)
		}
	}

	normalized := result.String()

	// Удаляем пробелы и апострофы, которые иногда встречаются в пиньине
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "'", "")
	normalized = strings.ReplaceAll(normalized, "’", "")

	// Приводим к нижнему регистру
	normalized = strings.ToLower(normalized)

	return normalized
}

// ParseDSLFiles парсит DSL файлы и возвращает структурированные записи
func ParseDSLFiles(config Config) ([]Entry, error) {
	var allEntries []Entry

	// Обрабатываем каждый файл
	for _, inputFile := range config.InputFiles {
		entries, err := parseSingleDSLFile(inputFile, config.MaxLines, config.SkipSeeAlso)
		if err != nil {
			return nil, fmt.Errorf("ошибка при парсинге файла %s: %w", inputFile, err)
		}
		allEntries = append(allEntries, entries...)
	}

	return allEntries, nil
}

// parseSingleDSLFile парсит один DSL файл
func parseSingleDSLFile(inputFile string, maxLines int, skipSeeAlso bool) ([]Entry, error) {
	// Открываем файл
	file, err := os.Open(inputFile)
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
		if maxLines > 0 && lineCount > maxLines {
			break
		}

		// Пропускаем пустые строки и метаданные
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Пропускаем ссылки "см." если настроено
		if skipSeeAlso && (strings.Contains(line, "см.") || strings.Contains(line, "см.[/p]")) {
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
				// Нормализуем пиньинь
				currentEntry.PinyinNormalized = normalizePinyin(currentEntry.Pinyin)
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

	fmt.Printf("Обработано %d строк из файла %s\n", lineCount, inputFile)
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
