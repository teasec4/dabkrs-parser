package parser_test

import (
    "parser/internal/parser"
    "testing"
)

func TestCheckEntries(t *testing.T) {
    data, _ := parser.ReadDSL("../../dabkrs/dabkrs_1.dsl")
    tokens := parser.Lex(data)
    ast := parser.Parse(tokens)
    entries := parser.ExtractEntries(ast, 0)
    
    t.Logf("Total entries: %d", len(entries))
    
    found := false
    for _, e := range entries {
        if e.Headword == "北京" || e.Headword == "上海" || e.Headword == "三比西河" {
            t.Logf("Found: %s [%s]", e.Headword, e.Pinyin)
            found = true
        }
    }
    
    if !found {
        t.Log("Looking for entries...")
        for i, e := range entries {
            if i < 20 {
                t.Logf("  %d: %s [%s]", i, e.Headword, e.Pinyin)
            }
        }
    }
}
