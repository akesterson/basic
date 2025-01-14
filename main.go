package main

import (
	"fmt"
	"os"
)

const (
	MAX_LEAVES = 32
	MAX_TOKENS = 32
	MAX_VALUES = 32
	BASIC_TRUE = -1
	BASIC_FALSE = 0
)

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
	var runtime BasicRuntime;
	var scanner BasicScanner;
	var parser BasicParser;
	runtime.init()
	parser.init(&runtime)
	scanner.init(&runtime, &parser)
	scanner.repl(os.Stdin)

	/*
	var err error;
	var leaf *BasicASTLeaf;
	scanner.scanTokens("10 \"Hello\" < \"World\"")
	leaf, err = parser.parse()
	if ( err != nil ) {
		fmt.Println(fmt.Sprintf("? %s", err))
	}
	if ( leaf != nil ) {
		fmt.Println(fmt.Sprintf("? %s", leaf.toString()))
	}
	runtime.interpret(leaf)
	
	scanner.scanTokens("10 PRINT \"HELLO\"")
	scanner.scanTokens("20 ABC#=3+2")
	scanner.scanTokens("30 XYZ%=(3+(4*5))")
	scanner.scanTokens("40 PRINT# = 123456")
	scanner.scanTokens("40 REM THIS IS A COMMENT !!!!")
	scanner.scanTokens("50 ABC# = (XYZ% * ABC#)")
	scanner.scanTokens("60 PRINT ABC#")

        var exprleaf BasicASTLeaf
	var unaryleaf BasicASTLeaf
	var unaryliteralleaf BasicASTLeaf
	var groupleaf BasicASTLeaf
	var groupleafexpr BasicASTLeaf
	err := unaryliteralleaf.newLiteralInt(123)
	if ( err != nil ) {
		panic(err)
	}
	err = unaryleaf.newUnary(MINUS, &unaryliteralleaf)
	if ( err != nil ) {
		panic(err)
	}
	err = groupleafexpr.newLiteralFloat(45.67)
	if ( err != nil ) {
		panic(err)
	}
	err = groupleaf.newGrouping(&groupleafexpr)
	if ( err != nil ) {
		panic(err)
	}
	err = exprleaf.newBinary(&unaryleaf, STAR, &groupleaf)
	if ( err != nil ) {
		panic(err)
	}
	fmt.Println(exprleaf.toString())
	*/
}
