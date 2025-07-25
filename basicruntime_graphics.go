package main

import (
	"fmt"
	"strings"
	"github.com/veandco/go-sdl2/sdl"
)

func (self *BasicRuntime) advanceCursor(x int32, y int32) error { var
	err error
	self.cursorX += x
	if ( self.cursorX >= self.maxCharsW ) {
		self.cursorX = 0
		self.cursorY += 1
	} else if ( self.cursorX < 0 ) {
		if ( self.cursorY > 0 ) {
			self.cursorY -=1
		}
		self.cursorX = self.maxCharsW
	}
	self.cursorY += y
	if ( self.cursorY >= self.maxCharsH - 1 ) {
		err = self.scrollWindow(0, 1)
		if ( err != nil ) {
			return err
		}
		self.cursorY -= 1
	}
	//fmt.Println("Cursor X, Y : %d, %d", self.cursorX, self.cursorY)
	return nil
}

func (self *BasicRuntime) drawCursor() error {
	return self.drawText(
		(self.cursorX * int32(self.fontWidth)),
		(self.cursorY * int32(self.fontHeight)),
		"_",
		true)
}

func (self *BasicRuntime) drawWrappedText(x int32, y int32, text string) error {
	var err error
	var curslice string
	var curstartidx int32 = 0
	var endidx int32 = 0

	// chop the text up into slices that will fit onto the screen after the cursor
	for ( curstartidx < int32(len(text)) ) {
		endidx = curstartidx + (self.maxCharsW - self.cursorX)
		if ( endidx >= int32(len(text)) ) {
			endidx = int32(len(text))
		}
		curslice = text[curstartidx:endidx]
		//fmt.Printf("Drawing \"%s\"\n", curslice)
		err = self.drawText(x, y, curslice, false)
		self.advanceCursor(int32(len(curslice)), 0)
		x = (self.cursorX * int32(self.fontWidth))
		y = (self.cursorY * int32(self.fontHeight))
		self.window.UpdateSurface()
		if ( err != nil ) {
			return err
		}
		if ( endidx == int32(len(text)) ) {
			break
		}
		curstartidx += int32(len(curslice))
	}
	return nil
}

func (self *BasicRuntime) drawText(x int32, y int32, text string, updateWindow bool) error {
	var windowSurface *sdl.Surface
	var textSurface *sdl.Surface
	var err error

	windowSurface, err = self.window.GetSurface()
	if ( err != nil ) {
		return err
	}

	textSurface, err = self.font.RenderUTF8Shaded(
		text,
		sdl.Color{R: 255, G: 255, B: 255, A: 255},
		sdl.Color{R: 0, G: 0, B: 0, A: 255})
	if ( err != nil ) {
		return err
	}
	defer textSurface.Free()
	
	err = textSurface.Blit(nil,
		windowSurface,
		&sdl.Rect{
			X: x,
			Y: y,
			W: 0,
			H: 0})
	if ( err != nil ) {
		return err
	}
	if ( updateWindow == true ) {
		self.window.UpdateSurface()
	}		
	return nil
}

func (self *BasicRuntime) scrollWindow(x int32, y int32) error {
	var err error
	var windowSurface *sdl.Surface
	var newTextHeight int32 = int32(self.fontHeight * strings.Count(self.printBuffer, "\n"))
	windowSurface, err = self.window.GetSurface()
	err = windowSurface.Blit(
		&sdl.Rect{
			X: 0, Y: 0,
			W: windowSurface.W, H: windowSurface.H},
		self.printSurface,
		&sdl.Rect{
			X: 0, Y:0,
			W: windowSurface.W, H: windowSurface.H})
	if ( err != nil ) {
		return err
	}
	err = windowSurface.FillRect(nil, 0x00000000)
	if ( err != nil ) {
		return err
	}
	err = self.printSurface.Blit(
		&sdl.Rect{
			X: 0, Y: newTextHeight,
			W: windowSurface.W, H: windowSurface.H - newTextHeight},
		windowSurface,
		&sdl.Rect{
			X: 0, Y: 0,
			W: windowSurface.W, H: newTextHeight})
	self.cursorX = 0
	self.cursorY = (self.maxCharsH - int32(strings.Count(self.printBuffer, "\n")))
	
	return nil
}

func (self *BasicRuntime) Write(text string) {
	fmt.Printf(text)
	self.drawWrappedText(
		(self.cursorX * int32(self.fontWidth)),
		(self.cursorY * int32(self.fontHeight)),
		text)
}

func (self *BasicRuntime) Println(text string) {
	fmt.Println(text)
	self.printBuffer += text + "\n"
	self.cursorY += int32(strings.Count(text, "\n"))
	self.cursorX = 0
}

func (self *BasicRuntime) drawPrintBuffer() error {
	var err error
	if ( len(self.printBuffer) == 0 ) {
		return nil
	}
	if ( self.cursorY >= self.maxCharsH - 1) {
		err = self.scrollWindow(0, int32(self.fontHeight * strings.Count(self.printBuffer, "\n"))+1)
		if ( err != nil ) {
			fmt.Println(err)
			return err
		}
		//fmt.Printf("Cursor X %d Y %d\n", self.cursorX, self.cursorY)
	}
	for _, line := range strings.Split(self.printBuffer, "\n") {
		if ( len(line) == 0 ) {
			break
		}
		err = self.drawWrappedText(
			(self.cursorX * int32(self.fontWidth)),
			(self.cursorY * int32(self.fontHeight)),
			line)
		if ( err != nil ) {
			fmt.Println(err)
			return err
		}
		self.cursorX = 0
		self.cursorY += 1
	}
	//fmt.Printf("Cursor X %d Y %d\n", self.cursorX, self.cursorY)
	if ( self.cursorY >= self.maxCharsH - 1) {
		//fmt.Println("Forcing cursor to bottom -1")
		self.cursorY = self.maxCharsH - 1
	}
	return nil
}
