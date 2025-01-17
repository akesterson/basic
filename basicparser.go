package main

import (
	"fmt"
	"errors"
	"slices"
	"strconv"
)

type BasicToken struct {
	tokentype BasicTokenType
	lineno int
literal string
	lexeme string
}

func (self *BasicToken) init() {
	self.tokentype = UNDEFINED
	self.lineno = 0
	self.literal = ""
	self.lexeme = ""
}

func (self BasicToken) toString() string {
	return fmt.Sprintf("%d %s %s", self.tokentype, self.lexeme, self.literal)
}

type BasicParser struct {
	runtime *BasicRuntime
	tokens [MAX_TOKENS]BasicToken
	errorToken *BasicToken
	nexttoken int
	curtoken int
	leaves [MAX_TOKENS]BasicASTLeaf
	nextleaf int
	immediate_commands []string
}

/*
   This hierarcy is as-per "Commodore 128 Programmer's Reference Guide" page 23

   program        -> line*
   line           -> expression? ( statement expression )
   statement      -> identifier expression*
   expression     -> logicalandor
   logicalandor   -> logicalnot ( "OR" "AND" ) logicalnot
   logicalnot     -> "NOT" relation
   relation       -> subtraction [ < <= = <> >= > ] subtraction
   subtraction    -> addition "-" addition
   addition       -> multiplication "+" multiplication
   multiplication -> division "*" division
   division       -> unary "/" unary
   unary          -> exponent "-" exponent
   primary        -> LITERAL_INT | LITERAL_FLOAT | LITERAL_STRING | "(" expression ")"
   
*/

func (self *BasicParser) init(runtime *BasicRuntime) error {
	if ( runtime == nil ) {
		return errors.New("nil runtime argument")
	}
	self.zero()
	self.runtime = runtime
	return nil
}

func (self *BasicParser) zero() {
	for i, _ := range self.leaves {
		self.leaves[i].init(LEAF_UNDEFINED)
	}
	for i, _ := range self.tokens {
		self.tokens[i].init()
	}
	self.curtoken = 0
	self.nexttoken = 0
	self.nextleaf = 0
}

func (self *BasicParser) newLeaf() (*BasicASTLeaf, error) {
	var leaf *BasicASTLeaf
	if ( self.nextleaf < 15 ) {
		leaf = &self.leaves[self.nextleaf]
		self.nextleaf += 1
		return leaf, nil
	} else {
		return nil, errors.New("No more leaves available")
	}
}

func (self *BasicParser) parse() (*BasicASTLeaf, error) {
	var leaf *BasicASTLeaf = nil
	var err error = nil
	leaf, err = self.line()
	if ( leaf != nil ) {
		//fmt.Printf("%+v\n", leaf)
	}
	return leaf, err
	// later on when we add statements we may need to handle the error
	// internally; for now just pass it straight out.
}

func (self *BasicParser) line() (*BasicASTLeaf, error) {
	var token *BasicToken = nil
	var err error = nil

	for self.match(LINE_NUMBER) {
		token, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		self.runtime.lineno, err = strconv.Atoi(token.lexeme)
		if ( err != nil ) {
			return nil, err
		}
		return self.command()
		
	}
	for self.check(COMMAND_IMMEDIATE) {
		//fmt.Println("Found immediate mode command token")
		// Some commands can run immediately without a line number...
		return self.command()
	}
	return nil, self.error(fmt.Sprintf("Expected line number or immediate mode command"))
}

func (self *BasicParser) command() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var righttoken *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	for self.match(COMMAND, COMMAND_IMMEDIATE) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		
		// some commands don't require an rval. Don't fail if there
		// isn't one. But fail if there is one and it fails to parse.
		righttoken = self.peek()
		if ( righttoken != nil && righttoken.tokentype != UNDEFINED ) {
			right, err = self.expression()
			if ( err != nil ) {
				return nil, err
			}
		}
		
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		if ( operator.tokentype == COMMAND_IMMEDIATE ) {
			expr.newImmediateCommand(operator.lexeme, right)
		} else {
			expr.newCommand(operator.lexeme, right)
		}
		//fmt.Printf("Returning %+v\n", expr)
		return expr, nil
	}
	return self.assignment()
}

func (self *BasicParser) assignment() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var right *BasicASTLeaf = nil
	var err error = nil
	var identifier_leaf_types = []BasicASTLeafType{
		LEAF_IDENTIFIER_INT,
		LEAF_IDENTIFIER_FLOAT,
		LEAF_IDENTIFIER_STRING,
	}

	expr, err = self.expression()
	if ( err != nil ) {
		return nil, err
	} else if ( ! slices.Contains(identifier_leaf_types, expr.leaftype) ) {
		return expr, err
	}
	for self.match(EQUAL) {
		right, err = self.expression()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(expr, ASSIGNMENT, right)
		return expr, nil
	}
	return expr, err
}

func (self *BasicParser) expression() (*BasicASTLeaf, error) {
	return self.logicalandor()
}

func (self *BasicParser) logicalandor() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var logicalnot *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

 	logicalnot, err = self.logicalnot()
	if ( err != nil ) {
		return nil, err
	}
	for self.match(AND, OR) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.logicalnot()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(logicalnot, operator.tokentype, right)
		return expr, nil
	}
	return logicalnot, nil
}

func (self *BasicParser) logicalnot() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	for self.match(NOT) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.relation()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newUnary(operator.tokentype, right)
		return expr, nil
	}
 	return self.relation()
}

func (self *BasicParser) relation() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var subtraction *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	subtraction, err = self.subtraction()
	if ( err != nil ) {
		return nil, err
	}
	for self.match(LESS_THAN, LESS_THAN_EQUAL, EQUAL, NOT_EQUAL, GREATER_THAN, GREATER_THAN_EQUAL) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.subtraction()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(subtraction, operator.tokentype, right)
		return expr, nil
	}
	return subtraction, nil
}

func (self *BasicParser) subtraction() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var addition *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	addition, err = self.addition()
	if ( err != nil ) {
		return nil, err
	}
	for self.match(MINUS) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.addition()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(addition, operator.tokentype, right)
		return expr, nil
	}
	return addition, nil
}

func (self *BasicParser) addition() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var multiplication *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	multiplication, err = self.multiplication()
	if ( err != nil ) {
		return nil, err
	}
	for self.match(PLUS) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.multiplication()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(multiplication, operator.tokentype, right)
		return expr, nil
	}
	return multiplication, nil
}

func (self *BasicParser) multiplication() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var division *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	division, err = self.division()
	if ( err != nil ) {
		return nil, err
	}
	for self.match(STAR) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.division()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(division, operator.tokentype, right)
		return expr, nil
	}
	return division, nil
}

func (self *BasicParser) division() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var unary *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	unary, err = self.unary()
	if ( err != nil ) {
		return nil, err
	}
	for self.match(LEFT_SLASH) {
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
		expr.newBinary(unary, operator.tokentype, right)
		return expr, nil
	}
	return unary, nil
}

func (self *BasicParser) unary() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	for self.match(MINUS) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.primary()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newUnary(operator.tokentype, right)
		return expr, nil
	}
	return self.exponent()
}

func (self *BasicParser) exponent() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var primary *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	primary, err = self.primary()
	if ( err != nil ) {
		return nil, err
	}
	for self.match(CARAT) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.primary()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(primary, operator.tokentype, right)
		return expr, nil
	}
	return primary, nil
}

func (self *BasicParser) primary() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var previous *BasicToken = nil
	var groupexpr *BasicASTLeaf = nil
	var err error = nil


	if self.match(LITERAL_INT, LITERAL_FLOAT, LITERAL_STRING, IDENTIFIER, IDENTIFIER_STRING, IDENTIFIER_FLOAT, IDENTIFIER_INT) {
		previous, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		switch (previous.tokentype) {
		case LITERAL_INT:
			expr.newLiteralInt(previous.lexeme)
		case LITERAL_FLOAT:
			expr.newLiteralFloat(previous.lexeme)
		case LITERAL_STRING:
			expr.newLiteralString(previous.lexeme)
		case IDENTIFIER_INT:
			expr.newIdentifier(LEAF_IDENTIFIER_INT, previous.lexeme)
		case IDENTIFIER_FLOAT:
			expr.newIdentifier(LEAF_IDENTIFIER_FLOAT, previous.lexeme)
		case IDENTIFIER_STRING:
			expr.newIdentifier(LEAF_IDENTIFIER_STRING, previous.lexeme)
		default:
			return nil, errors.New("Invalid literal type")
		}
		return expr, nil
	}
	if self.match(LEFT_PAREN) {
		groupexpr, err = self.expression()
		if ( err != nil ) {
			return nil, err
		}
		self.consume(RIGHT_PAREN, "Missing ) after expression")
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newGrouping(groupexpr)
		return expr, nil
	}
	return nil, self.error("Expected expression")
}

func (self *BasicParser) error(message string) error {
	self.errorToken = self.peek()
	if ( self.errorToken == nil ) {
		return errors.New("peek() returned nil token!")
	}
	if ( self.errorToken.tokentype == EOF ) {
		return errors.New(fmt.Sprintf("%d at end %s", self.errorToken.lineno, message))
	} else {
		return errors.New(fmt.Sprintf("%d at '%s', %s", self.errorToken.lineno, self.errorToken.lexeme, message))
	}
}

func (self *BasicParser) consume(tokentype BasicTokenType, message string) (*BasicToken, error) {
	if ( self.check(tokentype) ) {
		return self.advance()
	}

	return nil, self.error(message)
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
	var next_token *BasicToken
	if ( self.isAtEnd() ) {
		return false
	}
	next_token = self.peek()
	return (next_token.tokentype == tokentype)
}

func (self *BasicParser) advance() (*BasicToken, error) {
	if ( !self.isAtEnd() ) {
		self.curtoken += 1
	}
	return self.previous()
}

func (self *BasicParser) isAtEnd() bool {
	return (self.curtoken >= (MAX_TOKENS - 1))
}

func (self *BasicParser) peek() *BasicToken {
	if ( self.isAtEnd() ) {
		return nil
	}
	return &self.tokens[self.curtoken]
}

func (self *BasicParser) previous() (*BasicToken, error) {
	if ( self.curtoken == 0 ) {
		return  nil, errors.New("Current token is index 0, no previous token")
	}
	return &self.tokens[self.curtoken - 1], nil
}	


