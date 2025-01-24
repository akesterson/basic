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
	if ( len(os.Args) > 1 ) {
		f, err := os.Open(os.Args[1])
		if ( err != nil ) {
			panic(err.Error())
		}
		defer f.Close()
		runtime.run(f, MODE_RUNSTREAM)
	} else {
		runtime.run(os.Stdin, MODE_REPL)
	}
}
