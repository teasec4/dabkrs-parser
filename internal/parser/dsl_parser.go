package parser

import (
	"io"
)

func ParseDSL(data string) ([]Entry, error) {
	tokens := Lex(data)
	root := Parse(tokens)
	return ExtractEntries(root, 0), nil
}

func ParseFile(path string, limit int) ([]Entry, error) {
	data, err := ReadDSL(path)
	if err != nil {
		return nil, err
	}

	tokens := Lex(data)
	root := Parse(tokens)
	return ExtractEntries(root, limit), nil
}

func ParseStream(r io.Reader, limit int) ([]Entry, error) {
	tokens := make([]Token, 0)
	ch := make(chan Token, 1024)

	go LexStream(r, ch)

	for tok := range ch {
		tokens = append(tokens, tok)
	}

	root := Parse(tokens)
	return ExtractEntries(root, limit), nil
}
