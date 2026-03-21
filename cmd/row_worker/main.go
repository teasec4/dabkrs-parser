package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Entry struct {
	Word   string   `json:"word"`
	Pinyin string   `json:"pinyin"`
	Mean   []string `json:"mean"`
}

func main() {
	// Открываем файл output.txt
	file, err := os.Open("output.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var entries []Entry
	var current Entry
	var collectingMeaning bool
	var meaningBuffer strings.Builder

	// Регулярное выражение для удаления тегов
	tagRe := regexp.MustCompile(`\[[^\[\]]*?\]`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Пропускаем пустые строки
		if line == "" {
			continue
		}

		// Если строка начинается с китайского иероглифа и не содержит латинских букв
		// Это новое слово
		if isChineseWord(line) && current.Word == "" {
			current.Word = line
			continue
		}

		// Если строка содержит латинские буквы (пиньинь) и у нас уже есть слово
		if containsLatin(line) && current.Word != "" && current.Pinyin == "" {
			current.Pinyin = line
			continue
		}

		// Если строка содержит [m - это начало значения
		if strings.Contains(line, "[m") {
			collectingMeaning = true
			// Удаляем тег [mX] из начала строки
			line = regexp.MustCompile(`^\[m\d+\]`).ReplaceAllString(line, "")
		}

		// Если строка содержит [/m] - это конец значения
		if strings.Contains(line, "[/m]") {
			collectingMeaning = false
			// Удаляем тег [/m] из конца строки
			line = strings.ReplaceAll(line, "[/m]", "")

			if meaningBuffer.Len() > 0 {
				meaningBuffer.WriteString(" ")
			}
			meaningBuffer.WriteString(line)

			// Очищаем текст от всех тегов
			cleanedText := tagRe.ReplaceAllString(meaningBuffer.String(), "")
			cleanedText = strings.TrimSpace(cleanedText)

			if cleanedText != "" {
				// Разделяем значения, если они слиплись (например, "1)Турция2)Туркменистан")
				values := splitValues(cleanedText)
				current.Mean = append(current.Mean, values...)
			}

			meaningBuffer.Reset()

			// Если это было inline значение (всё в одной строке), сохраняем запись
			if !strings.Contains(line, "[m") {
				if current.Word != "" && current.Pinyin != "" && len(current.Mean) > 0 {
					entries = append(entries, current)
					current = Entry{}
				}
			}
			continue
		}

		// Если мы собираем значение
		if collectingMeaning {
			if meaningBuffer.Len() > 0 {
				meaningBuffer.WriteString(" ")
			}
			meaningBuffer.WriteString(line)
			continue
		}

		// Если строка содержит и [m и [/m] в одной строке (inline значение)
		if strings.Contains(line, "[m") && strings.Contains(line, "[/m]") {
			// Извлекаем все значения из строки
			re := regexp.MustCompile(`\[m\d+\](.*?)\[/m\]`)
			matches := re.FindAllStringSubmatch(line, -1)

			for _, match := range matches {
				if len(match) > 1 {
					cleanedText := tagRe.ReplaceAllString(match[1], "")
					cleanedText = strings.TrimSpace(cleanedText)
					if cleanedText != "" {
						values := splitValues(cleanedText)
						current.Mean = append(current.Mean, values...)
					}
				}
			}

			// Сохраняем запись
			if current.Word != "" && current.Pinyin != "" && len(current.Mean) > 0 {
				entries = append(entries, current)
				current = Entry{}
			}
			continue
		}

		// Если у нас есть полная запись и встречаем новое китайское слово
		if isChineseWord(line) && current.Word != "" && current.Pinyin != "" {
			if len(current.Mean) > 0 {
				entries = append(entries, current)
			}
			current = Entry{Word: line}
			collectingMeaning = false
			meaningBuffer.Reset()
		}
	}

	// Добавляем последнюю запись
	if current.Word != "" && current.Pinyin != "" && len(current.Mean) > 0 {
		entries = append(entries, current)
	}

	// Сохраняем в JSON
	jsonFile, err := os.Create("output.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(entries); err != nil {
		panic(err)
	}

	fmt.Printf("Успешно обработано %d записей\n", len(entries))
}

func isChineseWord(s string) bool {
	// Китайские иероглифы в диапазоне Unicode
	for _, r := range s {
		if (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
			(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
			(r >= 0x20000 && r <= 0x2A6DF) { // CJK Extension B
			return true
		}
	}
	return false
}

func containsLatin(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

func splitValues(s string) []string {
	var result []string

	// Пытаемся разделить по шаблону "X) текст"
	re := regexp.MustCompile(`(\d+\)[^0-9]+)`)
	matches := re.FindAllString(s, -1)

	if len(matches) > 0 {
		for _, match := range matches {
			// Добавляем пробел после скобки, если его нет
			match = regexp.MustCompile(`(\d+\))([^ ])`).ReplaceAllString(match, "$1 $2")
			match = strings.TrimSpace(match)
			if match != "" {
				result = append(result, match)
			}
		}
	} else {
		// Если не нашли разделенных значений, возвращаем как одно
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}

	return result
}
