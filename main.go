package main

import (
	//"os"
	"strings"
)

const (
	MAX_LEAVES = 32
	MAX_TOKENS = 32
	MAX_VALUES = 32
	MAX_SOURCE_LINES = 9999
	BASIC_TRUE = -1
	BASIC_FALSE = 0
	MODE_REPL = 1
	MODE_RUN = 2
	MODE_RUNSTREAM = 3
	MODE_QUIT = 4
)

func main() {
	var runtime BasicRuntime;
	runtime.init()
	//runtime.run(os.Stdin, MODE_REPL)
	runtime.run(strings.NewReader("10 IF 1 == 1 THEN PRINT \"HELLO\"\nRUN\nQUIT"), MODE_REPL)
	//runtime.run(strings.NewReader("10 PRINT \"Hello World\"\nRUN\nQUIT"), MODE_RUNSTREAM)
	//runtime.run(strings.NewReader("10 PRINT \"Hello World\"\nRUN\nQUIT"), MODE_REPL)
	//runtime.source[10] = "10 PRINT \"Hello World\""
	//runtime.source[20] = "QUIT"
	//runtime.run(strings.NewReader(""), MODE_RUN)

	/*
	var err error;
	var leaf *BasicASTLeaf;
	scanner.scanTokens("10 PRINT \"Hello, World!\"")
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
