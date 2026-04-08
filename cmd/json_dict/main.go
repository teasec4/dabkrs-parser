package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"parser/internal/parser"
	"strings"
	"unicode"
)

// DictionaryEntry — структура записи словаря для Chrome extension.
// Поле Word содержит иероглиф(ы), Pinyin — латинизированное произношение,
// Def — первое значение (строка).
type DictionaryEntry struct {
	Word   string `json:"word"`
	Pinyin string `json:"pinyin"`
	Def    string `json:"def"`
}

// containsChinese проверяет, содержит ли строка хотя бы один китайский иероглиф.
// Используется для предварительной проверки перед более строгой валидацией.
func containsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// isCommonChinese проверяет, состоит ли строка целиком из "обычных" китайских иероглифов.
// Отсеивает радикалы, редкие и диалектные символы, символы из расширенных блоков Unicode.
// Для Chrome extension мы хотим только основные, широко используемые иероглифы.
func isCommonChinese(s string) bool {
	for _, r := range s {
		if !isCommonChineseChar(r) {
			return false
		}
	}
	return true
}

// isCommonChineseChar проверяет, является ли символ "обычным" китайским иероглифом.
// Отсеиваются:
// - Радикалы (CJK Radicals Supplement, 2E80-2FFF)
// - Символы диалектных цифр (CJK Symbols and Punctuation, 3000-303F)
// - Символы совместимости (CJK Compatibility, F900-FAFF)
// - Расширение A (3400-4DBF) — редкие иероглифы
// - Расширения B-F (20000-2EBEF) — очень редкие иероглифы
func isCommonChineseChar(r rune) bool {
	if !unicode.Is(unicode.Han, r) {
		return false
	}
	if isRadical(r) {
		return false
	}
	// Диалектные цифры (〢, 〥, 〨 и т.д.)
	if r >= 0x3000 && r <= 0x303F {
		return false
	}
	// CJK Compatibility (декоративные варианты)
	if r >= 0xF900 && r <= 0xFAFF {
		return false
	}
	// Расширение A — редкие иероглифы
	if r >= 0x3400 && r <= 0x4DBF {
		return false
	}
	// Расширения B-F (очень редкие иероглифы)
	if r >= 0x20000 && r <= 0x2A6DF {
		return false
	}
	if r >= 0x2A700 && r <= 0x2B73F {
		return false
	}
	if r >= 0x2B740 && r <= 0x2B81F {
		return false
	}
	if r >= 0x2B820 && r <= 0x2CEAF {
		return false
	}
	if r >= 0x2CEB0 && r <= 0x2EBEF {
		return false
	}
	return true
}

// isRadical определяет, является ли символ китайским радикалом.
// Китайские радикалы — это базовые компоненты для построения иероглифов.
// Они расположены в блоках CJK Radicals Supplement (2E80-2EFF)
// и Kangxi Radicals (2F00-2FDF).
func isRadical(r rune) bool {
	return (r >= 0x2E80 && r <= 0x2EFF) ||
		(r >= 0x2F00 && r <= 0x2FDF) ||
		(r >= 0x2FF0 && r <= 0x2FFF)
}

func main() {
	var dslfiles string
	var output string

	flag.StringVar(&dslfiles, "import", "", "comma-separated list of DSL files to import")
	flag.StringVar(&output, "output", "dict.json", "output JSON file path")
	flag.Parse()

	if dslfiles == "" {
		flag.Usage()
		log.Fatal("Please provide DSL files with -import flag")
	}

	// result — карта словаря: ключ = иероглиф, значение = DictionaryEntry
	result := make(map[string]DictionaryEntry)
	added := 0
	skipped := 0

	fileList := strings.Split(dslfiles, ",")
	for _, file := range fileList {
		file = strings.TrimSpace(file)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", file)
			continue
		}

		fmt.Printf("Parsing %s...\n", file)

		r, err := parser.OpenDSL(file)
		if err != nil {
			log.Printf("OpenDSL %s: %v", file, err)
			continue
		}
		defer r.Close()

		parser.ParseFSMStream(r, func(entry parser.RawEntry) {
			hanzi := strings.TrimSpace(entry.Headword)
			runeCount := len([]rune(hanzi))

			// Фильтр 1: только 1-2 иероглифа
			if runeCount < 1 || runeCount > 2 {
				skipped++
				return
			}

			// Фильтр 2: без цифр и латинских букв
			if strings.ContainsAny(hanzi, "0123456789") || strings.ContainsAny(hanzi, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz") {
				skipped++
				return
			}

			// Фильтр 3: только обычные китайские иероглифы (без радикалов и редких символов)
			if !isCommonChinese(hanzi) {
				skipped++
				return
			}

			// Фильтр 4: дубликаты не добавляем
			if _, exists := result[hanzi]; exists {
				return
			}

			pinyin := parser.NormalizePinyin(entry.Pinyin)
			def := ""
			// Берём только первое значение (m1)
			if len(entry.Meanings) > 0 && entry.Meanings[0].Text != "" {
				def = entry.Meanings[0].Text
			}

			// Фильтр 5: отсеиваем перенаправления ("см." и "вм.")
			// Такие записи не помогают при поиске перевода
			if strings.Contains(def, "см.") || strings.Contains(def, "вм.") {
				skipped++
				return
			}

			result[hanzi] = DictionaryEntry{
				Word:   hanzi,
				Pinyin: pinyin,
				Def:    def,
			}
			added++

			if added%10000 == 0 {
				fmt.Printf("Processed %d entries (added %d, skipped %d)...\n", added+skipped, added, skipped)
			}
		})
	}

	fmt.Printf("\nTotal: added %d entries, skipped %d (didn't pass filters)\n", added, skipped)

	out, err := os.Create(output)
	if err != nil {
		log.Fatalf("Create output file: %v", err)
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(result); err != nil {
		log.Fatalf("Encode JSON: %v", err)
	}

	fmt.Printf("Saved to %s\n", output)
}
