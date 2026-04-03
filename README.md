# БКРС Parser

Парсер китайско-русского словаря БКРС (Большой Китайско-Русский Словарь) для извлечения данных в SQLite базу.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge)
![SQLite](https://img.shields.io/badge/SQLite-3-brightgreen?style=for-the-badge)
![DSL Format](https://img.shields.io/badge/DSL-v1-orange?style=for-the-badge)

## Возможности

- Парсинг DSL файлов формата Lingvo
- Поддержка UTF-16LE кодировки
- Извлечение: entries, meanings, refs, examples
- Нормализация пиньинь (удаление тонов)
- Экспорт в SQLite базу
- CLI для импорта и поиска

## Установка

```bash
# Клонирование
git clone <repo-url>
cd parser

# Сборка
go build -o dict ./cmd/dict/
```

## Использование

### Импорт словаря в базу

```bash
# Импорт одного файла
./dict -db dictionary.db -import dabkrs/dabkrs_1.dsl

# Импорт всех файлов
./dict -db dictionary.db -import dabkrs/dabkrs_1.dsl,dabkrs/dabkrs_2.dsl,dabkrs/dabkrs_3.dsl

# С лимитом для тестирования
./dict -db test.db -import dabkrs/dabkrs_1.dsl -limit 1000
```

### Поиск

```bash
# Поиск по иероглифам (prefix)
./dict -db dictionary.db -search 上海

# Поиск по пиньинь
./dict -db dictionary.db -search shang -pinyin

# Статистика базы
./dict -db dictionary.db
```

## Структура базы данных

### Таблица `entries`

Основная таблица словарных статей.

| Поле | Тип | Описание |
|------|-----|----------|
| id | INTEGER | PRIMARY KEY |
| hanzi | TEXT | Китайские иероглифы (UNIQUE) |
| pinyin | TEXT | Пиньинь с тонами |
| pinyin_normalized | TEXT | Пиньинь без тонов |

```sql
CREATE TABLE entries (
    id INTEGER PRIMARY KEY,
    hanzi TEXT NOT NULL UNIQUE,
    pinyin TEXT,
    pinyin_normalized TEXT
);
CREATE INDEX idx_entries_hanzi ON entries(hanzi);
CREATE INDEX idx_entries_pinyin_norm ON entries(pinyin_normalized);
```

### Таблица `meanings`

Значения и переводы.

| Поле | Тип | Описание |
|------|-----|----------|
| id | INTEGER | PRIMARY KEY |
| entry_id | INTEGER | FK → entries.id |
| text | TEXT | Текст перевода |
| part_of_speech | TEXT | Часть речи |
| order_num | INTEGER | Порядковый номер |

```sql
CREATE TABLE meanings (
    id INTEGER PRIMARY KEY,
    entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    part_of_speech TEXT,
    order_num INTEGER DEFAULT 0
);
CREATE INDEX idx_meanings_entry ON meanings(entry_id);
```

### Таблица `refs`

Перекрёстные ссылки между статьями.

| Поле | Тип | Описание |
|------|-----|----------|
| id | INTEGER | PRIMARY KEY |
| meaning_id | INTEGER | FK → meanings.id |
| target_entry_id | INTEGER | FK → entries.id (nullable) |
| target_hanzi | TEXT | Текст ссылки |

```sql
CREATE TABLE refs (
    id INTEGER PRIMARY KEY,
    meaning_id INTEGER NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    target_entry_id INTEGER REFERENCES entries(id) ON DELETE SET NULL,
    target_hanzi TEXT
);
CREATE INDEX idx_refs_meaning ON refs(meaning_id);
CREATE INDEX idx_refs_target ON refs(target_entry_id);
```

### Таблица `examples`

Примеры использования.

| Поле | Тип | Описание |
|------|-----|----------|
| id | INTEGER | PRIMARY KEY |
| meaning_id | INTEGER | FK → meanings.id |
| chinese | TEXT | Пример на китайском |
| translation | TEXT | Перевод примера |

```sql
CREATE TABLE examples (
    id INTEGER PRIMARY KEY,
    meaning_id INTEGER NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    chinese TEXT NOT NULL,
    translation TEXT
);
CREATE INDEX idx_examples_meaning ON examples(meaning_id);
```

## Структура проекта

```
parser/
├── cmd/
│   └── dict/           # CLI приложение
├── dabkrs/             # DSL файлы словаря
├── internal/
│   ├── parser/          # Парсер DSL
│   │   ├── lexer.go     # Токенизатор
│   │   ├── ast.go       # AST построение
│   │   ├── extractor.go # Извлечение entries
│   │   ├── dsl.go       # Чтение DSL файлов
│   │   └── dsl_parser.go
│   └── storage/
│       └── sql.go        # SQLite операции
├── migrations/          # SQL миграции
└── test/parser/        # Тесты
```

## API для Go

```go
import "parser/internal/parser"
import "parser/internal/storage"

// Парсинг DSL файла
entries, err := parser.ParseFile("dict.dsl", 0) // 0 = без лимита

// Работа с базой
db, _ := storage.NewDB("dict.db")
db.InsertEntries(entries, 1000)  // batchSize = 1000
db.ResolveRefs()

// Поиск
results, _ := db.Search("北京", 20)
entry, _ := db.GetEntryByHanzi("学习")
```

## Тесты

```bash
go test ./... -v
```

## Примеры данных

```sql
-- Найти слово
SELECT * FROM entries WHERE hanzi = '学习';
-- Result: 学习|xuéxí|xuexi

-- Найти по пиньинь
SELECT * FROM entries WHERE pinyin_normalized LIKE 'xue%';

-- Получить все значения слова
SELECT m.* FROM meanings m 
JOIN entries e ON m.entry_id = e.id 
WHERE e.hanzi = '学习';

-- Перекрёстные ссылки
SELECT e1.hanzi, e2.hanzi 
FROM refs r
JOIN entries e1 ON r.meaning_id = e1.id
JOIN entries e2 ON r.target_entry_id = e2.id;
```

## Формат DSL

DSL файлы используют XML-подобные теги:

```
上海
 shanghai
 [m1][p]см.[/p] [ref]北京[/ref][/m]
```

Теги:
- `[m1]` - новое значение
- `[p]` - часть речи
- `[ref]` - перекрёстная ссылка
- `[ex]` - пример
- `[i]` - курсив
- `[c]` - комментарий

## TODO

- [ ] Заполнение пиньинь через AI для записей без пиньинь
- [ ] Сегментация китайских фраз на слова
- [ ] Полнотекстовый поиск
- [ ] API сервер

## Лицензия

MIT
