package parser

import (
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
	Level int
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
				Level: level,
			
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
			}

			current := stack[len(stack)-1]
			current.Children = append(current.Children, node)
			stack = append(stack, node)

		case TokenTagClose:
			if len(stack) > 1 {
	        top := stack[len(stack)-1]
	
		        if strings.EqualFold(top.Value, tok.Value) {
		            stack = stack[:len(stack)-1]
		        } else {
		            // ❗ несоответствие тегов
		            // можно:
		            // 1. игнорировать
		            // 2. логировать
		            // 3. пытаться чинить стек
		        }
		    }
		}
	}
	return root
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
