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

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	return &DB{db: db}, nil
}

func initDB(path string) error {
	schema := `
CREATE TABLE IF NOT EXISTS entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hanzi TEXT NOT NULL UNIQUE,
    pinyin TEXT,
    pinyin_normalized TEXT
);
CREATE INDEX IF NOT EXISTS idx_entries_hanzi ON entries(hanzi);
CREATE INDEX IF NOT EXISTS idx_entries_pinyin_norm ON entries(pinyin_normalized);

CREATE TABLE IF NOT EXISTS meanings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    part_of_speech TEXT,
    order_num INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_meanings_entry ON meanings(entry_id);

CREATE TABLE IF NOT EXISTS refs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    meaning_id INTEGER NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    target_entry_id INTEGER REFERENCES entries(id) ON DELETE SET NULL,
    target_hanzi TEXT
);
CREATE INDEX IF NOT EXISTS idx_refs_meaning ON refs(meaning_id);
CREATE INDEX IF NOT EXISTS idx_refs_target ON refs(target_entry_id);

CREATE TABLE IF NOT EXISTS examples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    meaning_id INTEGER NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    chinese TEXT NOT NULL,
    translation TEXT
);
CREATE INDEX IF NOT EXISTS idx_examples_meaning ON examples(meaning_id);
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
		INSERT INTO entries (hanzi, pinyin, pinyin_normalized)
		VALUES (?, ?, ?)
		ON CONFLICT(hanzi) DO UPDATE SET
			pinyin = excluded.pinyin,
			pinyin_normalized = excluded.pinyin_normalized
		RETURNING id
	`, entry.Hanzi, entry.Pinyin, entry.PinyinNormalized).Scan(&entryID)
	if err != nil {
		return 0, fmt.Errorf("insert entry: %w", err)
	}

	for i, m := range entry.Meanings {
		var meaningID int64
		err = tx.QueryRow(`
			INSERT INTO meanings (entry_id, text, part_of_speech, order_num)
			VALUES (?, ?, ?, ?)
			RETURNING id
		`, entryID, m.Text, m.PartOfSpeech, i).Scan(&meaningID)
		if err != nil {
			return 0, fmt.Errorf("insert meaning: %w", err)
		}

		for _, ref := range m.Refs {
			_, err = tx.Exec(`
				INSERT INTO refs (meaning_id, target_hanzi)
				VALUES (?, ?)
			`, meaningID, ref)
			if err != nil {
				return 0, fmt.Errorf("insert ref: %w", err)
			}
		}

		for _, ex := range m.Examples {
			chinese, translation := splitExample(ex)
			_, err = tx.Exec(`
				INSERT INTO examples (meaning_id, chinese, translation)
				VALUES (?, ?, ?)
			`, meaningID, chinese, translation)
			if err != nil {
				return 0, fmt.Errorf("insert example: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return entryID, nil
}

func (s *DB) InsertEntries(entries []parser.Entry, batchSize int) (int, error) {
	inserted := 0
	for i, entry := range entries {
		if _, err := s.InsertEntry(&entry); err != nil {
			return inserted, fmt.Errorf("insert entry %d (%s): %w", i, entry.Hanzi, err)
		}
		inserted++

		if batchSize > 0 && inserted%batchSize == 0 {
			fmt.Printf("Inserted %d entries...\n", inserted)
		}
	}
	return inserted, nil
}

func (s *DB) ResolveRefs() error {
	_, err := s.db.Exec(`
		UPDATE refs
		SET target_entry_id = (
			SELECT id FROM entries WHERE entries.hanzi = refs.target_hanzi
		)
		WHERE target_entry_id IS NULL AND target_hanzi IS NOT NULL
	`)
	return err
}

func (s *DB) GetEntryByHanzi(hanzi string) (*parser.Entry, error) {
	var entry parser.Entry
	var entryID int64

	err := s.db.QueryRow(`
		SELECT id, hanzi, pinyin, pinyin_normalized
		FROM entries WHERE hanzi = ?
	`, hanzi).Scan(&entryID, &entry.Hanzi, &entry.Pinyin, &entry.PinyinNormalized)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT id, text, part_of_speech, order_num
		FROM meanings WHERE entry_id = ? ORDER BY order_num
	`, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m parser.Meaning
		var meaningID int64
		if err := rows.Scan(&meaningID, &m.Text, &m.PartOfSpeech, &m.Order); err != nil {
			return nil, err
		}

		refRows, err := s.db.Query(`SELECT target_hanzi FROM refs WHERE meaning_id = ?`, meaningID)
		if err != nil {
			return nil, err
		}
		for refRows.Next() {
			var ref string
			if err := refRows.Scan(&ref); err != nil {
				refRows.Close()
				return nil, err
			}
			m.Refs = append(m.Refs, ref)
		}
		refRows.Close()

		exRows, err := s.db.Query(`SELECT chinese, translation FROM examples WHERE meaning_id = ?`, meaningID)
		if err != nil {
			return nil, err
		}
		for exRows.Next() {
			var ex, trans string
			if err := exRows.Scan(&ex, &trans); err != nil {
				exRows.Close()
				return nil, err
			}
			m.Examples = append(m.Examples, ex+"|"+trans)
		}
		exRows.Close()

		entry.Meanings = append(entry.Meanings, m)
	}

	return &entry, nil
}

func (s *DB) SearchByHanzi(prefix string, limit int) ([]parser.Entry, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.Query(`
		SELECT id, hanzi, pinyin, pinyin_normalized
		FROM entries WHERE hanzi LIKE ?
		LIMIT ?
	`, prefix+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []parser.Entry
	for rows.Next() {
		var e parser.Entry
		if err := rows.Scan(&e.Hanzi, &e.Pinyin, &e.PinyinNormalized); err != nil {
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
		SELECT id, hanzi, pinyin, pinyin_normalized
		FROM entries 
		WHERE hanzi LIKE ? OR pinyin_normalized LIKE ?
		ORDER BY hanzi
		LIMIT ?
	`, prefix+"%", strings.ReplaceAll(prefix, " ", "%")+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []parser.Entry
	for rows.Next() {
		var e parser.Entry
		var id int64
		if err := rows.Scan(&id, &e.Hanzi, &e.Pinyin, &e.PinyinNormalized); err != nil {
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
	rows, err := s.db.Query(`
		SELECT id, hanzi, pinyin, pinyin_normalized
		FROM entries 
		WHERE pinyin_normalized LIKE ?
		ORDER BY pinyin_normalized
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
		if err := rows.Scan(&id, &e.Hanzi, &e.Pinyin, &e.PinyinNormalized); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

func splitExample(ex string) (chinese, translation string) {
	if idx := indexByte(ex, '|'); idx >= 0 {
		return ex[:idx], ex[idx+1:]
	}
	return ex, ""
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
