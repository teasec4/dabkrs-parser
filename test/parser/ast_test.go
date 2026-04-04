package parser_test

import (
	"parser/internal/parser"
	"testing"
)

func TestParse_BasicTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []parser.NodeType
	}{
		{
			name:     "paragraph",
			input:    "[p]text[/p]",
			expected: []parser.NodeType{parser.NodeParagraph},
		},
		{
			name:     "italic",
			input:    "[i]text[/i]",
			expected: []parser.NodeType{parser.NodeItalic},
		},
		{
			name:     "ref",
			input:    "[ref]word[/ref]",
			expected: []parser.NodeType{parser.NodeRef},
		},
		{
			name:     "container",
			input:    "[c]content[/c]",
			expected: []parser.NodeType{parser.NodeContainer},
		},
		{
			name:     "example",
			input:    "[ex]example[/ex]",
			expected: []parser.NodeType{parser.NodeExample},
		},
		{
			name:     "star",
			input:    "[*]marker",
			expected: []parser.NodeType{parser.NodeStar},
		},
		{
			name:     "meaning",
			input:    "[m]some meanign[/m]",
			expected: []parser.NodeType{parser.NodeUnknown},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := parser.Lex(tt.input)
			ast := parser.Parse(tokens)

			if len(ast.Children) != len(tt.expected) {
				t.Errorf("expected %d children, got %d", len(tt.expected), len(ast.Children))
			}

			for i, exp := range tt.expected {
				if i >= len(ast.Children) {
					break
				}
				if ast.Children[i].Type != exp {
					t.Errorf("child %d: expected %v, got %v", i, exp, ast.Children[i].Type)
				}
			}
		})
	}
}

func TestParse_NestedTags(t *testing.T) {
	input := "[p]text [i]italic[/i] more[/p]"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	if len(ast.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(ast.Children))
	}

	para := ast.Children[0]
	if para.Type != parser.NodeParagraph {
		t.Errorf("expected PARAGRAPH, got %v", para.Type)
	}

	if len(para.Children) != 3 {
		t.Errorf("expected 3 children in paragraph, got %d", len(para.Children))
	}
}

func TestParse_UnknownTags(t *testing.T) {
	input := "[m1]content[/m]"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	if len(ast.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(ast.Children))
	}

	if ast.Children[0].Type != parser.NodeMeaning {
		t.Errorf("expected MEANING for [m1], got %v", ast.Children[0].Type)
	}
}

func TestParse_TextOnly(t *testing.T) {
	input := "just text without tags"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	if len(ast.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(ast.Children))
	}

	if ast.Children[0].Type != parser.NodeText {
		t.Errorf("expected TEXT, got %v", ast.Children[0].Type)
	}

	if ast.Children[0].Value != input {
		t.Errorf("expected %q, got %q", input, ast.Children[0].Value)
	}
}

func TestParse_EmptyInput(t *testing.T) {
	tokens := parser.Lex("")
	ast := parser.Parse(tokens)

	if len(ast.Children) != 0 {
		t.Errorf("expected 0 children, got %d", len(ast.Children))
	}
}

func TestParse_MultipleSiblings(t *testing.T) {
	input := "[p]1[/p][i]2[/i][ref]3[/ref]"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	if len(ast.Children) != 3 {
		t.Errorf("expected 3 children, got %d", len(ast.Children))
	}

	expected := []parser.NodeType{parser.NodeParagraph, parser.NodeItalic, parser.NodeRef}
	for i, exp := range expected {
		if ast.Children[i].Type != exp {
			t.Errorf("child %d: expected %v, got %v", i, exp, ast.Children[i].Type)
		}
	}
}

func TestParse_DeepNesting(t *testing.T) {
	input := "[p][i][ref]text[/ref][/i][/p]"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	if len(ast.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(ast.Children))
	}

	// Navigate: CONTAINER -> PARAGRAPH -> ITALIC -> REF
	para := ast.Children[0]
	if para.Type != parser.NodeParagraph {
		t.Errorf("expected PARAGRAPH, got %v", para.Type)
	}

	if len(para.Children) != 1 {
		t.Errorf("PARAGRAPH should have 1 child, got %d", len(para.Children))
	}

	italic := para.Children[0]
	if italic.Type != parser.NodeItalic {
		t.Errorf("expected ITALIC, got %v", italic.Type)
	}

	if len(italic.Children) != 1 {
		t.Errorf("ITALIC should have 1 child, got %d", len(italic.Children))
	}

	ref := italic.Children[0]
	if ref.Type != parser.NodeRef {
		t.Errorf("expected REF, got %v", ref.Type)
	}
}

func TestParse_RealDSLntry(t *testing.T) {
	input := "[m1][p]г.[/p] Самаринда ([i]Индонезия[/i])[/m]"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	if len(ast.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(ast.Children))
	}

	meaning := ast.Children[0]
	if meaning.Type != parser.NodeMeaning {
		t.Errorf("expected MEANING, got %v", meaning.Type)
	}

	// Should have PARAGRAPH, TEXT, TEXT (with italic inside)
	if len(meaning.Children) < 3 {
		t.Errorf("expected at least 3 children in MEANING, got %d", len(meaning.Children))
	}
}

func TestParse_TextWithNewlines(t *testing.T) {
	input := "line1\nline2\r\nline3"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	if len(ast.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(ast.Children))
	}

	if ast.Children[0].Value != input {
		t.Errorf("text should preserve newlines")
	}
}

func TestParse_ChineeseAndPinyin(t *testing.T) {
	input := "三比西河\n sānbǐxīhé\n [m1]река[/m]"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	// First child is the text with hanzi and pinyin
	if len(ast.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(ast.Children))
	}

	text := ast.Children[0]
	if text.Type != parser.NodeText {
		t.Errorf("first child should be TEXT, got %v", text.Type)
	}

	// Should contain both hanzi and pinyin
	if text.Value == "" {
		t.Error("text should not be empty")
	}
}

func TestParse_TagWithoutClose(t *testing.T) {
	input := "[p]unclosed"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	// Should have PARAGRAPH with TEXT inside
	if len(ast.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(ast.Children))
	}

	para := ast.Children[0]
	if para.Type != parser.NodeParagraph {
		t.Errorf("expected PARAGRAPH, got %v", para.Type)
	}

	if len(para.Children) != 1 {
		t.Errorf("PARAGRAPH should have 1 child, got %d", len(para.Children))
	}

	if para.Children[0].Value != "unclosed" {
		t.Errorf("expected 'unclosed', got %q", para.Children[0].Value)
	}
}

func TestParse_OnlyClosingTags(t *testing.T) {
	input := "[/p]"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	// Should create CONTAINER with no children (or just root)
	if len(ast.Children) != 0 {
		t.Logf("Got %d children for only closing tag", len(ast.Children))
	}
}

func TestParse_StarWithContent(t *testing.T) {
	input := "[*]пример"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	if len(ast.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(ast.Children))
	}

	star := ast.Children[0]
	if star.Type != parser.NodeStar {
		t.Errorf("expected STAR, got %v", star.Type)
	}

	if len(star.Children) != 1 {
		t.Errorf("STAR should have 1 child, got %d", len(star.Children))
	}

	if star.Children[0].Type != parser.NodeText {
		t.Errorf("STAR child should be TEXT, got %v", star.Children[0].Type)
	}

	if star.Children[0].Value != "пример" {
		t.Errorf("expected 'пример', got %q", star.Children[0].Value)
	}
}

func TestParse_ExampleWithText(t *testing.T) {
	input := "[ex]usage example[/ex]"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	if len(ast.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(ast.Children))
	}

	ex := ast.Children[0]
	if ex.Type != parser.NodeExample {
		t.Errorf("expected EXAMPLE, got %v", ex.Type)
	}

	if len(ex.Children) != 1 {
		t.Errorf("EXAMPLE should have 1 child, got %d", len(ex.Children))
	}

	if ex.Children[0].Value != "usage example" {
		t.Errorf("expected 'usage example', got %q", ex.Children[0].Value)
	}
}

func TestParse_MixedContent(t *testing.T) {
	input := "[p]г.[/p]首都 ([i]城市[/i])"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	// Should have PARAGRAPH, TEXT, ITALIC as siblings
	if len(ast.Children) < 2 {
		t.Errorf("expected at least 2 children, got %d", len(ast.Children))
	}

	// First should be PARAGRAPH
	if ast.Children[0].Type != parser.NodeParagraph {
		t.Errorf("first child should be PARAGRAPH, got %v", ast.Children[0].Type)
	}
}

func TestParse_ContainerWithMultipleChildren(t *testing.T) {
	input := "[c]child1 [i]child2[/i] child3[/c]"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	if len(ast.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(ast.Children))
	}

	container := ast.Children[0]
	if container.Type != parser.NodeContainer {
		t.Errorf("expected CONTAINER, got %v", container.Type)
	}

	if len(container.Children) != 3 {
		t.Errorf("CONTAINER should have 3 children, got %d", len(container.Children))
	}
}

func TestParse_RefTagContent(t *testing.T) {
	input := "[ref]上海[/ref]"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	if len(ast.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(ast.Children))
	}

	ref := ast.Children[0]
	if ref.Type != parser.NodeRef {
		t.Errorf("expected REF, got %v", ref.Type)
	}

	// REF node should have TEXT child with the reference value
	if len(ref.Children) != 1 {
		t.Errorf("REF should have 1 child, got %d", len(ref.Children))
	}

	if ref.Children[0].Value != "上海" {
		t.Errorf("expected '上海', got %q", ref.Children[0].Value)
	}
}

func TestNodeType_String(t *testing.T) {
	tests := []struct {
		nodeType parser.NodeType
		expected string
	}{
		{parser.NodeText, "TEXT"},
		{parser.NodeParagraph, "PARAGRAPH"},
		{parser.NodeItalic, "ITALIC"},
		{parser.NodeRef, "REF"},
		{parser.NodeContainer, "CONTAINER"},
		{parser.NodeExample, "EXAMPLE"},
		{parser.NodeStar, "STAR"},
		{parser.NodeUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.nodeType.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.nodeType.String())
			}
		})
	}
}

func TestParse_AllTagTypes(t *testing.T) {
	input := "[p]1[/p][i]2[/i][c]3[/c][ref]4[/ref][ex]5[/ex][*]6[*]"
	tokens := parser.Lex(input)
	ast := parser.Parse(tokens)

	// We expect: PARAGRAPH, ITALIC, CONTAINER, REF, EXAMPLE, STAR, TEXT (for closing [*])
	expected := []parser.NodeType{
		parser.NodeParagraph,
		parser.NodeItalic,
		parser.NodeContainer,
		parser.NodeRef,
		parser.NodeExample,
		parser.NodeStar,
	}

	if len(ast.Children) < len(expected) {
		t.Errorf("expected at least %d children, got %d", len(expected), len(ast.Children))
	}

	for i, exp := range expected {
		if i >= len(ast.Children) {
			break
		}
		if ast.Children[i].Type != exp {
			t.Errorf("child %d: expected %v, got %v", i, exp, ast.Children[i].Type)
		}
	}
}
