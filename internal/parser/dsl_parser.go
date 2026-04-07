package parser

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type DSLReader struct{
	file *os.File
	reader *transform.Reader
}

func (r *DSLReader) Read(p []byte)(n int, err error){
	return r.reader.Read(p)
}

func (r *DSLReader) Close() error {
    return r.file.Close()
}

func OpenDSL(path string) (io.ReadCloser, error){
	file, err := os.Open(path)
	if err != nil {
        return nil, fmt.Errorf("не удалось открыть файл: %w", err)
    }
    
    decoder := unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewDecoder()
    reader := transform.NewReader(file, decoder)
    
    return &DSLReader{file:file, reader: reader},  nil
}


// reads entire file into memory (legacy, for backward compatibility)
func ParseStream(r io.Reader, limit int) ([]Entry, error) {
	tokens := make([]Token, 0)
	ch := make(chan Token, 1024)

	go LexStream(r, ch)

	for tok := range ch {
		tokens = append(tokens, tok)
	}

	root := Parse(tokens)
	return ExtractEntries(root, limit), nil
}

func ParseDSL(path string, limit int) ([]Entry, error) {
	data, err := ReadDSL(path)
	if err != nil {
		return nil, err
	}
	return ParseStream(strings.NewReader(data), limit)
}

func ParseDSLString(content string, limit int) ([]Entry, error) {
	return ParseStream(strings.NewReader(content), limit)
}

// reads entire file into memory (legacy, for backward compatibility)
// return Raw String from DSL
func ReadDSL(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close()

	// Декодер для UTF-16 LE (формат DSL файлов)
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewDecoder()
	reader := transform.NewReader(file, decoder)

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
