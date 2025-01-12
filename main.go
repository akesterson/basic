package main

import (
	"fmt"
	/*"os"*/
)

type BasicError int
const (
	IO           BasicError = iota
	PARSE
	SYNTAX
	EXECUTE	
)

type BasicType int
const (
	INTEGER      BasicType = iota
	STRING
)

type BasicLiteral struct {
	literaltype BasicType
	stringval string
	intval int
}

type BasicToken struct {
	tokentype BasicTokenType
	lineno int
	literal string
	lexeme string
}

func (self BasicToken) toString() string {
	return fmt.Sprintf("%d %s %s", self.tokentype, self.lexeme, self.literal)
}

type BasicContext struct {
	source [9999]string
	lineno int
}

func (self BasicContext) init() {
	self.lineno = 0
}

func errorCodeToString(errno BasicError) string {
	switch (errno) {
	case IO: return "IO ERROR"
	case PARSE: return "PARSE ERROR"
	case EXECUTE: return "EXEC ERROR"
	case SYNTAX: return "SYNTAX ERROR"
	}
	return "UNDEF"
}

func basicError(line int, errno BasicError, message string) {
	fmt.Printf("? %s %s", errorCodeToString(errno), message)
}

func main() {
	var context BasicContext;
	var scanner BasicScanner;
	context.init()
	scanner.init(context)
	//scanner.repl(os.Stdin)
	scanner.scanTokens("10 PRINT \"HELLO\"")
	scanner.scanTokens("20 ABC#=3+2")
	scanner.scanTokens("30 XYZ%=(3+(4*5))")
	scanner.scanTokens("40 PRINT# = 123456")
	scanner.scanTokens("40 REM THIS IS A COMMENT !!!!")
	scanner.scanTokens("50 ABC# = (XYZ% * ABC#)")
	scanner.scanTokens("60 PRINT ABC#")
}
