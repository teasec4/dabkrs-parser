package parser

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type TokenType int

const (
	TokenText     TokenType = iota
	TokenTagOpen            // [p]
	TokenTagClose           // [/p]
)

type Token struct {
	Type  TokenType
	Value string
}

func Lex(s string) []Token {
	ch := make(chan Token, 256)
	go LexStream(strings.NewReader(s), ch)

	var tokens []Token
	for tok := range ch {
		tokens = append(tokens, tok)
	}
	return tokens
}

func LexStream(r io.Reader, out chan<- Token) {
	reader := bufio.NewReader(r)

	for {
		ch, _, err := reader.ReadRune()
		if err == io.EOF {
			break
		}

		if ch == '[' {
			tag, err := reader.ReadString(']')
			if err != nil {
				out <- Token{Type: TokenText, Value: "[" + tag}
				continue
			}
			tag = strings.TrimSuffix(tag, "]")

			if tag == "" {
				out <- Token{Type: TokenText, Value: "[]"}
				continue
			}

			if strings.HasPrefix(tag, "/") {
				out <- Token{Type: TokenTagClose, Value: tag[1:]}
			} else {
				out <- Token{Type: TokenTagOpen, Value: tag}
			}
		} else {
			var sb strings.Builder
			sb.WriteRune(ch)

			for {
				ch, _, err := reader.ReadRune()
				if err != nil || ch == '[' {
					if ch == '[' {
						reader.UnreadRune()
					}
					break
				}
				sb.WriteRune(ch)
			}

			text := sb.String()
			if strings.TrimSpace(text) != "" {
				out <- Token{Type: TokenText, Value: text}
			}
		}
	}

	close(out)
}

func (t TokenType) String() string {
	switch t {
	case TokenText:
		return "TEXT"
	case TokenTagOpen:
		return "OPEN"
	case TokenTagClose:
		return "CLOSE"
	default:
		return "UNKNOWN"
	}
}

func DumpTokens(tokens []Token) string {
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}
