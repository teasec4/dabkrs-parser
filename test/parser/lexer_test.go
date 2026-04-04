package parser_test

import (
	"parser/internal/parser"
	"reflect"
	"strings"
	"testing"
	
)

// list of tests
func TestLexStream(t *testing.T) {
	tests := []struct{
		name string
		input string
		expected []parser.Token
	}{
		{
			name: "simple test",
			input: "hello world",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "hello world"},
			},
		},
		{
			name: "simple tag test",
			input: "[p]inside P tag[/p]",
			expected: []parser.Token{
				{Type: parser.TokenTagOpen, Value: "p"},
				{Type: parser.TokenText, Value: "inside P tag"},
				{Type: parser.TokenTagClose,Value: "p"},
			},
		},
		{
			name: "real example",
			input: "你好 nihao [m]Hello[/m]",
			expected: []parser.Token{
				{Type: parser.TokenText, Value: "你好 nihao "},
				{Type: parser.TokenTagOpen, Value: "m"},
				{Type: parser.TokenText, Value: "Hello"},
				{Type: parser.TokenTagClose,Value: "m"},
			},
		},
	}
	
	// tests
	for _, tt := range tests{
		t.Run(tt.name, func(t *testing.T) {
			// create chanel
			ch := make(chan parser.Token)
			// create reader with test input
			input := strings.NewReader(tt.input)
			
			go parser.LexStream(input, ch)
			
			var tokens []parser.Token
			for tok := range ch {
				tokens = append(tokens, tok)
			}
			
			if !reflect.DeepEqual(tokens, tt.expected){
				t.Errorf("expected %+v, got %+v", tt.expected, tokens)
			}
		})
	}

}


