package parser

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