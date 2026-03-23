package main

import (
	"fmt"
	"parser/internal/parser"
)

func main(){
	
	path := "./dabkrs/dabkrs_1.dsl"
	
	fmt.Printf("Start parsing, path: %s", path)
	parser.ParseDSLFile(path)
	fmt.Println("\nFinish parsing")
	
	
}