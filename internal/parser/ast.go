package parser

import (
	"fmt"
	"strings"
)

type NodeType int

const (
    NodeText NodeType = iota
    NodeParagraph   // p
    NodeItalic      // i
    NodeRef         // ref
    NodeContainer   // c
    NodeExample     // ex
    NodeUnknown
)

type Node struct {
    Type     NodeType
    Value    string      // только для текста
    Children []*Node
}

func tagToNodeType(tag string) NodeType {
    switch tag {
    case "p":
        return NodeParagraph
    case "i":
        return NodeItalic
    case "ref":
        return NodeRef
    case "c":
        return NodeContainer
    case "ex":
        return NodeExample
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

            current := stack[len(stack)-1]
            current.Children = append(current.Children, node)

            
            stack = append(stack, node)

        case TokenTagClose:
            if len(stack) > 1 {
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
    default:
        fmt.Printf("%sNODE: %v\n", prefix, node.Type)
    }

    for _, child := range node.Children {
        PrintAST(child, indent+1)
    }
}