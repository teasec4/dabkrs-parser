package parser_test

import (
	"parser/internal/parser"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []parser.Token
	}{
		{
			name:  "simple text",
			input: "hello",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "hello"},
			},
		},
		{
			name:  "tag open and close",
			input: "[p]hi[/p]",
			expected: []parser.Token{
				{Type: parser.TokenTagOpen, Value: "p"},
				{Type: parser.TokenText, Value: "hi"},
				{Type: parser.TokenTagClose, Value: "p"},
			},
		},
		{
			name:  "text + tag",
			input: "a [b] c",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "a "},
				{Type: parser.TokenTagOpen, Value: "b"},
				{Type: parser.TokenText, Value: " c"},
			},
		},
		{
			name:  "multiple tags",
			input: "[p]text1[/p][i]text2[/i]",
			expected: []parser.Token{
				{Type: parser.TokenTagOpen, Value: "p"},
				{Type: parser.TokenText, Value: "text1"},
				{Type: parser.TokenTagClose, Value: "p"},
				{Type: parser.TokenTagOpen, Value: "i"},
				{Type: parser.TokenText, Value: "text2"},
				{Type: parser.TokenTagClose, Value: "i"},
			},
		},
		{
			name:  "ref tag",
			input: "see [ref]word[/ref] here",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "see "},
				{Type: parser.TokenTagOpen, Value: "ref"},
				{Type: parser.TokenText, Value: "word"},
				{Type: parser.TokenTagClose, Value: "ref"},
				{Type: parser.TokenText, Value: " here"},
			},
		},
		{
			name:  "example tag",
			input: "[ex]example text[/ex]",
			expected: []parser.Token{
				{Type: parser.TokenTagOpen, Value: "ex"},
				{Type: parser.TokenText, Value: "example text"},
				{Type: parser.TokenTagClose, Value: "ex"},
			},
		},
		{
			name:  "star tag - self closing style",
			input: "[*]marker",
			expected: []parser.Token{
				{Type: parser.TokenTagOpen, Value: "*"},
				{Type: parser.TokenText, Value: "marker"},
			},
		},
		{
			name:  "container tag",
			input: "[c]content[/c]",
			expected: []parser.Token{
				{Type: parser.TokenTagOpen, Value: "c"},
				{Type: parser.TokenText, Value: "content"},
				{Type: parser.TokenTagClose, Value: "c"},
			},
		},
		{
			name:  "chinese characters",
			input: "三比西河",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "三比西河"},
			},
		},
		{
			name:  "chinese + pinyin",
			input: "三比西河 sānbǐxīhé",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "三比西河 sānbǐxīhé"},
			},
		},
		{
			name:  "pinyin with tones",
			input: "sānbǐxīhé",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "sānbǐxīhé"},
			},
		},
		{
			name:  "pinyin with smart quote",
			input: "tǔ'ěrqísītǎn",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "tǔ'ěrqísītǎn"},
			},
		},
		{
			name:  "pinyin with right single quote",
			input: "tǔ'ěrqísītǎn",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "tǔ'ěrqísītǎn"},
			},
		},
		{
			name:  "unclosed tag - becomes text",
			input: "[unclosed text",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "[unclosed text"},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:  "only tags",
			input: "[p][/p]",
			expected: []parser.Token{
				{Type: parser.TokenTagOpen, Value: "p"},
				{Type: parser.TokenTagClose, Value: "p"},
			},
		},
		{
			name:  "russian text",
			input: "река Замбези",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "река Замбези"},
			},
		},
		{
			name:  "mixed languages",
			input: "三比西河 река sānbǐxīhé",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "三比西河 река sānbǐxīhé"},
			},
		},
		{
			name:  "newline in text",
			input: "line1\nline2",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "line1\nline2"},
			},
		},
		{
			name:  "DSL entry format",
			input: "三比西河\n sānbǐxīhé\n [m1]text[/m]",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "三比西河\n sānbǐxīhé\n "},
				{Type: parser.TokenTagOpen, Value: "m1"},
				{Type: parser.TokenText, Value: "text"},
				{Type: parser.TokenTagClose, Value: "m"},
			},
		},
		{
			name:  "nested-like structure",
			input: "[p]text [i]italic[/i] more[/p]",
			expected: []parser.Token{
				{Type: parser.TokenTagOpen, Value: "p"},
				{Type: parser.TokenText, Value: "text "},
				{Type: parser.TokenTagOpen, Value: "i"},
				{Type: parser.TokenText, Value: "italic"},
				{Type: parser.TokenTagClose, Value: "i"},
				{Type: parser.TokenText, Value: " more"},
				{Type: parser.TokenTagClose, Value: "p"},
			},
		},
		{
			name:  "brackets treated as tag delimiters",
			input: "text with [brackets] inside",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "text with "},
				{Type: parser.TokenTagOpen, Value: "brackets"},
				{Type: parser.TokenText, Value: " inside"},
			},
		},
		{
			name:  "DSL header comment",
			input: "#NAME \"test\"",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "#NAME \"test\""},
			},
		},
		{
			name:  "multiple spaces preserved",
			input: "a  b  c",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "a  b  c"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Lex(tt.input)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestLexStream(t *testing.T) {
	input := strings.NewReader("a [b] c")

	ch := make(chan parser.Token)
	go parser.LexStream(input, ch)

	var tokens []parser.Token
	for tok := range ch {
		tokens = append(tokens, tok)
	}

	if len(tokens) == 0 {
		t.Error("no tokens produced")
	}

	expected := []parser.Token{
		{Type: parser.TokenText, Value: "a "},
		{Type: parser.TokenTagOpen, Value: "b"},
		{Type: parser.TokenText, Value: " c"},
	}

	if !reflect.DeepEqual(tokens, expected) {
		t.Errorf("expected %+v, got %+v", expected, tokens)
	}
}

func TestLexStreamChinese(t *testing.T) {
	input := strings.NewReader("三比西河 sānbǐxīhé [m1]текст[/m]")

	ch := make(chan parser.Token)
	go parser.LexStream(input, ch)

	var tokens []parser.Token
	for tok := range ch {
		tokens = append(tokens, tok)
	}

	if len(tokens) != 4 {
		t.Errorf("expected 4 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != parser.TokenText {
		t.Errorf("first token should be TEXT")
	}
}

func TestLexUnclosedTag(t *testing.T) {
	result := parser.Lex("[tag not closed")

	if len(result) != 1 {
		t.Errorf("expected 1 token for unclosed tag, got %d", len(result))
	}

	if result[0].Type != parser.TokenText {
		t.Errorf("unclosed tag should become TEXT")
	}
}

func TestLexMultipleTagsSameLine(t *testing.T) {
	result := parser.Lex("[p]a[/p][i]b[/i][ref]c[/ref]")

	expected := []parser.Token{
		{Type: parser.TokenTagOpen, Value: "p"},
		{Type: parser.TokenText, Value: "a"},
		{Type: parser.TokenTagClose, Value: "p"},
		{Type: parser.TokenTagOpen, Value: "i"},
		{Type: parser.TokenText, Value: "b"},
		{Type: parser.TokenTagClose, Value: "i"},
		{Type: parser.TokenTagOpen, Value: "ref"},
		{Type: parser.TokenText, Value: "c"},
		{Type: parser.TokenTagClose, Value: "ref"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}

func TestLexSpecialPinyin(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"ü with diaeresis", "lǖ"},
		{"all tones", "āáǎà ēéěè īíǐì ōóǒò ūúǔù ǖǘǚǜ"},
		{"apostrophe", "d'"},
		{"smart quote", "tǔ'ěrqísītǎn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Lex(tt.input)
			if len(result) != 1 || result[0].Type != parser.TokenText {
				t.Errorf("pinyin should be a single TEXT token")
			}
		})
	}
}

func TestLexTokenType(t *testing.T) {
	result := parser.Lex("[p]text[/p]")

	if result[0].Type != parser.TokenTagOpen {
		t.Errorf("first token should be TokenTagOpen, got %v", result[0].Type)
	}
	if result[1].Type != parser.TokenText {
		t.Errorf("middle token should be TokenText, got %v", result[1].Type)
	}
	if result[2].Type != parser.TokenTagClose {
		t.Errorf("last token should be TokenTagClose, got %v", result[2].Type)
	}
}

func TestLexPreservesStructure(t *testing.T) {
	input := "[m]заголовок[/m]"
	result := parser.Lex(input)

	if len(result) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(result))
	}

	if result[0].Value != "m" {
		t.Errorf("expected tag 'm', got '%s'", result[0].Value)
	}
	if result[1].Value != "заголовок" {
		t.Errorf("expected text 'заголовок', got '%s'", result[1].Value)
	}
}

func TestLexIntegration_DSLEntry(t *testing.T) {
	input := "三比西河\n sānbǐxīhé\n [m1]река Замбези ([i]Южная Африка[/i])[/m]"
	result := parser.Lex(input)

	if len(result) < 6 {
		t.Fatalf("expected at least 6 tokens, got %d", len(result))
	}

	if result[0].Type != parser.TokenText {
		t.Errorf("first token should be TEXT")
	}

	tagOpenFound := false
	for _, tok := range result {
		if tok.Type == parser.TokenTagOpen && tok.Value == "m1" {
			tagOpenFound = true
			break
		}
	}
	if !tagOpenFound {
		t.Errorf("tag [m1] not found in tokens")
	}
}

func TestLexIntegration_RealWorldExamples(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "geographical entry",
			input: "三马林达\n sānmǎlíndá\n [m1][p]г.[/p] Самаринда ([i]Индонезия[/i])[/m]",
		},
		{
			name:  "historical term",
			input: "三摩呾咤\n sānmódázhà\n [m1][p]ист.[/p] Саматата ([i]государство в Вост. Индии[/i])[/m]",
		},
		{
			name:  "reference entry",
			input: "上海市\n shànghǎi shì\n [m1][p]см.[/p] [ref]上海[/ref][/m]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Lex(tt.input)
			if len(result) == 0 {
				t.Errorf("no tokens produced for %s", tt.name)
			}
		})
	}
}

func TestLexEdgeCase_MultipleSpaces(t *testing.T) {
	result := parser.Lex("word1   word2")
	if len(result) != 1 {
		t.Errorf("expected 1 token, got %d", len(result))
	}
	if result[0].Value != "word1   word2" {
		t.Errorf("expected spaces preserved, got %q", result[0].Value)
	}
}

func TestLexEdgeCase_OnlyNumbers(t *testing.T) {
	result := parser.Lex("123 456")
	if len(result) != 1 {
		t.Errorf("expected 1 token, got %d", len(result))
	}
}

func TestLexEdgeCase_Cyrillic(t *testing.T) {
	result := parser.Lex("река Замбези")
	if len(result) != 1 {
		t.Errorf("expected 1 token, got %d", len(result))
	}
	if result[0].Value != "река Замбези" {
		t.Errorf("expected cyrillic text, got %q", result[0].Value)
	}
}

func TestLexEdgeCase_UnicodePunctuation(t *testing.T) {
	result := parser.Lex("text: «quoted» – dash")
	if len(result) != 1 {
		t.Errorf("expected 1 token, got %d", len(result))
	}
}

func TestLexEdgeCase_Emoji(t *testing.T) {
	result := parser.Lex("word 😊 more")
	if len(result) != 1 {
		t.Errorf("expected 1 token, got %d", len(result))
	}
}

func TestLex_CRLFLineEndings(t *testing.T) {
	result := parser.Lex("line1\r\nline2")
	if len(result) != 1 {
		t.Errorf("expected 1 token, got %d", len(result))
	}
}

func TestLex_TabCharacters(t *testing.T) {
	result := parser.Lex("word1\tword2")
	if len(result) != 1 {
		t.Errorf("expected 1 token, got %d", len(result))
	}
}

func TestLexStream_Basic(t *testing.T) {
	input := strings.NewReader("hello world")
	ch := make(chan parser.Token)
	go parser.LexStream(input, ch)

	var tokens []parser.Token
	for tok := range ch {
		tokens = append(tokens, tok)
	}

	if len(tokens) != 1 {
		t.Errorf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != parser.TokenText {
		t.Errorf("expected TEXT token")
	}
	if tokens[0].Value != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", tokens[0].Value)
	}
}

func TestLexStream_WithTags(t *testing.T) {
	input := strings.NewReader("[p]text[/p]")
	ch := make(chan parser.Token)
	go parser.LexStream(input, ch)

	var tokens []parser.Token
	for tok := range ch {
		tokens = append(tokens, tok)
	}

	if len(tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(tokens))
	}

	expected := []struct {
		Type  parser.TokenType
		Value string
	}{
		{parser.TokenTagOpen, "p"},
		{parser.TokenText, "text"},
		{parser.TokenTagClose, "p"},
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.Type {
			t.Errorf("token %d: expected type %v, got %v", i, exp.Type, tokens[i].Type)
		}
		if tokens[i].Value != exp.Value {
			t.Errorf("token %d: expected value '%s', got '%s'", i, exp.Value, tokens[i].Value)
		}
	}
}

func TestLexStream_ChineseAndPinyin(t *testing.T) {
	input := strings.NewReader("三比西河\n sānbǐxīhé\n [m1]текст[/m]")
	ch := make(chan parser.Token)
	go parser.LexStream(input, ch)

	var tokens []parser.Token
	for tok := range ch {
		tokens = append(tokens, tok)
	}

	if len(tokens) != 4 {
		t.Errorf("expected 4 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != parser.TokenText {
		t.Errorf("first token should be TEXT")
	}

	if tokens[1].Type != parser.TokenTagOpen {
		t.Errorf("second token should be OPEN")
	}
}

func TestLexStream_ChannelClosed(t *testing.T) {
	input := strings.NewReader("test")
	ch := make(chan parser.Token)

	done := make(chan bool)
	go func() {
		go parser.LexStream(input, ch)
		count := 0
		for range ch {
			count++
		}
		if count != 1 {
			t.Errorf("expected 1 token, got %d", count)
		}
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("timeout waiting for channel to close")
	}
}

func TestLexStream_EmptyInput(t *testing.T) {
	input := strings.NewReader("")
	ch := make(chan parser.Token)
	go parser.LexStream(input, ch)

	count := 0
	for range ch {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0 tokens for empty input, got %d", count)
	}
}

func TestLexStream_MultipleTags(t *testing.T) {
	input := strings.NewReader("[p]a[/p][i]b[/i]")
	ch := make(chan parser.Token)
	go parser.LexStream(input, ch)

	var tokens []parser.Token
	for tok := range ch {
		tokens = append(tokens, tok)
	}

	if len(tokens) != 6 {
		t.Errorf("expected 6 tokens, got %d", len(tokens))
	}
}

func TestLexStream_LargeInput(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString("[m]текст ")
	}
	sb.WriteString("[/m]")

	input := strings.NewReader(sb.String())
	ch := make(chan parser.Token)

	done := make(chan int)
	go func() {
		go parser.LexStream(input, ch)
		count := 0
		for range ch {
			count++
		}
		done <- count
	}()

	select {
	case count := <-done:
		if count == 0 {
			t.Error("expected tokens, got none")
		}
	case <-time.After(5 * time.Second):
		t.Error("timeout processing large input")
	}
}
