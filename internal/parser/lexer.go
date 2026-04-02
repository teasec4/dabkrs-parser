package parser

import (
	"bufio"
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

func Lex(input string) []Token {
	var tokens []Token
	i := 0
	n := len(input)

	for i < n {
		if input[i] == '[' {
			j := strings.IndexByte(input[i:], ']')
			if j == -1 {
				// незакрытый тег — добавляем как текст
				tokens = append(tokens, Token{Type: TokenText, Value: input[i:]})
				break
			}
			tagContent := input[i+1 : i+j]

			if strings.HasPrefix(tagContent, "/") {
				tokens = append(tokens, Token{
					Type:  TokenTagClose,
					Value: strings.TrimPrefix(tagContent, "/"),
				})
			} else {
				tokens = append(tokens, Token{
					Type:  TokenTagOpen,
					Value: tagContent,
				})
			}
			i += j + 1
		} else {
			// текст до следующего [
			j := strings.IndexByte(input[i:], '[')
			if j == -1 {
				j = n - i
			}
			text := input[i : i+j]
			if text != "" {
				tokens = append(tokens, Token{Type: TokenText, Value: text})
			}
			i += j
		}
	}

	return tokens
}

func LexStream(r io.Reader, out chan <- Token) {
	reader := bufio.NewReader(r)

	for {
		ch, _, err := reader.ReadRune()
		if err == io.EOF {
			break
		}

		if ch == '[' {
			tag, _ := reader.ReadString(']')
			tag = strings.TrimSuffix(tag, "]")
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
            
            out <- Token{Type: TokenText, Value: sb.String()}
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
