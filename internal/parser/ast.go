package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

type NodeType int

const (
	NodeText      NodeType = iota
	NodeParagraph          // [p]
	NodeItalic             // [i]
	NodeRef                // [ref]
	NodeContainer          // [c]
	NodeExample            // [ex]
	NodeStar               // [*]
	NodeMeaning            // [m1], [m2], [m3]...
	NodeUnknown
)

type Node struct {
	Type     NodeType
	Value    string
	Level    int
	Children []*Node
}

func tagToNodeType(tag string) NodeType {
	tag = strings.ToLower(tag)
	switch tag {
	case "p":
		return NodeParagraph
	case "i":
		return NodeItalic
	case "c":
		return NodeContainer
	case "ref":
		return NodeRef
	case "ex":
		return NodeExample
	case "*":
		return NodeStar
	case "m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8", "m9":
		return NodeMeaning
	default:
		return NodeUnknown
	}
}

func Parse(tokens []Token) *Node {
	root := &Node{Type: NodeContainer}
	stack := []*Node{root}

	for _, tok := range tokens {
		switch tok.Type {
		case TokenText:
			node := &Node{
				Type:  NodeText,
				Value: tok.Value,
			}
			current := stack[len(stack)-1]
			current.Children = append(current.Children, node)

		case TokenTagOpen:
			nodeType := tagToNodeType(tok.Value)
			level := 0
			if strings.HasPrefix(tok.Value, "m") {
				fmt.Sscanf(tok.Value, "m%d", &level)
			}

			node := &Node{
				Type:  nodeType,
				Value: tok.Value,
				Level: level,
			}

			current := stack[len(stack)-1]
			current.Children = append(current.Children, node)
			stack = append(stack, node)

		case TokenTagClose:
			if len(stack) > 1 {
				top := stack[len(stack)-1]

				topTag := top.Value
				closeTag := tok.Value

				// Check for exact match first (case-insensitive)
				if strings.EqualFold(topTag, closeTag) {
					stack = stack[:len(stack)-1]
					continue
				}

				// Handle m/m1/m2 style tags - allow generic "m" to close "mN"
				if strings.HasPrefix(topTag, "m") && strings.HasPrefix(closeTag, "m") {
					topNum := extractNum(topTag)
					closeNum := extractNum(closeTag)
					// "m" matches any "mN", or exact number match
					if (closeNum == 0 && topNum > 0) || (topNum == closeNum && topNum > 0) {
						stack = stack[:len(stack)-1]
					}
				}
			}
		}
	}
	return root
}

func extractNum(tag string) int {
	num := 0
	for _, c := range tag {
		if c >= '0' && c <= '9' {
			num = num*10 + int(c-'0')
		}
	}
	return num
}

func (n NodeType) String() string {
	switch n {
	case NodeText:
		return "TEXT"
	case NodeParagraph:
		return "PARAGRAPH"
	case NodeItalic:
		return "ITALIC"
	case NodeRef:
		return "REF"
	case NodeContainer:
		return "CONTAINER"
	case NodeExample:
		return "EXAMPLE"
	case NodeStar:
		return "STAR"
	case NodeMeaning:
		return "MEANING"
	default:
		return "UNKNOWN"
	}
}

func DumpAST(root *Node) string {
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}
