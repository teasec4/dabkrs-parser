-- DSL Dictionary Schema
-- Chinese-Russian Bilingual Dictionary (БКРС)

-- entries (китайские слова/символы)
CREATE TABLE IF NOT EXISTS entries (
    id SERIAL PRIMARY KEY,
    headword TEXT NOT NULL UNIQUE,
    pinyin TEXT,
    pinyin_normalized TEXT
);
CREATE INDEX IF NOT EXISTS idx_entries_headword ON entries(headword);
CREATE INDEX IF NOT EXISTS idx_entries_pinyin_norm ON entries(pinyin_normalized);

-- meanings (значения слова)
CREATE TABLE IF NOT EXISTS meanings (
    id SERIAL PRIMARY KEY,
    entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    level INTEGER DEFAULT 0,
    text TEXT NOT NULL,
    order_num INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_meanings_entry ON meanings(entry_id);

-- examples (примеры)
CREATE TABLE IF NOT EXISTS examples (
    id SERIAL PRIMARY KEY,
    meaning_id INTEGER NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    text TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_examples_meaning ON examples(meaning_id);

-- references (перекрёстные ссылки)
CREATE TABLE IF NOT EXISTS "references" (
    id SERIAL PRIMARY KEY,
    meaning_id INTEGER NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    target_headword TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_references_meaning ON "references"(meaning_id);
CREATE INDEX IF NOT EXISTS idx_references_target ON "references"(target_headword);