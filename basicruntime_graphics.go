package main

import (
	"fmt"
	"strings"
	"github.com/veandco/go-sdl2/sdl"
)

func (self *BasicRuntime) drawText(x int32, y int32, text string) error {
	var windowSurface *sdl.Surface
	var textSurface *sdl.Surface
	var err error

	windowSurface, err = self.window.GetSurface()
	if ( err != nil ) {
		return err
	}

	textSurface, err = self.font.RenderUTF8Blended(
		text,
		sdl.Color{R: 255, G: 255, B: 255, A: 255})
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
	self.window.UpdateSurface()
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

func (self *BasicRuntime) advanceCursor(text string) {
	var y int = strings.Count(text, "\n")
	var x = (len(text) - strings.LastIndex(text, "\n"))
	if ( int32(x) > self.maxCharsW ) {
		y += 1
	}
	if ( y > 0 ) {
		self.cursorX = int32(x)
	} else {
		self.cursorX += int32(x)
	}
	self.cursorY += int32(y)
	fmt.Println("New cursor X %d Y %d", self.cursorX, self.cursorY)
}

func (self *BasicRuntime) Println(text string) {
	fmt.Println(text)
	self.printBuffer += text + "\n"
}

func (self *BasicRuntime) drawPrintBuffer() error {
	var err error
	if ( len(self.printBuffer) == 0 ) {
		return nil
	}
	if ( self.cursorY >= self.maxCharsH ) {
		err = self.scrollWindow(0, int32(self.fontHeight * strings.Count(self.printBuffer, "\n"))+1)
		if ( err != nil ) {
			return err
		}
	}
	for _, line := range strings.Split(self.printBuffer, "\n") {
		if ( len(line) == 0 ) {
			break
		}
		err = self.drawText(
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
	fmt.Println("Cursor X %d Y %d", self.cursorX, self.cursorY)
	if ( self.cursorY >= self.maxCharsH - 2) {
		fmt.Println("Forcing cursor to bottom -2")
		self.cursorY = self.maxCharsH - 2
	}
	return nil
}
