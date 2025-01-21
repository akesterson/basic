package main

import (
	"os"
	//"strings"
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
	runtime.run(os.Stdin, MODE_REPL)
	//runtime.run(strings.NewReader("10 FOR A# = 1 TO 5\n20 PRINT A#\n30 NEXT A#\nRUN\nQUIT\n"), MODE_REPL)
}
