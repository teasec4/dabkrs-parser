# Анализ парсера BKRS DSL → SQLite

## 1. Цель проекта

Извлечь китайско-русский словарь из сырых DSL-файлов BKRS (DslBuilder формат ABBYY Lingvo)
в структурированную SQLite БД, отфильтровав мусор.

**Бизнес-цель:** чистая БД для translatechinese.online — быстрый поиск слов китайский→русский,
без фраз, без игровых entry, без мусорных headword.

## 2. Архитектура парсера

```
main.go (CLI entry point)
  │
  ├── OpenDSL(file) → io.Reader (UTF-16LE → UTF-8)
  │
  ├── ParseFSMStream(reader, callback)  [fsm_parser.go]
  │     │
  │     └── FSM с 3 состояниями:
  │           StateExpectHeadword → StateExpectPinyin → StateExpectMeaning
  │
  └── callback → convertSingleEntry(raw) → sql.InsertEntries(batch)
```

**Потоковый парсинг:** файл читается строку за строкой, память не аккумулируется.
Батчи по 1000 entry → INSERT транзакциями.

### 2.1 FSM-логика (fsm_parser.go)

| Состояние | Детекция | Действие |
|---|---|---|
| `StateExpectHeadword` | Строка с колонки 0 (без отступа), содержит китайские иероглифы | Начинает новый RawEntry |
| `StateExpectPinyin` | Строка без китайских, не начинается с `[` | Сохраняет пиньинь |
| | Строка с `[m` | Переход к парсингу meaning блоков |
| `StateExpectMeaning` | Строка с `[m` | Парсинг meaning блока через parseMeaningBlock() |
| | Китайские в колонке 0 | Новый headword (предыдущий entry завершён) |
| | Строка с отступом (whitespace) | **Fallback:** малформированный meaning без `[m` тега |
| | `#` | Сброс состояния (комментарий/#include) |

### 2.2 Фильтры в convertSingleEntry (main.go)

1. **Empty headword check** — пустой headword = skip
2. **No-Han filter** — headword без китайских иероглифов = skip
3. **GSE phrase filter** — headword с >2 сегментами GSE = skip (фраза, не слово)
4. **Meaning cleaner** — `cleanMeaningText()`:
   - Удаляет все китайские иероглифы из meaning text
   - Удаляет слова с тоновыми гласными (āáǎà…) — это пиньинь
   - Если после очистки осталось <2 символов, meaning удаляется

## 3. Проблемы сырых DSL-файлов

### 3.1 Формат DSL

```
兰州                     ← headword (колонка 0, без отступа)
 lánzhōu                 ← пиньинь (1 пробел отступа)
 [m1]Ланьчжоу[/m]       ← meaning (с тегом)

[m2][p]геогр.[/p] ... [/m]  ← несколько meaning уровней
[p]жарг.[/p]                 ← теги частей речи
[ref]см. XXX[/ref]           ← ссылки на другие entry
[ex]例句[/ex]                 ← примеры
```

### 3.2 Типы мусора

| Тип | Пример | Как обрабатывается |
|---|---|---|
| **False headword** — meaning lines, содержащие китайские | Meaning line с `[ref]楼主[/ref]` внутри → containsChinese() true | Исправлено: проверка `startsWithWhitespace()` — строка с отступом не может быть headword |
| **Малформированные meaning lines** (нет `[m` тега) | ` Ланьчжоу (...)[m1]-----[/m]...` — meaning текст без открывающего тега | Fallback по whitespace — строка с отступом без `[m` = implicit meaning |
| **Фразы/предложения как headword** | `中国共产党中央委员会政治局` (14 иероглифов) | GSE сегментация: если >2 сегментов → skip |
| **Не-китайские headword** | `c 数`, `ins风`, `Вот` | No-Han filter |
| **Китайские символы в meaning text** | `Ланьчжоу ([i]городской округ в провинции Ганьсу, КНР[/i])` — без китайских, но `[ref]楼主[/ref]` содержит | cleanMeaningText удаляет все Han символы |
| **Пиньинь в meaning text** | То же самое, если пиньинь просочился | cleanMeaningText удаляет слова с тоновыми гласными |
| **`-----` маркеры в DSL** | `[m1]-----[/m]` — маркер конца multiline meaning | extractMeaningText не вырезает сам текст `-----` (визуальный шум) |
| **Ссылочные entry** | `边减` → meaning только `[ref]轮边减速器[/ref]` | cleanMeaningText удаляет китайские → meaning пустой → entry без meanings |

### 3.3 Статистика DSL-файлов

- Три файла: `dabkrs_1.dsl`, `dabkrs_2.dsl`, `dabkrs_3.dsl` (UTF-16 LE)
- `dabkrs_1.dsl` включает 2 и 3 через `#INCLUDE` (но парсер скипает `#` строки)
- ~611 малформированных meaning lines (без `[m` тега)

## 4. Сравнение баз данных

| Метрика | Старая БД (dictionary.db) | Новая БД (dabkrs_clean.db) |
|---|---|---|
| **Entries** | 3,433,812 | 1,700,401 (−50.5%) |
| **Zero-meanings** | 5,134 (0.15%) | 4,606 (0.27%) |
| **Без пиньиня** (_ или null) | 2,649,923 (77%) | 1,138,726 (67%) |
| **Meanings total** | 3,611,215 | 1,937,243 |
| **Avg meanings/entry** | 1.05 | 1.14 |
| **Размер файла** | 638 MB | 312 MB |

### 4.1 Почему новая БД в 2 раза меньше?

1. **False headword устранены** — старый парсер создавал фейковые entry из meaning lines, содержащих китайские иероглифы (типа `[ref]XXX[/ref]`). Это была основная причина 3.4M → 1.7M.

2. **No-Han filter** — entry без иероглифов в headword не попадают в БД.

3. **GSE phrase filter** — длинные фразы (>2 сегментов) отсекаются.

4. **cleanMeaningText** — китайские символы из meaning text удалены (примеры, ссылки).

### 4.2 Верификация ключевых entry

| Entry | Статус | Было | Стало |
|---|---|---|---|
| 兰州 | ✅ | 0 meanings (false headword) | Есть: "Ланьчжоу (городской округ...)" |
| 湖南 | ✅ | Не найден | Есть: 2 meanings (Хунань + Хонам) |
| 肛门 | ✅ | 0 meanings (malformed) | Есть: "заднепроходное отверстие, анус" |
| 怪兽 | ✅ | 0 meanings | Есть: "чудовище, чудище, монстр" |
| 纽伦堡 | ✅ | 0 meanings (no `[m` tag) | Есть: "Нюрнберг (город в Германии)" |

## 5. История изменений

### Этап 1: Исправление false headword detection
- **Проблема:** `containsChinese()` в `StateExpectMeaning` срабатывал даже на meaning lines с `[ref]楼主[/ref]` — строка содержит китайские, детектится как новый headword
- **Решение:** проверка `startsWithWhitespace(rawLine)` — если строка начинается с пробела/таба, это не headword

### Этап 2: Малформированные meaning lines
- **Проблема:** 611 строк-meaning без `[m` тега. Старый fallback (`!containsChinese(line)`) пропускал те, что содержат китайские внутри тегов
- **Решение:** fallback по whitespace — любая строка с отступом, не начинающаяся с `[m` или китайских в колонке 0 = implicit meaning

### Этап 3: Увеличение лимита meanings
- **Проблема:** лимит 5 meanings отбрасывал большую часть значений для многозначных слов (типа 是 с 156 meanings)
- **Решение:** лимит убран полностью (было 5, потом 20)

### Этап 4: Чистка meaning text
- **Проблема:** в meaning text попадали китайские иероглифы (примеры: `Ланьчжоу ([i]городской округ...[i])`) и пиньинь
- **Решение:** `cleanMeaningText()` — вырезает все Han символы и слова с тоновыми гласными

### Этап 5: No-Han filter
- **Проблема:** entry без китайских в headword (типа `c 数`, `Вот`) не имеют смысла в китайско-русском словаре
- **Решение:** skip entry, у которых headword не содержит ни одного китайского иероглифа

### Этап 6: GSE phrase filter
- **Проблема:** длинные фразы и предложения как headword (GSE сегментирует их в >2 частей)
- **Решение:** если GSE выдаёт >2 сегментов для headword → entry скипается

## 6. Оставшиеся проблемы

### 6.1 Zero-meanings (4,606 entry, 0.27%)
Это entry, у которых все meaning text после `cleanMeaningText()` стали пустыми.
Причина: meaning состоял только из китайских символов (`[ref]轮边减速器[/ref]`).
Такие entry корректно существуют в БД, но без перевода.

**Возможные решения:**
- Не сохранять entry, у которых 0 meanings после очистки
- Сохранять `[ref]` как redirect / cross-reference отдельно

### 6.2 "-----" в meaning text
DSL-формат использует `[m1]-----[/m]` как маркер окончания multiline блока.
Функция `extractMeaningText()` удаляет теги, но `-----` остаётся в тексте.

**Пример:** `Ланьчжоу (городской округ...)-----см.`

**Решение:** добавить `-----` в replacer в `extractMeaningText()`

### 6.3 Нет дедупликации дублей
Если один и тот же headword встречается в DSL несколько раз с разными meanings,
`ON CONFLICT(headword) DO UPDATE` UPSERT'ит существующую запись, затирая старые meanings.

### 6.4 GSE path resolution (Docker)
При сборке Docker нужно указывать `GSE_DATA_DIR` на уровень выше `data/dict/zh/`,
потому что `gse.LoadDict()` делает `path.Join(path.Dir(GSE_DATA_DIR), "data")`.
См. skill `gse-docker-dict-path`.

## 7. Схема БД

```sql
entries (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    headword          TEXT NOT NULL UNIQUE,
    pinyin            TEXT,
    pinyin_normalized TEXT
)

meanings (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id  INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    level     INTEGER DEFAULT 0,
    text      TEXT NOT NULL,
    order_num INTEGER DEFAULT 0
)

examples (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    meaning_id INTEGER NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    text       TEXT NOT NULL
)

references (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    meaning_id      INTEGER NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    target_headword TEXT NOT NULL
)
```

## 8. Идеи для улучшения

1. **Чистка `-----`** — добавить `"-----"` в replacer extractMeaningText()
2. **FTS5 индекс** — полнотекстовый поиск по русским meaning
3. **Cross-reference отдельно** — entry с only-ref meanings как redirects
4. **GSE threshold настраиваемый** — `--max-segments=N` флаг
5. **Эвристика: длина headword > 20 иероглифов = фраза** — дубль к GSE
6. **Merge meanings при UPSERT** — при дубликате headword не затирать старые meanings, а мержить
7. **Deferred constraints** — включить foreign keys в SQLite

## 9. Команды

```bash
# Сборка
go build -o dict ./cmd/dict/

# Импорт всех трёх файлов
./dict -db dabkrs_clean.db -import "dabkrs/dabkrs_1.dsl,dabkrs/dabkrs_2.dsl,dabkrs/dabkrs_3.dsl"

# Поиск
./dict -db dabkrs_clean.db -search "兰州"
./dict -db dabkrs_clean.db -pinyin -search "lanzhou"

# Статистика
./dict -db dabkrs_clean.db

# GSE path (macOS локально, не Docker)
GSE_DATA_DIR=/Users/yg_kovalev/go/pkg/mod/github.com/go-ego/gse@v1.0.2/ ./dict -db test.db -import "..."
```
