package main

import (
	"errors"
)

type BasicParser struct {
	context *BasicContext
	token [16]BasicToken
	nexttoken int
	curtoken int
	leaves [16]BasicASTLeaf
	nextleaf int
}

func (self *BasicParser) init(context *BasicContext) error {
	if ( context == nil ) {
		return errors.New("nil context argument")
	}
	self.nexttoken = 0
	self.context = context
	self.nextleaf = 0
	return nil
}

func (self *BasicParser) parse() error {
	return nil
}

func (self *BasicParser) match(types ...BasicTokenType) bool {
	for _, tokentype := range types {
		if ( self.check(tokentype) ) {
			self.advance()
			return true
		}
	}
	return false
}

func (self *BasicParser) check(tokentype BasicTokenType) bool {
	if ( self.isAtEnd() ) {
		return false
	}
	next_token = self.peek()
	return (next_token.tokentype == tokentype)
}

func (self *BasicParser) advance() *BasicToken, error {
	if ( !self.isAtEnd() ) {
		self.curtoken += 1
	}
	return self.previous()
}

func (self *BasicParser) isAtEnd() bool {
	return (self.curtoken >= 15)
}

func (self *BasicParser) peek() *BasicToken {
	if ( self.isAtEnd() ) {
		return  nilx
	}
	return &self.tokens[self.curtoken]
}

func (self *BasicParser) previous() *BasicToken {
	if ( self.curtoken > 0 ) {
		return  nil
	}
	return &self.tokens[self.curtoken - 1]
}	


