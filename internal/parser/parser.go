package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type Header struct{
	Hanzi string
	Pinyin string
}

type RawMeaning struct{
	Level int
	Raw string
	Order int
}

type DebugEntry struct{
	Hanzi string
	Pinyin string
	Meanings []RawMeaning
}

// catch Caracters and Pinyin
var entryStartRe = regexp.MustCompile(
    `([\p{Han}_]+)\s*([a-zA-ZāáǎàēéěèīíǐìōóǒòūúǔùǖǘǚǜüÜ\s]+?)\[m1]`,
)

// catch meaning 
var meaningRe = regexp.MustCompile(`\[m(\d+)](.*?)\[/m]`)

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
	
	// CleanString to avoid "\n"
	return Header{
		Hanzi: match[1],
		Pinyin: CleanString(match[2]),
	}, nil
}

func ExtractMeaning(entry string) []RawMeaning{
	mathes := meaningRe.FindAllStringSubmatch(entry, -1)
	
	var result []RawMeaning
	
	for i, m := range mathes{
		level, _ := strconv.Atoi(m[1])
		
		result = append(result, RawMeaning{
			Level: level,
			Raw: strings.TrimSpace(m[2]),
			Order: i,
		})
	}
	
	return result
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
    
    
    var entriesToJSON []DebugEntry
    
    for i, e := range entries{
    	// create Entiry
     	var entry DebugEntry
    	// get Header
    	header, err := ExtractHeader(e)
    	if err != nil {
            fmt.Println("skip:", err)
            continue
        }
        
        entry.Hanzi = header.Hanzi
        entry.Pinyin = header.Pinyin
        
        fmt.Printf("%d: %s | %s\n", i, header.Hanzi, header.Pinyin)
        
        // get Meaning
        meanings := ExtractMeaning(e)
        for _, m := range meanings{
       		fmt.Printf(" m%d: %s\n", m.Level, m.Raw)
        }
        
        entry.Meanings = meanings
        
        entriesToJSON = append(entriesToJSON, entry)
        
        if i > 20 {
            break
        }
        
    }
    
    for _, entry := range entriesToJSON{
    	for _, m := range entry.Meanings{
     		fmt.Println("====Raw====")
     		fmt.Println(m.Raw)
       
       		tokens := Lex(m.Raw)
         
         	for _, t := range tokens{
        		fmt.Printf("%s: %q\n", t.Type, t.Value)
          }
     }
    }
    
    data, _ := json.MarshalIndent(entriesToJSON, "", " ")
    os.WriteFile("debug.json", data, 0644)
    
    return nil
}

func CleanString(s string) string {
    return strings.TrimSpace(strings.ReplaceAll(s, "\n", ""))
}


