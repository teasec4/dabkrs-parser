package main

import (
	"encoding/json"
	"fmt"

	"os"
	"parser/internal/parser"
	"parser/internal/storage"
)

func main(){
	// stream flag
	byStream := true
	
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
	
	if !byStream{
		entries := parser.ExtractEntries(ast, 100)

		err = SaveJSON("output.json", entries)
		if err != nil {
		    panic(err)
		}
		
		return
	}
	
	err = storage.StreamEntiresToJSON(ast, "output_stream.json", 100)
	if err != nil {
	    panic(err)
	}
	
}

func SaveJSON(filename string, data any) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")

    return encoder.Encode(data)
}