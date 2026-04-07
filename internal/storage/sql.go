package storage

import (
	"database/sql"
	"fmt"
	"os"
	"parser/internal/parser"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

func NewDB(path string) (*DB, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := initDB(path); err != nil {
			return nil, fmt.Errorf("init db: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", path+"?_busy_timeout=60000&_synchronous=OFF&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(1)

	return &DB{db: db}, nil
}

func initDB(path string) error {
	schema := `
CREATE TABLE IF NOT EXISTS entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    headword TEXT NOT NULL UNIQUE,
    pinyin TEXT,
    pinyin_normalized TEXT,
    frequency INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_entries_headword ON entries(headword);
CREATE INDEX IF NOT EXISTS idx_entries_frequency ON entries(frequency DESC);
CREATE INDEX IF NOT EXISTS idx_entries_pinyin_norm ON entries(pinyin_normalized);

CREATE TABLE IF NOT EXISTS meanings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    level INTEGER DEFAULT 0,
    text TEXT NOT NULL,
    order_num INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_meanings_entry ON meanings(entry_id);
CREATE INDEX IF NOT EXISTS idx_meanings_level ON meanings(level);

CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    meaning_id INTEGER NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    value TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tags_meaning ON tags(meaning_id);
CREATE INDEX IF NOT EXISTS idx_tags_type ON tags(type);
`

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(schema)
	if err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}

	return nil
}

func (s *DB) Close() error {
	return s.db.Close()
}

func (s *DB) InsertEntry(entry *parser.Entry) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var entryID int64
	err = tx.QueryRow(`
		INSERT INTO entries (headword, pinyin, pinyin_normalized)
		VALUES (?, ?, ?)
		ON CONFLICT(headword) DO UPDATE SET
			pinyin = excluded.pinyin,
			pinyin_normalized = excluded.pinyin_normalized
		RETURNING id
	`, entry.Headword, entry.Pinyin, entry.PinyinNormalized).Scan(&entryID)
	if err != nil {
		return 0, fmt.Errorf("insert entry: %w", err)
	}

	for i, m := range entry.Meanings {
		var meaningID int64
		err = tx.QueryRow(`
			INSERT INTO meanings (entry_id, level, text, order_num)
			VALUES (?, ?, ?, ?)
			RETURNING id
		`, entryID, m.Level, m.Text, i).Scan(&meaningID)
		if err != nil {
			return 0, fmt.Errorf("insert meaning: %w", err)
		}

		for _, tag := range m.Tags {
			_, err = tx.Exec(`
				INSERT INTO tags (meaning_id, type, value)
				VALUES (?, ?, ?)
			`, meaningID, tag.Type, tag.Value)
			if err != nil {
				return 0, fmt.Errorf("insert tag: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return entryID, nil
}

func (s *DB) InsertEntriesBatch(entries []parser.Entry, batchSize int) (int, error) {
	if batchSize <= 0 {
		batchSize = 5000
	}

	_, err := s.db.Exec("PRAGMA synchronous = OFF")
	if err != nil {
		return 0, err
	}
	_, err = s.db.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		return 0, err
	}

	totalInserted := 0
	totalBatches := (len(entries) + batchSize - 1) / batchSize

	for batchNum := 0; batchNum < totalBatches; batchNum++ {
		start := batchNum * batchSize
		end := start + batchSize
		if end > len(entries) {
			end = len(entries)
		}
		batch := entries[start:end]

		tx, err := s.db.Begin()
		if err != nil {
			return totalInserted, fmt.Errorf("begin tx: %w", err)
		}

		entryIDs := make(map[string]int)

		for _, entry := range batch {
			var entryID int64
			err := tx.QueryRow(`
				INSERT INTO entries (headword, pinyin, pinyin_normalized)
				VALUES (?, ?, ?)
				ON CONFLICT(headword) DO UPDATE SET
					pinyin = excluded.pinyin,
					pinyin_normalized = excluded.pinyin_normalized
				RETURNING id
			`, entry.Headword, entry.Pinyin, entry.PinyinNormalized).Scan(&entryID)

			if err != nil {
				tx.Rollback()
				return totalInserted, fmt.Errorf("insert entry (%s): %w", entry.Headword, err)
			}

			entryIDs[entry.Headword] = int(entryID)
			totalInserted++

			for i, m := range entry.Meanings {
				result, err := tx.Exec(`
					INSERT INTO meanings (entry_id, level, text, order_num)
					VALUES (?, ?, ?, ?)
				`, entryID, m.Level, m.Text, i)
				if err != nil {
					tx.Rollback()
					return totalInserted, fmt.Errorf("insert meaning: %w", err)
				}

				meaningID, _ := result.LastInsertId()

				for _, tag := range m.Tags {
					_, err = tx.Exec(`
						INSERT INTO tags (meaning_id, type, value)
						VALUES (?, ?, ?)
					`, meaningID, tag.Type, tag.Value)
					if err != nil {
						tx.Rollback()
						return totalInserted, fmt.Errorf("insert tag: %w", err)
					}
				}
			}
		}

		if err = tx.Commit(); err != nil {
			return totalInserted, fmt.Errorf("commit: %w", err)
		}

		fmt.Printf("Batch %d/%d: inserted %d entries (total: %d)\n", batchNum+1, totalBatches, len(batch), totalInserted)
	}

	return totalInserted, nil
}

func (s *DB) InsertEntries(entries []parser.Entry, batchSize int) (int, error) {
	return s.InsertEntriesBatch(entries, batchSize)
}

func (s *DB) GetEntryByHeadword(headword string) (*parser.Entry, error) {
	var entry parser.Entry
	var entryID int64

	err := s.db.QueryRow(`
		SELECT id, headword, pinyin
		FROM entries WHERE headword = ?
	`, headword).Scan(&entryID, &entry.Headword, &entry.Pinyin)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT id, level, text, order_num
		FROM meanings WHERE entry_id = ? ORDER BY order_num
	`, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m parser.Meaning
		var meaningID int64
		if err := rows.Scan(&meaningID, &m.Level, &m.Text, &m.Order); err != nil {
			return nil, err
		}

		tagRows, err := s.db.Query(`
			SELECT type, value FROM tags WHERE meaning_id = ?
		`, meaningID)
		if err != nil {
			return nil, err
		}
		for tagRows.Next() {
			var tag parser.Tag
			if err := tagRows.Scan(&tag.Type, &tag.Value); err != nil {
				tagRows.Close()
				return nil, err
			}
			m.Tags = append(m.Tags, tag)
		}
		tagRows.Close()

		entry.Meanings = append(entry.Meanings, m)
	}

	return &entry, nil
}

func (s *DB) SearchByHeadword(prefix string, limit int) ([]parser.Entry, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.Query(`
		SELECT id, headword, pinyin
		FROM entries WHERE headword LIKE ?
		LIMIT ?
	`, prefix+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []parser.Entry
	for rows.Next() {
		var e parser.Entry
		if err := rows.Scan(&e.Headword, &e.Pinyin); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

func (s *DB) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM entries").Scan(&count)
	return count, err
}

func (s *DB) Search(prefix string, limit int) ([]parser.Entry, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.Query(`
		SELECT id, headword, pinyin
		FROM entries
		WHERE headword LIKE ?
		ORDER BY headword
		LIMIT ?
	`, prefix+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []parser.Entry
	for rows.Next() {
		var e parser.Entry
		var id int64
		if err := rows.Scan(&id, &e.Headword, &e.Pinyin); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

func (s *DB) SearchByPinyin(pinyin string, limit int) ([]parser.Entry, error) {
	if limit <= 0 {
		limit = 20
	}

	normalized := strings.ToLower(pinyin)
	normalized = strings.NewReplacer(
		"ā", "a", "á", "a", "ǎ", "a", "à", "a",
		"ē", "e", "é", "e", "ě", "e", "è", "e",
		"ī", "i", "í", "i", "ǐ", "i", "ì", "i",
		"ō", "o", "ó", "o", "ǒ", "o", "ò", "o",
		"ū", "u", "ú", "u", "ǔ", "u", "ù", "u",
		"ǖ", "v", "ǘ", "v", "ǚ", "v", "ǜ", "v",
	).Replace(normalized)

	rows, err := s.db.Query(`
		SELECT id, headword, pinyin
		FROM entries
		WHERE LOWER(pinyin) LIKE ?
		ORDER BY pinyin
		LIMIT ?
	`, normalized+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []parser.Entry
	for rows.Next() {
		var e parser.Entry
		var id int64
		if err := rows.Scan(&id, &e.Headword, &e.Pinyin); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}
