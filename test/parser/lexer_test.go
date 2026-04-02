package parser_test

import (
	"parser/internal/parser"
	"reflect"
	"strings"
	"testing"
)

func TestLex_Simple(t *testing.T){
	input := "hello [p]world[/p]"
	
	tokens := parser.Lex(input)
	
	expected := []parser.Token{
		{Type: parser.TokenText, Value: "hello "},
        {Type: parser.TokenTagOpen, Value: "p"},
        {Type: parser.TokenText, Value: "world"},
        {Type: parser.TokenTagClose, Value: "p"},
	}
	
	if !reflect.DeepEqual(tokens, expected) {
        t.Errorf("expected %+v, got %+v", expected, tokens)
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