package parser_test

import (
	"parser/internal/parser"
	"reflect"
	"strings"
	"testing"
)

func TestLex(t *testing.T){
	tests := []struct{
		name string
		input string
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
	}
	
	for _, tt := range tests{
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
}