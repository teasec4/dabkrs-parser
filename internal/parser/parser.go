package parser

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type Header struct{
	Hanzi string
	Pinyin string
}

// catch Caracters and Pinyin
var entryStartRe = regexp.MustCompile(
    `([\p{Han}]+)\s*([a-zA-ZāáǎàēéěèīíǐìōóǒòūúǔùǖǘǚǜüÜ\s]+?)\[m1]`,
)

func SplitEntries(context string) []string{
	matches := entryStartRe.FindAllStringIndex(context, -1)
	
	var entries []string
	
	for i := range matches{
		start := matches[i][0]
		
		var end int
		if i+1 < len(matches){
			end = matches[i+1][0]
		} else{
			end = len(context)
		}
		
		entry := strings.TrimSpace(context[start:end])
		entries = append(entries, entry)
	}
	
	return entries
}

func ExtractHeader(entry string)(Header, error){
	match := entryStartRe.FindStringSubmatch(entry)
	if len(match) < 3{
		return Header{}, fmt.Errorf("не удалось распарсить header: %s", entry[:50])
	}
	
	return Header{
		Hanzi: match[1],
		Pinyin: match[2],
	}, nil
}

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

// parse row text for undestanding
func ParseDSLFile(path string)(error){
	content, err := ReadDSL(path)
	if err != nil {
        return err
    }
    
    entries := SplitEntries(content)
    fmt.Println("\n entries:", len(entries))
    
    for i, e := range entries{
    	header, err := ExtractHeader(e)
    	if err != nil {
            fmt.Println("skip:", err)
            continue
        }
        
        fmt.Printf("%d: %s | %s\n", i, header.Hanzi, header.Pinyin)
        
        if i > 20 {
            break
        }
    }
    
    return nil
}