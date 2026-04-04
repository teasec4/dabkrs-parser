package parser

import (
	"fmt"
	"strings"
)

type NodeType int

const (
	NodeText      NodeType = iota
	NodeParagraph          // p
	NodeItalic             // i
	NodeRef                // ref
	NodeContainer          // c
	NodeExample            // ex
	NodeStar               // [*]  (используется для примеров/маркеров)
	NodeMeaning            // m1, m2, m3... (для значений)
	NodeUnknown
)

type Node struct {
	Type     NodeType
	Value    string // только для текста
	Children []*Node
}

func tagToNodeType(tag string) NodeType {
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

			node := &Node{
				Type: nodeType,
			}

			// Специальная обработка для [ref] — сохраняем имя ссылки
			if nodeType == NodeRef {
				node.Value = tok.Value // можно позже добавить атрибуты
			}

			current := stack[len(stack)-1]
			current.Children = append(current.Children, node)
			stack = append(stack, node)

		case TokenTagClose:
			if len(stack) > 1 {
				// Проверяем, что закрывается правильный тег (опционально)
				stack = stack[:len(stack)-1] // pop
			}
		}
	}
	return root
}

func PrintAST(node *Node, indent int) {
	prefix := strings.Repeat("  ", indent)

	switch node.Type {
	case NodeText:
		fmt.Printf("%sTEXT: %q\n", prefix, node.Value)
	case NodeRef:
		fmt.Printf("%sREF: %s\n", prefix, node.Value)
	default:
		fmt.Printf("%s%s\n", prefix, node.Type)
	}

	for _, child := range node.Children {
		PrintAST(child, indent+1)
	}
}

func (n NodeType) String() string {
	switch n {
	case NodeText:
		return "TEXT"
	case NodeParagraph:
		return "PARAGRAPH"
	case NodeItalic:
		return "ITALIC"
	case NodeContainer:
		return "CONTAINER"
	case NodeRef:
		return "REF"
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
