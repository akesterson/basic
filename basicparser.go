package main

import (
	"fmt"
	"errors"
	"slices"
	"reflect"
)

type BasicToken struct {
	tokentype BasicTokenType
	lineno int64
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
   line           -> (line_number ( command | expression )) (immediate_command expression)
   command        -> command (expression)
   expression     -> logicalandor
   logicalandor   -> logicalnot ( "OR" "AND" ) logicalnot
   logicalnot     -> "NOT" relation
   relation       -> subtraction* [ < <= = <> >= > ] subtraction*
   subtraction    -> addition* "-" addition*
   addition       -> multiplication* "+" multiplication*
   multiplication -> division* "*" division*
   division       -> unary* "/" unary*
   unary          -> "-" exponent
   primary        -> IDENTIFIER | LITERAL_INT | LITERAL_FLOAT | LITERAL_STRING | "(" expression ")"
   
*/

func (self *BasicParser) init(runtime *BasicRuntime) error {
	if ( runtime == nil ) {
		return errors.New("nil runtime argument")
	}
	self.zero()
	self.runtime = runtime
	return nil
}

func (self *BasicParser) dump() {
	for idx, value := range(self.tokens) {
		fmt.Printf("token[%d] = %+v\n", idx, value)
	}
}

func (self *BasicParser) zero() {
	if ( self == nil ) {
		panic("nil self reference!")
	}
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
	if ( self.nextleaf < MAX_LEAVES ) {
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
	leaf, err = self.statement()
	if ( leaf != nil ) {
		//fmt.Printf("%+v\n", leaf)
	}
	return leaf, err
	// later on when we add statements we may need to handle the error
	// internally; for now just pass it straight out.
}

func (self *BasicParser) statement() (*BasicASTLeaf, error) {
	return self.command()
	return nil, self.error(fmt.Sprintf("Expected command or expression"))
}

func (self *BasicParser) commandByReflection(root string, command string) (*BasicASTLeaf, error) {
	var methodiface interface{}
	var reflector reflect.Value
	var rmethod reflect.Value

	// TODO : There is some possibility (I think, maybe) that the way I'm
	// getting the method through reflection might break the receiver
	// assignment on the previously bound methods. If `self.` starts
	// behaving strangely on command methods, revisit this.
	
	reflector = reflect.ValueOf(self)
	if ( reflector.IsNil() || reflector.Kind() != reflect.Ptr ) {
		return nil, errors.New("Unable to reflect runtime structure to find command method")
	}
	rmethod = reflector.MethodByName(fmt.Sprintf("%s%s", root, command))
	if ( !rmethod.IsValid() ) {
		// It's not an error to have no parser function, this just means our rval
		// gets parsed as an expression
		return nil, nil
	}
	if ( !rmethod.CanInterface() ) {
		return nil, fmt.Errorf("Unable to execute command %s", command)
	}
	methodiface = rmethod.Interface()
	
	methodfunc, ok := methodiface.(func() (*BasicASTLeaf, error))
	if ( !ok ) {
		return nil, fmt.Errorf("ParseCommand%s has an invalid function signature", command)
	}
	return methodfunc()
}

func (self *BasicParser) command() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var righttoken *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	if self.match(COMMAND, COMMAND_IMMEDIATE) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}

		// Is it a command that requires special parsing?
		expr, err = self.commandByReflection("ParseCommand", operator.lexeme)
		if ( err != nil ) {
			return nil, err
		}
		if ( expr != nil ) {
			return expr, nil
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
			//fmt.Printf("Command : %s->%s\n", expr.toString(), expr.right.toString())
		}
		return expr, nil
	}
	return self.assignment()
}

func (self *BasicParser) assignment() (*BasicASTLeaf, error) {
	var identifier *BasicASTLeaf = nil
	var expr *BasicASTLeaf = nil
	var right *BasicASTLeaf = nil
	var err error = nil
	var identifier_leaf_types = []BasicASTLeafType{
		LEAF_IDENTIFIER_INT,
		LEAF_IDENTIFIER_FLOAT,
		LEAF_IDENTIFIER_STRING,
	}

	identifier, err = self.expression()
	if ( err != nil ) {
		return nil, err
	} else if ( ! slices.Contains(identifier_leaf_types, identifier.leaftype) ) {
		return identifier, err
	}
	if self.match(ASSIGNMENT) {
		right, err = self.expression()
		if ( err != nil ) {
			return nil, err
		}
		//fmt.Printf("%+v\n", right)
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(identifier, ASSIGNMENT, right)
		return expr, nil
	}
	return identifier, err
}

func (self *BasicParser) argumentList() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var first *BasicASTLeaf = nil
	var err error = nil

	// argument lists are just (.right) joined expressions continuing
	// ad-infinitum.

	if ( !self.match(LEFT_PAREN) ) {
		//return nil, errors.New("Expected argument list (expression, ...)")
		//fmt.Printf("No left paren\n")
		return nil, nil
	}
	expr, err = self.expression()
	if ( err != nil ) {
		return nil, err
	}
	first = expr
	//fmt.Printf("Before loop: %+v\n", expr)
	for ( expr != nil && self.match(COMMA) ) {
		expr.right, err = self.expression()
		if ( err != nil ) {
			return nil, err
		}
		expr = expr.right
		//fmt.Printf("Argument : %+v\n", expr)
	}
	//fmt.Println("Done with loop")
	if ( !self.match(RIGHT_PAREN) ) {
		return nil, errors.New("Unbalanced parenthesis")
	}
	return first, nil
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
	if self.match(LESS_THAN, LESS_THAN_EQUAL, EQUAL, NOT_EQUAL, GREATER_THAN, GREATER_THAN_EQUAL) {
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
	var left *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	left, err = self.addition()
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
		if ( expr != nil ) {
			left = expr
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(left, operator.tokentype, right)
		return expr, nil
	}
	return left, nil
}

func (self *BasicParser) addition() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var left *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	left, err = self.multiplication()
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
		if ( expr != nil ) {
			left = expr
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(left, operator.tokentype, right)
	}
	if ( expr != nil ) {
		return expr, nil
	}
	return left, nil
}

func (self *BasicParser) multiplication() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var left *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	left, err = self.division()
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
		if ( expr != nil ) {
			left = expr
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(left, operator.tokentype, right)
	}
	if ( expr != nil ) {
		return expr, nil
	}
	return left, nil
}

func (self *BasicParser) division() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var left *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	left, err = self.unary()
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
		if ( expr != nil ) {
			left = expr
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(left, operator.tokentype, right)
	}
	if ( expr != nil ) {
		return expr, nil
	}
	return left, nil
}

func (self *BasicParser) unary() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	if self.match(MINUS) {
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
	var left *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var right *BasicASTLeaf = nil
	var err error = nil

	left, err = self.function()
	if ( err != nil ) {
		return nil, err
	}
	for self.match(CARAT) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		right, err = self.function()
		if ( err != nil ) {
			return nil, err
		}
		if ( expr != nil ) {
			left = expr
		}
		expr, err = self.newLeaf()
		if ( err != nil ) {
			return nil, err
		}
		expr.newBinary(left, operator.tokentype, right)
		return expr, nil
	}
	if ( expr != nil ) {
		return expr, nil
	}
	return left, nil
}


func (self *BasicParser) function() (*BasicASTLeaf, error) {
	var arglist *BasicASTLeaf = nil
	var leafptr *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var refarglen int = 0
	var defarglen int = 0
	var fndef *BasicFunctionDef = nil
	var err error = nil

	// This is ONLY called for function CALLS, not for function DEFs.
	if ( self.match(FUNCTION) ) {
		operator, err = self.previous()
		if ( err != nil ) {
			return nil, err
		}
		//fmt.Printf("Checking for existence of user function %s...\n", operator.lexeme)

		fndef = self.runtime.environment.getFunction(operator.lexeme)
		if ( fndef == nil ) {
			return nil, fmt.Errorf("No such function %s", operator.lexeme)
		}
		if ( fndef != nil ) {
			// All we can do here is collect the argument list and
			// check the length
			arglist, err = self.argumentList()
			if ( err != nil ) {
				return nil, err
			}
			leafptr = arglist
			for ( leafptr != nil ) {
				defarglen += 1
				leafptr = leafptr.right
			}
			leafptr = fndef.arglist
			for ( leafptr != nil ) {
				refarglen += 1
				leafptr = leafptr.right
			}
			if ( defarglen != refarglen ) {
				return nil, fmt.Errorf("function %s takes %d arguments, received %d", fndef.name, refarglen, defarglen)
			}
			leafptr, err = self.newLeaf()
			if ( err != nil ) {
				return nil, err
			}
			leafptr.newFunction(operator.lexeme, arglist)
			//fmt.Printf("%s\n", leafptr.toString())
			return leafptr, nil
		}
	}
	return self.primary()
}

func (self *BasicParser) primary() (*BasicASTLeaf, error) {
	var expr *BasicASTLeaf = nil
	var previous *BasicToken = nil
	var groupexpr *BasicASTLeaf = nil
	var err error = nil

	if self.match(LITERAL_INT, LITERAL_FLOAT, LITERAL_STRING, IDENTIFIER, IDENTIFIER_STRING, IDENTIFIER_FLOAT, IDENTIFIER_INT, FUNCTION) {
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
			expr.right, err = self.argumentList()
			if ( err != nil ) {
				return nil, err
			}
		case IDENTIFIER_FLOAT:
			expr.newIdentifier(LEAF_IDENTIFIER_FLOAT, previous.lexeme)
			expr.right, err = self.argumentList()
			if ( err != nil ) {
				return nil, err
			}
		case IDENTIFIER_STRING:
			expr.newIdentifier(LEAF_IDENTIFIER_STRING, previous.lexeme)
			expr.right, err = self.argumentList()
			if ( err != nil ) {
				return nil, err
			}
		case FUNCTION: fallthrough
		case IDENTIFIER:
			expr.newIdentifier(LEAF_IDENTIFIER, previous.lexeme)
		default:
			return nil, errors.New("Invalid literal type, command or function name")
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
	//fmt.Printf("At curtoken %d\n", self.curtoken)
	return nil, self.error("Expected expression or literal")
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
	if (self.curtoken >= (MAX_TOKENS - 1) || self.curtoken >= self.nexttoken ) {
		return true
	}
	return false
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


