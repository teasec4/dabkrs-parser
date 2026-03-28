package parser

import "strings"

type TokenType int

const (
    TokenText TokenType = iota
    TokenTagOpen   // [p]
    TokenTagClose  // [/p]
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