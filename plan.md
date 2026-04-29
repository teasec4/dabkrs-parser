# План: вынести примеры ([ex]) в отдельную таблицу

## Текущая ситуация

Парсер находит `[ex]...[/ex]` в DSL, извлекает через `extractTags()` с типом "ex".
При импорте (мой последний фикс) meanings с тегом "ex" целиком скипаются — примеры теряются.

В схеме БД уже есть таблица `examples`, но она привязана к `meaning_id` (FK).
Если meaning скипается, пример в неё не попадает.

## Цель

Сохранять примеры в отдельную таблицу, НЕ создавая для них фейковый meaning-перевод.

## Варианты

### A. Привязать examples к entry напрямую

- Добавить `entry_id` в таблицу `examples` (рядом с существующим `meaning_id`, сделав его nullable)
- Или заменить `meaning_id` на `entry_id`
- При обработке meaning с `[ex]`:
  - Не создавать meaning row
  - Создавать example row: `(entry_id, text, order_num)`

### B. Отдельная таблица entry_examples

```sql
CREATE TABLE entry_examples (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id   INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    text       TEXT NOT NULL,
    order_num  INTEGER DEFAULT 0
);
```

Независима от meanings — примеры живут своей жизнью, не плодят мусорных переводов.

## Изменения в коде

1. **cmd/dict/main.go** — `convertSingleEntry()`: если meaning содержит `[ex]`, не скипать целиком,
   а собирать examples в Entry.Examples []string
2. **internal/parser/type.go** — добавить `Examples []string` в Entry
3. **internal/storage/sql.go** — `InsertEntry()`/`InsertEntriesBatch()`: вставлять examples
   в entry_examples (или в examples с entry_id)
4. **Schema** — обновить `initDB()` под новую структуру
5. **PARSER_ANALYSIS.md** — обновить схему БД в разделе 7

## Приоритет

Сначала пересобрать БД с текущим фиксом (скип [ex]).
Потом реализовать план и пересобрать снова.
