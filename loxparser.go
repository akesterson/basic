package main

import (
	"errors"
)

type LoxParser struct {
	context *BasicContext
	token [16]BasicToken
	nexttoken int
	curtoken int
	leaves [16]BasicASTLeaf
	nextleaf int
}

func (self *LoxParser) init(context *BasicContext) error {
	if ( context == nil ) {
		return errors.New("nil context argument")
	}
	self.nexttoken = 0
	self.context = context
	self.nextleaf = 0
	return nil
}

func (self *LoxParser) nextLeaf() *BasicASTLeaf, error {
	var i int
	if self.nextleaf < 16 {
		self.nextleaf += 1
		return &self.leaves[nextLeaf], nil
	}
	return nil, errors.New("No available leaves in the parser")
}

func (self *LoxParser) parse() error {
	return nil
}

func (self *LoxParser) expression() *BasicASTLeaf, error {
	return self.equality()
}


func (self *LoxParser) equality() *BasicASTLeaf, error {
	var expr *BasicASTLeaf = nil
	var comparison *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTToken = nil
	var err error = nil

	comparison, err = self.comparison()
	if ( err != nil ) {
		return nil, err
	}
	for match(EQUAL, NOT_EQUAL) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.comparison()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(comparison, operator, right)
	}
	return expr, nil
}

func (self *LoxParser) equality() *BasicASTLeaf, error {
	var expr *BasicASTLeaf = nil
	var term *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTToken = nil
	var err error = nil

	term, err = self.term()
	if ( err != nil ) {
		return nil, err
	}
	while (match(LESS_THAN, LESS_THAN_EQUAL, GREATER_THAN, GREATER_THAN_EQUAL)) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.term()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(term, operator, right)
	}
	return expr, nil
}
	
func (self *LoxParser) term() *BasicASTLeafe, error {
	var expr *BasicASTLeaf = nil
	var factor *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTToken = nil
	var err error = nil

	factor, err = self.factor()
	if ( err != nil ) {
		return nil, err
	}
	while (match(PLUS, MINUS)) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.factor()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(factor, operator, right)
	}
	return expr, nil
}

func (self *LoxParser) factor() *BasicASTLeafe, error {
	var expr *BasicASTLeaf = nil
	var unary *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTToken = nil
	var err error = nil

	unary, err = self.unary()
	if ( err != nil ) {
		return nil, err
	}
	while (match(SLASH, STAR)) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.unary()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(unary, operator, right)
	}
	return expr, nil
}

func (self *LoxParser) unary() *BasicASTLeafe, error {
	var expr *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTToken = nil
	var err error = nil

	if (match(NOT, MINUS)) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.unary()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newUnary(operator, right)
		return expr, nil
	}
	return self.primary()
}

func (self *LoxParser) primary() *BasicASTLeafe, error {
	var expr *BasicASTLeaf = nil
	var previous *BasicToken = nil
	var groupexpr *BasicASTToken = nil
	var err error = nil


	if match(LITERAL_NUMBER, LITERAL_STRING) {
		previous, err = self.previous()
		if ( err != nil ) {
			return err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return err
		}
		switch (previous.tokentype) {
		case LITERAL_INT:
			expr.newLiteralInt(previous.literal_int)
		case LITERAL_FLOAT:
			expr.newLiteralFloat(previous.literal_float)
		case LITERAL_STRING:
			expr.newLiteralString(previous.literal_string)
		default:
			return errors.new("Invalid literal type")
		}
		return expr, nil
	}
	if match(LEFT_PAREN) {
		groupexpr, err = self.expression()
		if ( err != nil ) {
			return err
		}
		self.consume(RIGHT_PAREN, "Missing ) after expression")
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return err
		}
		expr.newGrouping(groupexpr)
		return expr, nil
	}
}

func (self *LoxParser) match(types ...BasicTokenType) bool {
	for _, tokentype := range types {
		if ( self.check(tokentype) ) {
			self.advance()
			return true
		}
	}
	return false
}

func (self *LoxParser) check(tokentype BasicTokenType) bool {
	if ( self.isAtEnd() ) {
		return false
	}
	next_token = self.peek()
	return (next_token.tokentype == tokentype)
}

func (self *LoxParser) advance() *BasicToken, error {
	if ( !self.isAtEnd() ) {
		self.curtoken += 1
	}
	return self.previous()
}

func (self *LoxParser) isAtEnd() bool {
	return (self.curtoken >= 15)
}

func (self *LoxParser) peek() *BasicToken {
	if ( self.isAtEnd() ) {
		return  nil
	}
	return &self.tokens[self.curtoken]
}

func (self *LoxParser) previous() *BasicToken {
	if ( self.curtoken > 0 ) {
		return  nil
	}
	return &self.tokens[self.curtoken - 1]
}	


