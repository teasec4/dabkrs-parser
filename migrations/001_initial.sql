-- DSL Dictionary Schema
-- Chinese-Russian Bilingual Dictionary (БКРС)

-- entries (китайские слова/символы)
CREATE TABLE IF NOT EXISTS entries (
    id SERIAL PRIMARY KEY,
    hanzi TEXT NOT NULL UNIQUE,
    pinyin TEXT,
    pinyin_normalized TEXT
);
CREATE INDEX IF NOT EXISTS idx_entries_hanzi ON entries(hanzi);
CREATE INDEX IF NOT EXISTS idx_entries_pinyin_norm ON entries(pinyin_normalized);

-- meanings (значения слова)
CREATE TABLE IF NOT EXISTS meanings (
    id SERIAL PRIMARY KEY,
    entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    part_of_speech TEXT,
    order_num INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_meanings_entry ON meanings(entry_id);

-- refs (перекрёстные ссылки -> entries.hanzi)
CREATE TABLE IF NOT EXISTS refs (
    id SERIAL PRIMARY KEY,
    meaning_id INTEGER NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    target_entry_id INTEGER REFERENCES entries(id) ON DELETE SET NULL,
    target_hanzi TEXT
);
CREATE INDEX IF NOT EXISTS idx_refs_meaning ON refs(meaning_id);
CREATE INDEX IF NOT EXISTS idx_refs_target ON refs(target_entry_id);

-- examples (примеры предложений)
CREATE TABLE IF NOT EXISTS examples (
    id SERIAL PRIMARY KEY,
    meaning_id INTEGER NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    chinese TEXT NOT NULL,
    translation TEXT
);
CREATE INDEX IF NOT EXISTS idx_examples_meaning ON examples(meaning_id);
