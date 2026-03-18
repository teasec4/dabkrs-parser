package main

import (
	"fmt"
	"parser/internal/parser"
)

func main() {
	// Конфигурация парсера
	config := parser.Config{
		MaxLines: 0, // 0 = обработать все строки
		InputFiles: []string{
			"./dabkrs/dabkrs_1.dsl",
			"./dabkrs/dabkrs_2.dsl",
			"./dabkrs/dabkrs_3.dsl",
		},
		SkipSeeAlso: true, // Пропускать ссылки "см."
	}

	fmt.Println("=== Парсинг DSL словаря ===")
	fmt.Printf("Файлы: %v\n", config.InputFiles)
	fmt.Printf("Максимальное количество строк на файл: %d (0 = все)\n", config.MaxLines)
	fmt.Println()

	// Парсим DSL файлы
	entries, err := parser.ParseDSLFiles(config)
	if err != nil {
		fmt.Printf("Ошибка при парсинге DSL файлов: %v\n", err)
		return
	}

	// Выводим информацию о распарсенных записях
	fmt.Printf("Найдено записей: %d\n", len(entries))
	fmt.Println("\nПервые 5 записей с нормализованным пиньинем:")
	for i, entry := range entries {
		if i >= 5 {
			fmt.Printf("... и еще %d записей\n", len(entries)-5)
			break
		}
		fmt.Printf("%d. %s [%s] -> нормализованный: %s\n", i+1, entry.Chinese, entry.Pinyin, entry.PinyinNormalized)
		for j, meaning := range entry.Meanings {
			fmt.Printf("   %d) %s\n", j+1, meaning)
		}
		fmt.Println()
	}

	// Сохраняем результаты в JSON
	fmt.Println("=== Сохранение результатов ===")
	err = parser.SaveAsJSON(entries, "./dictionary.json")
	if err != nil {
		fmt.Printf("Ошибка при сохранении в JSON: %v\n", err)
		return
	}
	fmt.Println("JSON файл сохранен: dictionary.json")

	fmt.Println("\n=== Парсинг завершен успешно ===")
	fmt.Println("Теперь вы можете создать базу данных командой: go run cmd/createdb/main.go")
}
