package parser_test

import (
	"encoding/json"
	"os"
	"parser/internal/parser"
	"path/filepath"
	"strings"
	"testing"
)

func TestGolden(t *testing.T) {
	dslFiles, err := filepath.Glob("../../testdata/*.dsl")
	if err != nil {
		t.Fatal(err)
	}

	if len(dslFiles) == 0 {
		t.Fatal("no .dsl files found in testdata")
	}

	updateMode := os.Getenv("UPDATE_GOLDEN") == "1"

	for _, dslPath := range dslFiles {
		name := filepath.Base(dslPath)
		t.Run(name, func(t *testing.T) {
			content, err := os.ReadFile(dslPath)
			if err != nil {
				t.Fatalf("failed to read %s: %v", dslPath, err)
			}

			entries := runPipeline(string(content))
			gotBytes, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				t.Fatalf("failed to marshal entries: %v", err)
			}
			got := string(gotBytes)

			goldenPath := strings.TrimSuffix(dslPath, ".dsl") + ".golden.json"
			expectedBytes, err := os.ReadFile(goldenPath)
			if err != nil {
				if updateMode {
					err := os.WriteFile(goldenPath, gotBytes, 0644)
					if err != nil {
						t.Fatalf("failed to create golden file: %v", err)
					}
					t.Logf("Created golden file: %s", goldenPath)
					return
				}
				t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
			}
			expected := string(expectedBytes)

			if updateMode {
				err := os.WriteFile(goldenPath, gotBytes, 0644)
				if err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
				t.Logf("Updated golden file: %s", goldenPath)
				return
			}

			if got != expected {
				t.Errorf("mismatch for %s\n\nGOT:\n%s\n\nEXPECTED:\n%s",
					name, got, expected)
			}
		})
	}
}

func runPipeline(input string) []parser.Entry {
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)
	return parser.ExtractEntries(ast, 0)
}
