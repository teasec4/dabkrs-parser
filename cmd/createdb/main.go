package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Entry представляет запись словаря
type Entry struct {
	Chinese          string   `json:"chinese"`
	Pinyin           string   `json:"pinyin"`
	PinyinNormalized string   `json:"pinyin_normalized"`
	Meanings         []string `json:"meanings"`
}

func main() {
	// Параметры
	jsonFile := "./dictionary.json"
	dbFile := "./dictionary.db"
	appendMode := false // Режим добавления к существующей базе

	fmt.Println("=== Создание/обновление оптимизированной базы данных словаря ===")
	fmt.Printf("Исходный файл: %s\n", jsonFile)
	fmt.Printf("Целевая БД: %s\n", dbFile)
	fmt.Println()

	// Проверяем, существует ли база данных
	if _, err := os.Stat(dbFile); err == nil {
		fmt.Printf("База данных %s уже существует\n", dbFile)
		fmt.Print("Добавить данные к существующей базе? (y/n): ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) == "y" {
			appendMode = true
			fmt.Println("Режим: добавление данных к существующей базе")
		} else {
			fmt.Printf("Удаляем старую базу данных: %s\n", dbFile)
			if err := os.Remove(dbFile); err != nil {
				log.Fatalf("Ошибка удаления старой БД: %v", err)
			}
			fmt.Println("Режим: создание новой базы данных")
		}
	} else {
		fmt.Println("Режим: создание новой базы данных")
	}

	// Создаем или обновляем базу данных
	if err := createOptimizedDatabase(jsonFile, dbFile, appendMode); err != nil {
		log.Fatalf("Ошибка создания/обновления базы данных: %v", err)
	}

	// Показываем статистику
	if err := showDatabaseStats(dbFile); err != nil {
		log.Printf("Предупреждение: ошибка получения статистики: %v", err)
	}

	fmt.Println("\n=== Готово! ===")
	fmt.Println("Оптимизированная база данных успешно создана/обновлена.")
	fmt.Println("\nСтруктура базы:")
	fmt.Println("  words    - таблица слов (hanzi, pinyin)")
	fmt.Println("  meanings - таблица значений (word_id, meaning)")
}

// createOptimizedDatabase создает или обновляет оптимизированную базу данных
func createOptimizedDatabase(jsonFile, dbFile string, appendMode bool) error {
	startTime := time.Now()

	// Открываем JSON файл
	fmt.Println("Чтение JSON файла...")
	file, err := os.Open(jsonFile)
	if err != nil {
		return fmt.Errorf("ошибка открытия JSON файла: %w", err)
	}
	defer file.Close()

	// Декодируем JSON потоково
	decoder := json.NewDecoder(file)

	// Читаем открывающую скобку массива
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("ошибка чтения JSON: %w", err)
	}

	// Создаем или открываем базу данных
	if appendMode {
		fmt.Println("Открываем существующую базу данных...")
	} else {
		fmt.Println("Создание новой базы данных...")
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return fmt.Errorf("ошибка открытия базы данных: %w", err)
	}
	defer db.Close()

	// Настраиваем параметры для производительности
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Создаем таблицы если их нет
	if !appendMode {
		fmt.Println("Создаем таблицы...")
		if err := createTables(db); err != nil {
			return fmt.Errorf("ошибка создания таблиц: %w", err)
		}
	} else {
		fmt.Println("Проверяем существующие таблицы...")
		// Проверяем, что таблицы существуют
		if err := checkTablesExist(db); err != nil {
			return fmt.Errorf("ошибка проверки таблиц: %w", err)
		}
	}

	// Начинаем транзакцию
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}

	// Получаем существующие слова если в режиме добавления
	existingWords := make(map[string]int64)
	if appendMode {
		fmt.Println("Загружаем существующие слова...")
		rows, err := db.Query("SELECT id, hanzi, pinyin FROM words")
		if err != nil {
			return fmt.Errorf("ошибка загрузки существующих слов: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id int64
			var hanzi, pinyin string
			if err := rows.Scan(&id, &hanzi, &pinyin); err != nil {
				return fmt.Errorf("ошибка чтения существующего слова: %w", err)
			}
			key := hanzi + "||" + pinyin
			existingWords[key] = id
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("ошибка итерации по существующим словам: %w", err)
		}
		fmt.Printf("Загружено %d существующих слов\n", len(existingWords))
	}

	// Подготавливаем statements
	wordStmt, err := tx.Prepare(`
		INSERT INTO words (hanzi, pinyin)
		VALUES (?, ?)
	`)
	if err != nil {
		return fmt.Errorf("ошибка подготовки statement для слов: %w", err)
	}
	defer wordStmt.Close()

	meaningStmt, err := tx.Prepare(`
		INSERT INTO meanings (word_id, meaning)
		VALUES (?, ?)
	`)
	if err != nil {
		return fmt.Errorf("ошибка подготовки statement для значений: %w", err)
	}
	defer meaningStmt.Close()

	// Обрабатываем записи
	entryCount := 0
	newWordsCount := 0
	existingWordsUsed := 0
	wordMap := make(map[string]int64) // (hanzi + pinyin) -> word_id

	// Копируем существующие слова в wordMap
	for key, id := range existingWords {
		wordMap[key] = id
	}

	lastReport := time.Now()

	fmt.Println("Обработка записей...")

	for decoder.More() {
		var entry Entry
		if err := decoder.Decode(&entry); err != nil {
			return fmt.Errorf("ошибка декодирования записи %d: %w", entryCount, err)
		}

		// Очищаем и нормализуем данные
		hanzi := cleanHanzi(entry.Chinese)
		pinyin := convertPinyinToNumbered(entry.Pinyin)

		// Пропускаем пустые записи
		if hanzi == "" || len(entry.Meanings) == 0 {
			entryCount++
			continue
		}

		// Очищаем значения
		meanings := cleanMeanings(entry.Meanings)
		if len(meanings) == 0 {
			entryCount++
			continue
		}

		// Проверяем, есть ли уже такое слово
		key := hanzi + "||" + pinyin
		var wordID int64

		if existingID, exists := wordMap[key]; exists {
			// Слово уже существует, используем существующий ID
			wordID = existingID
			existingWordsUsed++
		} else {
			// Вставляем новое слово
			result, err := wordStmt.Exec(hanzi, pinyin)
			if err != nil {
				return fmt.Errorf("ошибка вставки слова %d: %w", entryCount, err)
			}

			wordID, err = result.LastInsertId()
			if err != nil {
				return fmt.Errorf("ошибка получения ID слова %d: %w", entryCount, err)
			}

			wordMap[key] = wordID
			newWordsCount++
		}

		// Вставляем значения
		for _, meaning := range meanings {
			if _, err := meaningStmt.Exec(wordID, meaning); err != nil {
				return fmt.Errorf("ошибка вставки значения для слова %d: %w", entryCount, err)
			}
		}

		entryCount++

		// Выводим прогресс
		if entryCount%10000 == 0 || time.Since(lastReport) > 5*time.Second {
			elapsed := time.Since(startTime)
			rate := float64(entryCount) / elapsed.Seconds()
			fmt.Printf("Обработано: %d записей (%.1f записей/сек) [новых слов: %d]\n",
				entryCount, rate, newWordsCount)
			lastReport = time.Now()
		}
	}

	// Читаем закрывающую скобку массива
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("ошибка чтения JSON: %w", err)
	}

	// Завершаем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка коммита транзакции: %w", err)
	}

	// Удаляем короткие значения (только если не в режиме добавления)
	if !appendMode {
		fmt.Println("Удаление коротких значений...")
		if _, err := db.Exec("DELETE FROM meanings WHERE length(meaning) < 2"); err != nil {
			return fmt.Errorf("ошибка удаления коротких значений: %w", err)
		}

		// Удаляем слова без значений
		fmt.Println("Удаление слов без значений...")
		if _, err := db.Exec("DELETE FROM words WHERE id NOT IN (SELECT DISTINCT word_id FROM meanings)"); err != nil {
			return fmt.Errorf("ошибка удаления пустых слов: %w", err)
		}
	}

	// Создаем индексы если их нет
	fmt.Println("Проверка/создание индексов...")
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("ошибка создания индексов: %w", err)
	}

	// Выполняем VACUUM (только если не в режиме добавления для экономии времени)
	if !appendMode {
		fmt.Println("Выполнение VACUUM...")
		if _, err := db.Exec("VACUUM"); err != nil {
			return fmt.Errorf("ошибка выполнения VACUUM: %w", err)
		}
	} else {
		fmt.Println("Пропускаем VACUUM в режиме добавления...")
	}

	// Выводим статистику
	elapsed := time.Since(startTime)
	fmt.Printf("\n=== Статистика %s ===\n", map[bool]string{true: "добавления", false: "создания"}[appendMode])
	fmt.Printf("Всего обработано записей: %d\n", entryCount)
	fmt.Printf("Добавлено новых слов: %d\n", newWordsCount)
	fmt.Printf("Использовано существующих слов: %d\n", existingWordsUsed)
	fmt.Printf("Общее время: %v\n", elapsed.Round(time.Second))
	fmt.Printf("Скорость: %.1f записей/сек\n", float64(entryCount)/elapsed.Seconds())

	return nil
}

// createTables создает оптимизированные таблицы
func createTables(db *sql.DB) error {
	// Таблица слов
	wordsTable := `
	CREATE TABLE words (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hanzi TEXT NOT NULL,
		pinyin TEXT
	)`

	// Таблица значений
	meaningsTable := `
	CREATE TABLE meanings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word_id INTEGER NOT NULL,
		meaning TEXT NOT NULL,
		FOREIGN KEY (word_id) REFERENCES words(id) ON DELETE CASCADE
	)`

	// Создаем таблицы
	for _, tableSQL := range []string{wordsTable, meaningsTable} {
		if _, err := db.Exec(tableSQL); err != nil {
			return fmt.Errorf("ошибка создания таблицы: %w\nSQL: %s", err, tableSQL)
		}
	}

	return nil
}

// checkTablesExist проверяет существование таблиц
func checkTablesExist(db *sql.DB) error {
	tables := []string{"words", "meanings"}

	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='%s'", table)
		if err := db.QueryRow(query).Scan(&count); err != nil {
			return fmt.Errorf("ошибка проверки таблицы %s: %w", table, err)
		}
		if count == 0 {
			return fmt.Errorf("таблица %s не существует", table)
		}
	}

	return nil
}

// createIndexes создает индексы
func createIndexes(db *sql.DB) error {
	indexes := []string{
		"CREATE INDEX idx_words_hanzi ON words(hanzi)",
		"CREATE INDEX idx_words_pinyin ON words(pinyin)",
		"CREATE INDEX idx_meanings_word_id ON meanings(word_id)",
		"CREATE INDEX idx_meanings_meaning ON meanings(meaning)",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("ошибка создания индекса: %w\nSQL: %s", err, indexSQL)
		}
	}

	return nil
}

// cleanHanzi очищает китайские иероглифы
func cleanHanzi(hanzi string) string {
	if hanzi == "" {
		return ""
	}

	// Убираем пробелы по краям
	hanzi = strings.TrimSpace(hanzi)

	// Убираем двойные пробелы
	re := regexp.MustCompile(`\s+`)
	hanzi = re.ReplaceAllString(hanzi, " ")

	return hanzi
}

// convertPinyinToNumbered конвертирует пиньинь в числовой формат (shang4 hai3)
func convertPinyinToNumbered(pinyin string) string {
	if pinyin == "" {
		return ""
	}

	// Карта тоновых символов
	tonesMap := map[rune]string{
		'ā': "a1", 'á': "a2", 'ǎ': "a3", 'à': "a4",
		'ē': "e1", 'é': "e2", 'ě': "e3", 'è': "e4",
		'ī': "i1", 'í': "i2", 'ǐ': "i3", 'ì': "i4",
		'ō': "o1", 'ó': "o2", 'ǒ': "o3", 'ò': "o4",
		'ū': "u1", 'ú': "u2", 'ǔ': "u3", 'ù': "u4",
		'ǖ': "ü1", 'ǘ': "ü2", 'ǚ': "ü3", 'ǜ': "ü4",
	}

	// Заменяем тоновые символы
	var result strings.Builder
	for _, r := range pinyin {
		if replacement, ok := tonesMap[r]; ok {
			result.WriteString(replacement)
		} else {
			result.WriteRune(r)
		}
	}

	converted := result.String()

	// Убираем лишние пробелы и апострофы
	converted = strings.ReplaceAll(converted, "'", "")
	converted = strings.ReplaceAll(converted, "’", "")

	// Убираем двойные пробелы
	re := regexp.MustCompile(`\s+`)
	converted = re.ReplaceAllString(converted, " ")

	// Приводим к нижнему регистру
	converted = strings.ToLower(converted)

	return strings.TrimSpace(converted)
}

// cleanMeanings очищает значения
func cleanMeanings(meanings []string) []string {
	var cleaned []string
	seen := make(map[string]bool)

	for _, meaning := range meanings {
		// Очищаем значение
		meaning = cleanMeaning(meaning)

		// Пропускаем пустые значения
		if meaning == "" {
			continue
		}

		// Пропускаем дубликаты
		if seen[meaning] {
			continue
		}

		seen[meaning] = true
		cleaned = append(cleaned, meaning)
	}

	return cleaned
}

// cleanMeaning очищает одно значение
func cleanMeaning(meaning string) string {
	// Убираем пробелы по краям
	meaning = strings.TrimSpace(meaning)

	// Убираем двойные пробелы
	re := regexp.MustCompile(`\s+`)
	meaning = re.ReplaceAllString(meaning, " ")

	// Убираем пустые скобки
	meaning = strings.ReplaceAll(meaning, "()", "")
	meaning = strings.ReplaceAll(meaning, "( )", "")

	// Убираем лишние точки с запятой
	meaning = strings.ReplaceAll(meaning, "; ;", ";")
	meaning = strings.ReplaceAll(meaning, ";;", ";")
	meaning = strings.Trim(meaning, ";")

	// Убираем лишние запятые
	meaning = strings.ReplaceAll(meaning, ",,", ",")
	meaning = strings.Trim(meaning, ",")

	// Убираем лишние точки
	meaning = strings.ReplaceAll(meaning, "..", ".")
	meaning = strings.Trim(meaning, ".")

	return meaning
}

// showDatabaseStats показывает статистику базы данных
func showDatabaseStats(dbFile string) error {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return fmt.Errorf("ошибка открытия базы данных: %w", err)
	}
	defer db.Close()

	var wordCount, meaningCount int

	// Количество слов
	if err := db.QueryRow("SELECT COUNT(*) FROM words").Scan(&wordCount); err != nil {
		return fmt.Errorf("ошибка подсчета слов: %w", err)
	}

	// Количество значений
	if err := db.QueryRow("SELECT COUNT(*) FROM meanings").Scan(&meaningCount); err != nil {
		return fmt.Errorf("ошибка подсчета значений: %w", err)
	}

	// Среднее количество значений на слово
	avgMeanings := 0.0
	if wordCount > 0 {
		avgMeanings = float64(meaningCount) / float64(wordCount)
	}

	fmt.Printf("\n=== Статистика базы данных ===\n")
	fmt.Printf("Количество слов: %d\n", wordCount)
	fmt.Printf("Количество значений: %d\n", meaningCount)
	fmt.Printf("Среднее значений на слово: %.2f\n", avgMeanings)

	// Примеры записей
	fmt.Printf("\nПримеры записей:\n")
	rows, err := db.Query(`
		SELECT w.hanzi, w.pinyin, m.meaning
		FROM words w
		JOIN meanings m ON w.id = m.word_id
		ORDER BY w.id
		LIMIT 5
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var hanzi, pinyin, meaning string
			rows.Scan(&hanzi, &pinyin, &meaning)
			fmt.Printf("  %s [%s] - %s\n", hanzi, pinyin, meaning)
		}
	}

	return nil
}
