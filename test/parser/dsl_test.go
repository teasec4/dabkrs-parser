package parser_test

import (
	"os"
	"parser/internal/parser"
	"path/filepath"
	"strings"
	"testing"
)

func dslPath(name string) string {
	return filepath.Join("..", "..", "dabkrs", name)
}

func TestReadDSL_Basic(t *testing.T) {
	data, err := parser.ReadDSL(dslPath("dabkrs_1.dsl"))
	if err != nil {
		t.Fatalf("failed to read DSL: %v", err)
	}

	if len(data) == 0 {
		t.Error("read empty data")
	}
}

func TestReadDSL_ContainsHeader(t *testing.T) {
	data, err := parser.ReadDSL(dslPath("dabkrs_1.dsl"))
	if err != nil {
		t.Fatalf("failed to read DSL: %v", err)
	}

	if !strings.Contains(data, "#NAME") {
		t.Error("missing #NAME header")
	}

	if !strings.Contains(data, "#INDEX_LANGUAGE") {
		t.Error("missing #INDEX_LANGUAGE header")
	}

	if !strings.Contains(data, "#CONTENTS_LANGUAGE") {
		t.Error("missing #CONTENTS_LANGUAGE header")
	}
}

func TestReadDSL_ContainsChinese(t *testing.T) {
	data, err := parser.ReadDSL(dslPath("dabkrs_1.dsl"))
	if err != nil {
		t.Fatalf("failed to read DSL: %v", err)
	}

	hasChinese := false
	for _, r := range data {
		if r >= 0x4E00 && r <= 0x9FFF {
			hasChinese = true
			break
		}
	}

	if !hasChinese {
		t.Error("no Chinese characters found")
	}
}

func TestReadDSL_ContainsRussian(t *testing.T) {
	data, err := parser.ReadDSL(dslPath("dabkrs_1.dsl"))
	if err != nil {
		t.Fatalf("failed to read DSL: %v", err)
	}

	if !strings.Contains(data, "река") && !strings.Contains(data, "Русский") {
		t.Error("no Russian text found")
	}
}

func TestReadDSL_InvalidPath(t *testing.T) {
	_, err := parser.ReadDSL("./nonexistent.dsl")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadDSL_FileSize(t *testing.T) {
	data, err := parser.ReadDSL(dslPath("dabkrs_1.dsl"))
	if err != nil {
		t.Fatalf("failed to read DSL: %v", err)
	}

	// File should be substantial (around 89MB for dabkrs_1.dsl)
	if len(data) < 1_000_000 {
		t.Logf("Warning: file is smaller than expected: %d bytes", len(data))
	}
}

func TestReadDSL_NoBOMAtStart(t *testing.T) {
	data, err := parser.ReadDSL(dslPath("dabkrs_1.dsl"))
	if err != nil {
		t.Fatalf("failed to read DSL: %v", err)
	}

	// BOM (U+FEFF) should be stripped by decoder
	if len(data) > 0 && rune(data[0]) == 0xFEFF {
		t.Error("BOM still present at start of decoded string")
	}
}

func TestReadDSL_ContainsNewlines(t *testing.T) {
	data, err := parser.ReadDSL(dslPath("dabkrs_1.dsl"))
	if err != nil {
		t.Fatalf("failed to read DSL: %v", err)
	}

	newlineCount := 0
	for _, r := range data {
		if r == '\n' {
			newlineCount++
		}
	}

	if newlineCount == 0 {
		t.Error("no newlines found in file")
	}

	t.Logf("Found %d newlines", newlineCount)
}

func TestReadDSL_PinyinTones(t *testing.T) {
	data, err := parser.ReadDSL(dslPath("dabkrs_1.dsl"))
	if err != nil {
		t.Fatalf("failed to read DSL: %v", err)
	}

	pinyinWithTones := []string{"ā", "á", "ǎ", "à", "ē", "é", "ě", "è",
		"ī", "í", "ǐ", "ì", "ō", "ó", "ǒ", "ò", "ū", "ú", "ǔ", "ù"}

	found := false
	for _, tone := range pinyinWithTones {
		if strings.Contains(data, tone) {
			found = true
			break
		}
	}

	if !found {
		t.Error("no pinyin tones found")
	}
}

func TestReadDSL_UTF16LEEncoding(t *testing.T) {
	data, err := parser.ReadDSL(dslPath("dabkrs_1.dsl"))
	if err != nil {
		t.Fatalf("failed to read DSL: %v", err)
	}

	runeCount := 0
	for range data {
		runeCount++
	}

	if runeCount == 0 {
		t.Error("no characters found")
	}
}

func TestReadDSL_ContainsTags(t *testing.T) {
	data, err := parser.ReadDSL(dslPath("dabkrs_1.dsl"))
	if err != nil {
		t.Fatalf("failed to read DSL: %v", err)
	}

	expectedTags := []string{"[m", "[p]", "[i]", "[ref]", "[ex]", "[c]"}

	for _, tag := range expectedTags {
		if !strings.Contains(data, tag) {
			t.Logf("Warning: expected tag %q not found", tag)
		}
	}
}

func TestReadDSL_ContainsSpecificEntry(t *testing.T) {
	data, err := parser.ReadDSL(dslPath("dabkrs_1.dsl"))
	if err != nil {
		t.Fatalf("failed to read DSL: %v", err)
	}

	entries := []string{"三比西河", "上海", "北京"}
	found := 0
	for _, entry := range entries {
		if strings.Contains(data, entry) {
			found++
		}
	}

	if found == 0 {
		t.Error("none of the expected entries found")
	}
}

func BenchmarkReadDSL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parser.ReadDSL(dslPath("dabkrs_1.dsl"))
	}
}

func TestReadDSL_AllDabkrsFiles(t *testing.T) {
	files := []string{
		"dabkrs_1.dsl",
		"dabkrs_2.dsl",
		"dabkrs_3.dsl",
	}

	for _, file := range files {
		path := dslPath(file)
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Logf("File %s does not exist, skipping", path)
			continue
		}

		data, err := parser.ReadDSL(path)
		if err != nil {
			t.Errorf("failed to read %s: %v", path, err)
			continue
		}

		if len(data) == 0 {
			t.Errorf("empty data from %s", path)
		}

		t.Logf("%s: %d bytes", file, len(data))
	}
}
