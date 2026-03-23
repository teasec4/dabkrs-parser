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

    for i < len(input) {
    	// check if it's tag
        if input[i] == '[' {
        	// j - end of tag
            j := strings.IndexByte(input[i:], ']')
            if j == -1 {
                break
            }
            
            // inside of [] we got tag
            tag := input[i+1 : i+j]
            
            // check if tag is open or close
            if strings.HasPrefix(tag, "/") {
                tokens = append(tokens, Token{
                    Type:  TokenTagClose,
                    Value: tag[1:],
                })
            } else {
                tokens = append(tokens, Token{
                    Type:  TokenTagOpen,
                    Value: tag,
                })
            }
            
            // move forward to body of tag
            i += j + 1
        } else {
        	// body inside tag
            j := strings.IndexByte(input[i:], '[')
            if j == -1 {
                j = len(input) - i
            }

            tokens = append(tokens, Token{
                Type:  TokenText,
                Value: input[i : i+j],
            })

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