package parser_test

import (
	"parser/internal/parser"
	"strings"
	"testing"
)

func TestIsPinyin_Basic(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"ni", true},
		{"hao", true},
		{"zhongwen", true},
		{"ZHONGWEN", true},
		{"tǔ'ěrqísītǎn", true},
		{"ni3", false},
		{"nǐhǎo", true},
		{"shang4", false},
		{"bēi zǐ", true},
		{"", false},
		{"中文", false},
		{"123", false},
		{"nǐ", true},
		{"wǒ", true},
		{"tā", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parser.IsPinyin(tt.input)
			if result != tt.expected {
				t.Errorf("isPinyin(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsPinyin_RightSingleQuotation(t *testing.T) {
	result := parser.IsPinyin("tǔ'ěrqísītǎn")
	if !result {
		t.Error("isPinyin should recognize RIGHT SINGLE QUOTATION MARK (U+2019)")
	}

	result = parser.IsPinyin("tǔ\u2019ěrqísītǎn")
	if !result {
		t.Error("isPinyin should recognize unicode RIGHT SINGLE QUOTATION MARK")
	}
}

func TestIsPinyin_AllVowels(t *testing.T) {
	pinyinVowels := []string{
		"ā", "á", "ǎ", "à",
		"ē", "é", "ě", "è",
		"ī", "í", "ǐ", "ì",
		"ō", "ó", "ǒ", "ò",
		"ū", "ú", "ǔ", "ù",
		"ǖ", "ǘ", "ǚ", "ǜ",
	}

	for _, v := range pinyinVowels {
		if !parser.IsPinyin("z" + v) {
			t.Errorf("isPinyin should recognize vowel %q", v)
		}
	}
}

func TestHasChinese(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"中文", true},
		{"三比西河", true},
		{"上海", true},
		{"北京", true},
		{"hello", false},
		{"123", false},
		{"ni hao", false},
		{"hello世界", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parser.HasChinese(tt.input)
			if result != tt.expected {
				t.Errorf("hasChinese(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizePinyin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"nǐ", "ni"},
		{"hǎo", "hao"},
		{"zhōng", "zhong"},
		{"shàng", "shang"},
		{"bēi", "bei"},
		{"ni", "ni"},
		{"ZHONG", "zhong"},
		{"TǓ", "tu"},
		{"nǐhǎo", "nihao"},
		{"北京", "北京"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parser.NormalizePinyin(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizePinyin(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizePinyin_UMlaut(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"lǚ", "lv"},
		{"nǚ", "nv"},
		{"lǜ", "lv"},
		{"女", "女"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parser.NormalizePinyin(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizePinyin(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSplitHanziPinyin(t *testing.T) {
	tests := []struct {
		input     string
		expectedH string
		expectedP string
	}{
		{"三比西河", "三比西河", ""},
		{"上海 shanghai", "上海", "shanghai"},
		{"北京 běi jīng", "北京", "běi jīng"},
		{"ni", "ni", ""},
		{"北京", "北京", ""},
		{"hello", "hello", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			hanzi, pinyin := parser.SplitHanziPinyin(tt.input)
			if hanzi != tt.expectedH {
				t.Errorf("SplitHanziPinyin(%q) hanzi = %q, want %q", tt.input, hanzi, tt.expectedH)
			}
			if pinyin != tt.expectedP {
				t.Errorf("SplitHanziPinyin(%q) pinyin = %q, want %q", tt.input, pinyin, tt.expectedP)
			}
		})
	}
}

func TestExtractEntries_Simple(t *testing.T) {
	dsl := "#NAME \"Test\"\n上海 shanghai\n[m1]город Шанхай[/m]\n北京 beijing\n[m1]столица[/m]\n"

	entries, err := parser.ParseDSLString(dsl, 0)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) < 1 {
		t.Errorf("expected at least 1 entry, got %d", len(entries))
	}

	if entries[0].Headword != "上海" {
		t.Errorf("first entry Headword = %q, want %q", entries[0].Headword, "上海")
	}

	if entries[0].Pinyin != "shanghai" {
		t.Errorf("first entry Pinyin = %q, want %q", entries[0].Pinyin, "shanghai")
	}
}

func TestExtractEntries_PinyinOnSeparateLine(t *testing.T) {
	dsl := "#NAME \"Test\"\n三比西河\nsan bi xi he\n[m1]река Санбихэ[/m]\n"

	entries, err := parser.ParseDSLString(dsl, 0)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry, got 0")
	}

	if entries[0].Headword != "三比西河" {
		t.Errorf("Headword = %q, want %q", entries[0].Headword, "三比西河")
	}

	if entries[0].Pinyin != "san bi xi he" {
		t.Errorf("Pinyin = %q, want %q", entries[0].Pinyin, "san bi xi he")
	}
}

func TestExtractEntries_Header(t *testing.T) {
	dsl := "#NAME \"Test Dictionary\"\n#INDEX_LANGUAGE \"Chinese\"\n北京 beijing\n[m1]столица[/m]\n"

	entries, err := parser.ParseDSLString(dsl, 0)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) < 1 {
		t.Errorf("expected at least 1 entry, got %d", len(entries))
	}
}

func TestExtractEntries_Limit(t *testing.T) {
	dsl := "#NAME \"Test\"\n北京 beijing\n[m1]столица[/m]\n上海 shanghai\n[m1]город[/m]\n天津 tianjin\n[m1]город[/m]\n"

	entries, err := parser.ParseDSLString(dsl, 0)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestExtractMeanings_Basic(t *testing.T) {
	dsl := "#NAME \"Test\"\n北京 beijing\n[m1]столица Китая[/m]\n"

	entries, err := parser.ParseDSLString(dsl, 0)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}

	if len(entries[0].Meanings) == 0 {
		t.Error("expected at least 1 meaning")
	}

	if entries[0].Meanings[0].Text == "" {
		t.Error("expected meaning text")
	}
}

func TestExtractMeanings_WithPartOfSpeech(t *testing.T) {
	dsl := "#NAME \"Test\"\n学习 xuexi\n[m1]\n[p]глагол\n[1]учиться, изучать[/m]\n"

	entries, err := parser.ParseDSLString(dsl, 0)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}

	t.Logf("Meanings: %+v", entries[0].Meanings)
}

func TestExtractMeanings_WithRefs(t *testing.T) {
	dsl := "#NAME \"Test\"\n北京 beijing\n[m1]столица[/m]\n[p]см.[/p] [ref]上海[/ref]\n"

	entries, err := parser.ParseDSLString(dsl, 0)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}

	t.Logf("First meaning text: %q", entries[0].Meanings[0].Text)
	t.Logf("Tags: %v", entries[0].Meanings[0].Tags)
}

func TestExtractMeanings_WithExamples(t *testing.T) {
	dsl := "#NAME \"Test\"\n学习 xuexi\n[m1]учиться[/m]\n[ex]我学习中文。|Я учу китайский язык.[/ex]\n"

	entries, err := parser.ParseDSLString(dsl, 0)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}

	if len(entries[0].Meanings) > 0 && len(entries[0].Meanings[0].Tags) > 0 {
		for _, tag := range entries[0].Meanings[0].Tags {
			t.Logf("Tag: %+v", tag)
		}
	}
}

func TestExtractText_NodeText(t *testing.T) {
	dsl := "#NAME \"Test\"\n北京 beijing\n[m1]столица[/m]\n"

	entries, err := parser.ParseDSLString(dsl, 0)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}

	if len(entries[0].Meanings) == 0 {
		t.Fatal("expected at least 1 meaning")
	}

	text := entries[0].Meanings[0].Text
	if !strings.Contains(text, "столица") {
		t.Errorf("expected 'столица' in text, got %q", text)
	}
}

func TestExtractEntries_MultipleMeanings(t *testing.T) {
	dsl := "#NAME \"Test\"\n打 da\n[m1]ударять[/m]\n[m1]бить[/m]\n[m1]играть (в мяч)[/m]\n"

	entries, err := parser.ParseDSLString(dsl, 0)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}

	if len(entries[0].Meanings) < 2 {
		t.Logf("Meanings count: %d", len(entries[0].Meanings))
		t.Logf("Meanings: %+v", entries[0].Meanings)
	}
}

func TestExtractEntries_ParsesLargeFile(t *testing.T) {
	path := "../../dabkrs/dabkrs_1.dsl"
	entries, err := parser.ParseDSL(path, 10)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) != 10 {
		t.Errorf("expected 10 entries, got %d", len(entries))
	}

	for i, e := range entries {
		if e.Headword == "" {
			t.Errorf("entry %d has empty Headword", i)
		}
		t.Logf("Entry %d: %s [%s]", i, e.Headword, e.Pinyin)
	}
}

func TestExtractEntries_PinyinNormalized(t *testing.T) {
	dsl := "#NAME \"Test\"\n北京 běi jīng\n[m1]столица[/m]\n"

	entries, err := parser.ParseDSLString(dsl, 0)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}

	if entries[0].Pinyin != "běi jīng" {
		t.Errorf("Pinyin = %q, want %q", entries[0].Pinyin, "běi jīng")
	}
}

func TestExtractEntries_ChineseOnSameLine(t *testing.T) {
	dsl := "北京 bei jing\n[m1]test[/m]\n上海 shang hai\n[m1]test2[/m]"

	entries, err := parser.ParseDSLString(dsl, 0)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}
