package main

import (
	"fmt"
	"parser/internal/parser"
)

func main() {
	// Конфигурация парсера
	config := parser.Config{
		MaxLines:    1000,                    // Обработать первые 1000 строк (0 = все строки)
		InputFile:   "./dabkrs/dabkrs_1.dsl", // Путь к DSL файлу
		SkipSeeAlso: true,                    // Пропускать ссылки "см."
	}

	fmt.Println("=== Парсинг DSL словаря ===")
	fmt.Printf("Файл: %s\n", config.InputFile)
	fmt.Printf("Максимальное количество строк: %d\n", config.MaxLines)
	fmt.Println()

	// Парсим DSL файл
	entries, err := parser.ParseDSLFile(config)
	if err != nil {
		fmt.Printf("Ошибка при парсинге DSL файла: %v\n", err)
		return
	}

	// Выводим информацию о распарсенных записях
	fmt.Printf("Найдено записей: %d\n", len(entries))
	fmt.Println("\nПервые 10 записей:")
	for i, entry := range entries {
		if i >= 10 {
			fmt.Printf("... и еще %d записей\n", len(entries)-10)
			break
		}
		fmt.Printf("%d. %s [%s]\n", i+1, entry.Chinese, entry.Pinyin)
		for j, meaning := range entry.Meanings {
			fmt.Printf("   %d) %s\n", j+1, meaning)
		}
		fmt.Println()
	}

	// Сохраняем результаты в текстовый файл
	fmt.Println("=== Сохранение результатов ===")
	err = parser.SaveCleanedResults(entries, "./cleaned_results.txt")
	if err != nil {
		fmt.Printf("Ошибка при сохранении в текстовый файл: %v\n", err)
		return
	}
	fmt.Println("Текстовый файл сохранен: cleaned_results.txt")

	// Сохраняем результаты в JSON
	err = parser.SaveAsJSON(entries, "./dictionary.json")
	if err != nil {
		fmt.Printf("Ошибка при сохранении в JSON: %v\n", err)
		return
	}
	fmt.Println("JSON файл сохранен: dictionary.json")

	fmt.Println("\n=== Парсинг завершен успешно ===")
}
