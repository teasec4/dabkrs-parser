package parser_test

import (
    "parser/internal/parser"
    "testing"
)

func TestCheckIncludes(t *testing.T) {
    // Check if #INCLUDE is being processed
    data, _ := parser.ReadDSL("../../dabkrs/dabkrs_1.dsl")
    
    // Count #INCLUDE directives
    count := 0
    for i := 0; i < len(data)-10; i++ {
        if data[i:i+8] == "#INCLUDE" {
            count++
            // Get the file name
            end := i + 8
            for end < len(data) && data[end] != '\n' {
                end++
            }
            t.Logf("Found INCLUDE: %s", data[i:end])
        }
    }
    t.Logf("Total INCLUDE directives: %d", count)
    
    // Now manually parse all 3 files
    entries := 0
    for _, fname := range []string{"../../dabkrs/dabkrs_1.dsl", "../../dabkrs/dabkrs_2.dsl", "../../dabkrs/dabkrs_3.dsl"} {
        data, _ := parser.ReadDSL(fname)
        tokens := parser.Lex(data)
        ast := parser.Parse(tokens)
        ents := parser.ExtractEntries(ast, 0)
        entries += len(ents)
        t.Logf("%s: %d entries", fname, len(ents))
    }
    t.Logf("Total: %d entries", entries)
}
