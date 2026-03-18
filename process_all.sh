#!/bin/bash

# Скрипт для обработки всех DSL файлов и создания единой базы данных

set -e  # Выход при ошибке

echo "=== Обработка всех DSL файлов словаря ==="
echo ""

# Переменные
JSON_FILE="./dictionary.json"
DB_FILE="./dictionary.db"
DSL_FILES=(
    "./dabkrs/dabkrs_1.dsl"
    "./dabkrs/dabkrs_2.dsl"
    "./dabkrs/dabkrs_3.dsl"
)

# Функция для вывода разделителя
separator() {
    echo ""
    echo "================================================"
    echo ""
}

# 1. Удаляем старые файлы если существуют
echo "1. Очистка старых файлов..."
if [ -f "$JSON_FILE" ]; then
    echo "   Удаляем $JSON_FILE"
    rm "$JSON_FILE"
fi

if [ -f "$DB_FILE" ]; then
    echo "   Удаляем $DB_FILE"
    rm "$DB_FILE"
fi

separator

# 2. Парсим все DSL файлы
echo "2. Парсинг DSL файлов..."
echo "   Файлы для обработки:"
for file in "${DSL_FILES[@]}"; do
    echo "   - $file"
done

echo ""
echo "   Запуск парсера..."
go run cmd/parser/main.go

if [ ! -f "$JSON_FILE" ]; then
    echo "   ОШИБКА: JSON файл не создан!"
    exit 1
fi

JSON_SIZE=$(du -h "$JSON_FILE" | cut -f1)
echo "   JSON файл создан: $JSON_FILE ($JSON_SIZE)"

separator

# 3. Создаем базу данных
echo "3. Создание базы данных..."
echo "   Запуск создания БД..."
go run cmd/createdb/main.go

if [ ! -f "$DB_FILE" ]; then
    echo "   ОШИБКА: База данных не создана!"
    exit 1
fi

DB_SIZE=$(du -h "$DB_FILE" | cut -f1)
echo "   База данных создана: $DB_FILE ($DB_SIZE)"

separator

# 4. Проверяем результат
echo "4. Проверка результата..."
echo "   Статистика базы данных:"

if command -v sqlite3 &> /dev/null; then
    # Получаем статистику из базы данных
    echo ""
    sqlite3 "$DB_FILE" << 'EOF'
SELECT
    (SELECT COUNT(*) FROM words) as word_count,
    (SELECT COUNT(*) FROM meanings) as meaning_count,
    ROUND((SELECT COUNT(*) FROM meanings)*1.0/(SELECT COUNT(*) FROM words), 2) as avg_meanings;
EOF

    echo ""
    echo "   Примеры записей:"
    sqlite3 "$DB_FILE" << 'EOF'
SELECT
    w.hanzi as "Иероглифы",
    w.pinyin as "Пиньинь",
    m.meaning as "Значение"
FROM words w
JOIN meanings m ON w.id = m.word_id
ORDER BY w.id
LIMIT 3;
EOF
else
    echo "   sqlite3 не установлен, пропускаем проверку"
fi

separator

# 5. Финальный отчет
echo "5. Финальный отчет"
echo ""
echo "Успешно обработаны все DSL файлы:"
for file in "${DSL_FILES[@]}"; do
    if [ -f "$file" ]; then
        file_size=$(du -h "$file" | cut -f1)
        echo "   ✓ $(basename "$file") ($file_size)"
    else
        echo "   ✗ $(basename "$file") (не найден)"
    fi
done

echo ""
echo "Созданные файлы:"
echo "   ✓ $JSON_FILE ($JSON_SIZE)"
echo "   ✓ $DB_FILE ($DB_SIZE)"

echo ""
echo "=== Обработка завершена успешно! ==="
echo ""
echo "База данных готова к использованию."
echo "Структура:"
echo "   words    - таблица слов (hanzi, pinyin)"
echo "   meanings - таблица значений (word_id, meaning)"
echo ""
echo "Примеры SQL запросов:"
echo "   # Поиск по иероглифам"
echo "   SELECT * FROM words WHERE hanzi LIKE '%中国%';"
echo ""
echo "   # Поиск по пиньиню"
echo "   SELECT * FROM words WHERE pinyin LIKE '%zhong1guo2%';"
echo ""
echo "   # Получить слова со значениями"
echo "   SELECT w.hanzi, w.pinyin, m.meaning"
echo "   FROM words w JOIN meanings m ON w.id = m.word_id"
echo "   WHERE w.hanzi LIKE '%中国%';"
