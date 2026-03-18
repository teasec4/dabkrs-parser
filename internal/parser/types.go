// Package parser содержит типы данных и функции для парсинга DSL словарей
package parser

// Entry представляет запись в словаре
type Entry struct {
	Chinese  string   `json:"chinese"`  // Китайские иероглифы
	Pinyin   string   `json:"pinyin"`   // Транскрипция пиньинь
	Meanings []string `json:"meanings"` // Очищенные переводы/значения
}

// Config содержит конфигурацию для парсера
type Config struct {
	MaxLines    int    // Максимальное количество строк для обработки (0 = все строки)
	InputFile   string // Путь к входному DSL файлу
	SkipSeeAlso bool   // Пропускать ли ссылки "см."
}
