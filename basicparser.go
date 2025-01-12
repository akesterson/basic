package main

import (
	"errors"
)

type BasicParser struct {
	context *BasicContext
	token [16]BasicToken
	nexttoken int
}

func (self *BasicParser) init(context *BasicContext) error {
	if ( context == nil ) {
		return errors.New("nil context argument")
	}
	self.nexttoken = 0
	self.context = context
	return nil
}

