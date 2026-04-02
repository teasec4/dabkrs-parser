package parser

import (
	"fmt"
	"io"
	"os"
	
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// return Raw String from DSL
func ReadDSL(path string)(string, error){
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



