package parser_test

import (
	"parser/internal/parser"
	"testing"
)

func TestFSMBasic(tt *testing.T) {
	data := `三比西河
sānbǐxīhé
[m1]река Замбези[/m]
三个泉
sāngèquán
[m1]Саньгэцюань[/m]`

	entries := parser.ParseFSM(data)

	if len(entries) != 2 {
		tt.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	if entries[0].Headword != "三比西河" {
		tt.Errorf("Expected headword 三比西河, got %s", entries[0].Headword)
	}
	if entries[0].Pinyin != "sānbǐxīhé" {
		tt.Errorf("Expected pinyin sānbǐxīhé, got %s", entries[0].Pinyin)
	}
	if len(entries[0].Meanings) != 1 {
		tt.Errorf("Expected 1 meaning, got %d", len(entries[0].Meanings))
	}
	if entries[0].Meanings[0].Text != "река Замбези" {
		tt.Errorf("Expected meaning text 'река Замбези', got %s", entries[0].Meanings[0].Text)
	}

	tt.Logf("Entry 1: headword=%s pinyin=%s meanings=%d", entries[0].Headword, entries[0].Pinyin, len(entries[0].Meanings))
	tt.Logf("Entry 2: headword=%s pinyin=%s meanings=%d", entries[1].Headword, entries[1].Pinyin, len(entries[1].Meanings))
}

func TestFSMWithRef(tt *testing.T) {
	data := `上海
shànghǎi
[m1][p]см.[/p] [ref]上海[/ref][/m]`

	entries := parser.ParseFSM(data)

	if len(entries) != 1 {
		tt.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	// Text includes ref content since we don't remove ref tags from text
	tt.Logf("Entry: headword=%s meaning=%q tags=%v", entries[0].Headword, entries[0].Meanings[0].Text, entries[0].Meanings[0].Tags)

	if len(entries[0].Meanings[0].Tags) != 1 {
		tt.Fatalf("Expected 1 tag, got %d", len(entries[0].Meanings[0].Tags))
	}
	if entries[0].Meanings[0].Tags[0].Type != "ref" {
		tt.Errorf("Expected tag type 'ref', got %s", entries[0].Meanings[0].Tags[0].Type)
	}
	if entries[0].Meanings[0].Tags[0].Value != "上海" {
		tt.Errorf("Expected tag value '上海', got %s", entries[0].Meanings[0].Tags[0].Value)
	}
}

func TestFSMMultipleMeanings(tt *testing.T) {
	data := `土国
tǔguó
[m1]1) Турция[/m][m1]2) Туркменистан[/m]`

	entries := parser.ParseFSM(data)

	if len(entries) != 1 {
		tt.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	tt.Logf("Entry: headword=%s meanings=%d", entries[0].Headword, len(entries[0].Meanings))
	for i, m := range entries[0].Meanings {
		tt.Logf("  Meaning %d: level=%d text=%q", i, m.Level, m.Text)
	}
}

func TestFSMFullFile(tt *testing.T) {
	data, _ := parser.ReadDSL("../../dabkrs/dabkrs_1.dsl")
	// Just first 2000 bytes to see structure
	data = data[:2000]

	entries := parser.ParseFSM(data)

	tt.Logf("Parsed %d entries", len(entries))
	for i, e := range entries {
		if i < 15 {
			tt.Logf("Entry %d: headword=%q pinyin=%q meanings=%d", i, e.Headword, e.Pinyin, len(e.Meanings))
			for j, m := range e.Meanings {
				tt.Logf("  Meaning %d: level=%d text=%q", j, m.Level, m.Text)
			}
		}
	}
}
