package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("No File Passed")
		os.Exit(2)
	}

	jsonFile, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	lexer := NewLexer(jsonFile)
	json := make(map[string]interface{})

	pos, tok, _ := lexer.NextToken()
	if tok != OpenBrace {
		fmt.Fprintf(os.Stderr, "Expected '{' at line %d column %d", pos.line, pos.column)
	}

	for {
		pos, tok, key := lexer.NextToken()
		if tok == EOF {
			fmt.Fprintf(os.Stderr, "Unexpected EOF")
			os.Exit(1)
		}

		if tok != String {
			fmt.Fprintf(os.Stderr, "Invalid key at line %d column %d got %v", pos.line, pos.column, key)
			os.Exit(1)
		}

		_, delimiter, _ := lexer.NextToken()
		if delimiter != Colon {
			fmt.Fprintf(os.Stderr, "Expected ':' at line %d column %d", pos.line, pos.column)
			os.Exit(1)
		}

		pos, tok, value := lexer.NextToken()
		switch tok {
		case String:
			json[key] = value
		case Number:
			json[key], err = strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing number at line %d column %d", pos.line, pos.column)
				os.Exit(1)

			}
		case Bool:
			if value == "true" {
				json[key] = true
			} else if value == "false" {
				json[key] = false
			} else {
				fmt.Println(value)
				fmt.Fprintf(os.Stderr, "Error parsing boolean at line %d column %d", pos.line, pos.column)
				os.Exit(1)
			}
		case Null:
			if value != "null" {
				fmt.Fprintf(os.Stderr, "Error parsing null at line %d column %d", pos.line, pos.column)
			}
			json[key] = nil

		}

		_, end, _ := lexer.NextToken()
		if end == CloseBrace {
			break
		} else if end != Comma {
			fmt.Fprintf(os.Stderr, "Expected ',' or '}' at line %d column %d", pos.line, pos.column)
			os.Exit(1)
		}
	}

	fmt.Println(json)
}
