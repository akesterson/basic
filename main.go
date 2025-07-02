package main

import (
	"os"
	//"fmt"
	//"strings"
	//"unsafe"
	"io"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	MAX_LEAVES = 32
	MAX_TOKENS = 32
	MAX_VALUES = 64
	MAX_SOURCE_LINES = 9999
	MAX_ARRAY_DEPTH = 64
	MAX_VARIABLES = 128
	BASIC_TRUE = -1
	BASIC_FALSE = 0
	MODE_REPL = 1
	MODE_RUN = 2
	MODE_RUNSTREAM = 3
	MODE_QUIT = 4
)

func main() {
	var runtime BasicRuntime;

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if ( err != nil ) {
		panic(err)
	}
	defer sdl.Quit()
	runtime.init()
	if ( len(os.Args) > 1 ) {
		f := sdl.RWFromFile(os.Args[1], "r")
		if ( f == nil ) {
			panic(sdl.GetError())
		}
		defer io.Closer.Close(f)
		runtime.run(f, MODE_RUNSTREAM)
	} else {
		runtime.run(os.Stdin, MODE_REPL)
	}
}
