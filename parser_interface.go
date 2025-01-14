package main

import (
	"fmt"
	"errors"
)

type LanguageParser interface {
	parse() error
	nextLeaf() (*BasicASTLeaf, error)
	getToken(idx int) (*BasicToken, error)
	addToken(idx int) 
}

