// Package parser содержит функции для сохранения результатов парсинга
package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// SaveCleanedResults сохраняет распарсенные записи в текстовый файл
func SaveCleanedResults(entries []Entry, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("не удалось создать файл: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Фильтруем невалидные записи (без китайских иероглифов или значений)
	var validEntries []Entry
	for _, entry := range entries {
		// Включаем только записи с китайскими иероглифами и хотя бы одним значением
		if entry.Chinese != "" && len(entry.Meanings) > 0 {
			validEntries = append(validEntries, entry)
		}
	}

	// Записываем каждую валидную запись
	for _, entry := range validEntries {
		// Записываем китайские иероглифы и пиньинь
		line := fmt.Sprintf("%s\t%s", entry.Chinese, entry.Pinyin)

		// Добавляем значения
		if len(entry.Meanings) > 0 {
			// Объединяем значения и очищаем
			meaningsStr := strings.Join(entry.Meanings, " | ")
			// Удаляем оставшиеся пустые скобки
			meaningsStr = strings.ReplaceAll(meaningsStr, "()", "")
			meaningsStr = strings.ReplaceAll(meaningsStr, "( )", "")
			meaningsStr = strings.TrimSpace(meaningsStr)

			if meaningsStr != "" {
				line += "\t" + meaningsStr
			}
		}

		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("ошибка записи в файл: %w", err)
		}
	}

	fmt.Printf("Сохранено %d валидных записей (отфильтровано из %d всего)\n", len(validEntries), len(entries))
	return nil
}

// SaveAsJSON сохраняет распарсенные записи в формате JSON
func SaveAsJSON(entries []Entry, filename string) error {
	// Фильтруем невалидные записи (без китайских иероглифов или значений)
	var validEntries []Entry
	for _, entry := range entries {
		// Включаем только записи с китайскими иероглифами и хотя бы одним значением
		if entry.Chinese != "" && len(entry.Meanings) > 0 {
			validEntries = append(validEntries, entry)
		}
	}

	// Создаем JSON данные
	jsonData, err := json.MarshalIndent(validEntries, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка создания JSON: %w", err)
	}

	// Сохраняем в файл
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("не удалось создать файл: %w", err)
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("ошибка записи JSON в файл: %w", err)
	}

	fmt.Printf("Сохранено %d валидных записей в формате JSON\n", len(validEntries))
	return nil
}
