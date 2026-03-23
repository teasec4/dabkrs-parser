package parser

type NodeType string

const(
	TextNode NodeType = "text"
	TagNode NodeType = "tag"
)

type Node struct{
	Type NodeType
	Tag string
	Content string
	Children []Node
}