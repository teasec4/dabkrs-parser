// Package parser содержит функции для сохранения результатов парсинга
package parser

import (
	"encoding/json"
	"fmt"
	"os"
)

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

	fmt.Printf("Сохранено %d валидных записей (отфильтровано из %d всего) в формате JSON\n", len(validEntries), len(entries))
	return nil
}
