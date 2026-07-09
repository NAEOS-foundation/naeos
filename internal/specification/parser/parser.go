package parser

import "fmt"

type Parser interface {
	Parse(input string) (*SpecDocument, error)
}

type ParserFunc func(input string) (*SpecDocument, error)

func (f ParserFunc) Parse(input string) (*SpecDocument, error) {
	return f(input)
}

type SpecDocument struct {
	Raw string
}

func NewParser() Parser {
	return ParserFunc(func(input string) (*SpecDocument, error) {
		if input == "" {
			return nil, fmt.Errorf("input cannot be empty")
		}
		return &SpecDocument{Raw: input}, nil
	})
}
