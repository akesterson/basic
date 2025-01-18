package main

import (
	"fmt"
	"errors"
)

func (self *BasicRuntime) CommandPRINT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	if ( rval == nil ) {
		return nil, errors.New("Expected expression")
	}
	fmt.Println(rval.toString())
	return nil, nil	
}

func (self *BasicRuntime) CommandGOTO(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	if ( rval == nil ) {
		return nil, errors.New("Expected expression")
	}
	if ( rval.valuetype != TYPE_INTEGER ) {
		return nil, errors.New("Expected integer")
	}
	self.nextline = int(rval.intval)
	return nil, nil
}

func (self *BasicRuntime) CommandRUN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	//fmt.Println("Processing RUN")
	if ( rval == nil ) {
		self.nextline = 0
	} else {
		if ( rval.valuetype != TYPE_INTEGER ) {
			return nil, errors.New("Expected integer")
		}
		self.nextline = int(rval.intval)
	}
	self.mode = MODE_RUN
	//fmt.Printf("Set mode %d with nextline %d\n", self.mode, self.nextline)
	return nil, nil
}

func (self *BasicRuntime) CommandQUIT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	self.mode = MODE_QUIT
	return nil, nil
}

func (self *BasicRuntime) CommandLET(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	// LET is not expressly required in our basic implementation or in
	// Commodore 128 BASIC 7.0. Assignments to variables are handled as
	// part of expression evaluation, LET doesn't need to manage it.
	return nil, nil
}
