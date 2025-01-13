import (
	"fmt"
	"errors"
)

type LanguageParser interface {
	parse() error
	nextLeaf() *BasicASTLeaf, error
}

