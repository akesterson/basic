package main

import (
	"os"
	//"fmt"
	//"strings"
	//"unsafe"
	"io"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
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
	var window *sdl.Window
	var font *ttf.Font
	//var surface *sdl.Surface
	//var text *sdl.Surface

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if ( err != nil ) {
		panic(err)
	}
	defer sdl.Quit()

	err = ttf.Init()
	if ( err != nil ) {
		panic(err)
	}

	window, err = sdl.CreateWindow(
		"BASIC",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		800, 600,
		sdl.WINDOW_SHOWN)
	if ( err != nil ) {
		return
	}
	defer window.Destroy()

	//if surface, err = window.GetSurface(); err != nil {
	//	return
	//}

	// Load the font for our text
	font, err = ttf.OpenFont("./fonts/C64_Pro_Mono-STYLE.ttf", 16)
	if ( err != nil ) {
		return
	}
	defer font.Close()

	runtime.init(window, font)
	
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
