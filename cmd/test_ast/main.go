package main

import (
	"bufio"

	"fmt"
	"io"
	"os"
	"strings"

	"parser/internal/parser"
)

const MaxDepth = 5
const MaxNodes = 100

func main() {
	// path to first part of Dictionray
	path := "./dabkrs/dabkrs_1.dsl"
	fmt.Printf("Start parsing, path: %s \n", path)

	// read DSL and return raw string
	data, err := parser.ReadDSL(path)
	if err != nil {
		fmt.Println("Wrong path or something happend to read the file")
		return
	}

	// tokenize the raw string
	tokens := parser.Lex(data)

	// root Node and childresn on List
	ast := parser.Parse(tokens)

	// write to .txt for debuging struct of nodes
	err = WriteToFileWithBUffer("output_ast.txt", ast)
	if err != nil {
		fmt.Errorf("Ошибка записи %w", err)
	}

	fmt.Println("\nFinish parsing")
}

var nodeCount int

func WriteAST(node *parser.Node, depth int, w io.Writer) error {
	prefix := strings.Repeat("	", depth)

	nodeCount++
	if nodeCount > MaxNodes {
		return nil
	}

	line := fmt.Sprintf("%s%s:%s\n", prefix, node.Type, node.Value)

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

func WriteToFileWithBUffer(filename string, ast *parser.Node) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	// create buffer
	buffer := bufio.NewWriterSize(file, 64*1024)

	// log every line
	logger := &LoggingWriter{w: buffer}

	// create writer with log
	writer := io.MultiWriter(logger)

	// write ast
	if err := WriteAST(ast, 0, writer); err != nil {
		return fmt.Errorf("write ast: %w", err)
	}

	// flush
	if err := buffer.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}

	return nil
}

type LoggingWriter struct {
	w io.Writer
}

func (lw *LoggingWriter) Write(p []byte) (int, error) {
	fmt.Printf("[WRITE] %d bytes\n", len(p))
	return lw.w.Write(p)
}
