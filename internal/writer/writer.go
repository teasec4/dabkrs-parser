package writer

import (
	"fmt"
	"io"
	"parser/internal/parser"
	"strings"
)

func WriteAST(node *parser.Node, depth int, w io.Writer) error {
	prefix := strings.Repeat("  ", depth)

	var line string

	switch node.Type {
	case parser.NodeText:
		line = fmt.Sprintf("%s%s: %s\n", prefix, node.Type, node.Value)
	case parser.NodeRef:
		line = fmt.Sprintf("%s%s: %s\n", prefix, node.Type, node.Value)
	default:
		line = fmt.Sprintf("%s%s\n", prefix, node.Type)
	}

	_, err := w.Write([]byte(line))
	if err != nil {
		return err
	}

	for _, child := range node.Children {
		if err := WriteAST(child, depth+1, w); err != nil {
			return err
		}
	}

	return nil
}