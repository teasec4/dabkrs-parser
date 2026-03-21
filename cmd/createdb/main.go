package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	
	_ "github.com/mattn/go-sqlite3"
)

type Entry struct {
	Word    string   `json:"Word"`
	Pinyin  string   `json:"Pinyin"`
	Mean    []string `json:"Mean"`
}

func main() {
	// Параметры
	jsonFile := "./output.json"
	dbFile := "./dictionary_test_lite.db"
	
	// открыть БД
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	
	// включить FK
	db.Exec("PRAGMA foreign_keys = ON")
	
	// создать таблицы
	if err := createTables(db); err != nil {
		panic(err)
	}
	
	// индексы
	if err := createIndexes(db); err != nil {
		panic(err)
	}
	
	// загрузить JSON
	entries, err := loadJSON(jsonFile)
	if err != nil {
			panic(err)
	}
	
	fmt.Printf("Загружено записей: %d\n", len(entries))
	
	// вставка
	if err := insertData(db, entries); err != nil {
			panic(err)
	}

	fmt.Println("=== Создание базы данных словаря ===")
	fmt.Printf("Исходный файл: %s\n", jsonFile)
	fmt.Printf("Целевая БД: %s\n", dbFile)
	fmt.Println()
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

// load json
func loadJSON(path string) ([]Entry, error){
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var entries []Entry
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&entries); err != nil {
		return nil, err
	}
	
	return entries, nil
}

func insertData(db *sql.DB, entries []Entry) error{
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	
	wordStmt, err := tx.Prepare("INSERT INTO words(hanzi, pinyin) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer wordStmt.Close()
	
	meanStmt, err := tx.Prepare("INSERT INTO meanings(word_id, meaning) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer meanStmt.Close()
	
	for _, e := range entries {
		// пропускаем пустые
		if e.Word == "" {
			continue
		}

		res, err := wordStmt.Exec(e.Word, e.Pinyin)
		if err != nil {
			return err
		}

		wordID, _ := res.LastInsertId()

		for _, m := range e.Mean {
			if m == "" {
				continue
			}
			_, err := meanStmt.Exec(wordID, m)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()

}


// createIndexes создает индексы
func createIndexes(db *sql.DB) error {
	indexes := []string{
		"CREATE INDEX idx_words_hanzi ON words(hanzi)",
		"CREATE INDEX idx_words_pinyin ON words(pinyin)",
		"CREATE INDEX idx_meanings_word_id ON meanings(word_id)",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("ошибка создания индекса: %w\nSQL: %s", err, indexSQL)
		}
	}

	return nil
}

