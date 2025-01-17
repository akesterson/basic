package main

import (
	"fmt"
	"errors"
	"strings"
	"io"
	"bufio"
	"os"
)

type BasicError int
const (
	IO           BasicError = iota
	PARSE
	SYNTAX
	RUNTIME
)

type BasicRuntime struct {
	source [MAX_SOURCE_LINES]string
	lineno int
	values [MAX_VALUES]BasicValue
	nextvalue int
	nextline int
	mode int
	scanner BasicScanner
	parser BasicParser
}

func (self *BasicRuntime) zero() {
	for i, _ := range self.values {
		self.values[i].init()
	}
	self.nextvalue = 0
}


func (self *BasicRuntime) init() {
	self.lineno = 0
	self.nextline = 0
	self.zero()
}

func (self *BasicRuntime) errorCodeToString(errno BasicError) string {
	switch (errno) {
	case IO: return "IO ERROR"
	case PARSE: return "PARSE ERROR"
	case RUNTIME: return "RUNTIME ERROR"
	case SYNTAX: return "SYNTAX ERROR"
	}
	return "UNDEF"
}

func (self *BasicRuntime) basicError(errno BasicError, message string) {
	fmt.Printf("? %d : %s %s", self.lineno, self.errorCodeToString(errno), message)
}

func (self BasicRuntime) newValue() (*BasicValue, error) {
	var value *BasicValue
	if ( self.nextvalue < MAX_VALUES ) {
		value = &self.values[self.nextvalue]
		self.nextvalue += 1
		return value, nil
	}
	return nil, errors.New("Maximum values per line reached")
}

func (self BasicRuntime) isTrue(value *BasicValue) (bool, error) {
	if ( value.valuetype == TYPE_STRING ) {
		return false, errors.New("strings cannot evaluate to true (-1) or false (0)")
	}
	if ( value.intval == BASIC_TRUE || value.floatval == BASIC_TRUE ) {
		return true, nil
	}
	return false, nil
}

func (self BasicRuntime) evaluate(expr *BasicASTLeaf) (*BasicValue, error) {
	var lval *BasicValue
	var rval *BasicValue
	var err error = nil
	
	lval, err = self.newValue()
	if ( err != nil ) {
		return nil, err
	}
	lval.init()
	
	switch (expr.leaftype) {
	case LEAF_GROUPING: return self.evaluate(expr.expr)
	case LEAF_LITERAL_INT:
		lval.valuetype = TYPE_INTEGER
		lval.intval = expr.literal_int
	case LEAF_LITERAL_FLOAT:
		lval.valuetype = TYPE_FLOAT
		lval.floatval = expr.literal_float
	case LEAF_LITERAL_STRING:
		lval.valuetype = TYPE_STRING
		lval.stringval = expr.literal_string
	case LEAF_UNARY:
		rval, err = self.evaluate(expr.right)
		if ( err != nil ) {
			return nil, err
		}
		switch (expr.operator) {
		case MINUS:
			err = rval.invert()
			if ( err != nil ) {
				return nil, err
			}
			return rval, nil
		case NOT:
			err = rval.bitwiseNot()
			if ( err != nil ) {
				return nil, err
			}
			return rval, nil
		default:
			return nil, errors.New(fmt.Sprintf("Don't know how to perform operation %d on unary type %d", expr.operator, rval.valuetype))
		}
	case LEAF_COMMAND:
		if ( expr.right != nil ) {
			rval, err = self.evaluate(expr.right)
			if ( err != nil ) {
				return nil, err
			}
		}
		if ( strings.Compare(expr.identifier, "PRINT") == 0 ) {
			if ( rval == nil ) {
				return nil, errors.New("Expected expression")
			}
			fmt.Println(rval.toString())
			return nil, nil
		} else if ( strings.Compare(expr.identifier, "RUN" ) == 0 ) {
			if ( rval == nil ) {
				self.nextline = 0
			} else {
				self.nextline = int(rval.intval)
			}
			self.mode = MODE_RUN
			return nil, nil
		} else if ( strings.Compare(expr.identifier, "QUIT" ) == 0 ) {
			self.mode = MODE_QUIT
		}
	case LEAF_BINARY:
		lval, err = self.evaluate(expr.left)
		if ( err != nil ) {
			return nil, err
		}
		rval, err = self.evaluate(expr.right)
		if ( err != nil ) {
			return nil, err
		}
		switch (expr.operator) {
		case ASSIGNMENT:
			return nil, errors.New("Assignment not implemented yet")
		case MINUS:
			err = lval.mathMinus(rval)
		case PLUS:
			err = lval.mathPlus(rval)
		case LEFT_SLASH:
			err = lval.mathDivide(rval)
		case STAR:
			err = lval.mathMultiply(rval)
		case AND:
			err = lval.bitwiseAnd(rval)
		case OR:
			err = lval.bitwiseOr(rval)
		case LESS_THAN:
			err = lval.lessThan(rval)
		case LESS_THAN_EQUAL:
			err = lval.lessThanEqual(rval)
		case EQUAL:
			err = lval.isEqual(rval)
		case NOT_EQUAL:
			err = lval.isNotEqual(rval)
		case GREATER_THAN:
			err = lval.greaterThan(rval)
		case GREATER_THAN_EQUAL:
			err = lval.greaterThanEqual(rval)
		}
		if ( err != nil ) {
			return nil, err
		}
	}
	return lval, nil
}

func (self *BasicRuntime) interpret(expr *BasicASTLeaf) *BasicValue{
	var value *BasicValue
	var err error
	value, err = self.evaluate(expr)
	if ( err != nil ) {
		fmt.Println(err)
		self.mode = MODE_REPL
		return nil
	}
	return value
}

func (self *BasicRuntime) run(fileobj io.Reader, mode int) {
	var readbuff = bufio.NewScanner(fileobj)
	var leaf *BasicASTLeaf = nil
	var err error = nil
	var enable_repl = true
	var line string

	self.parser.init(self)
	self.scanner.init(self, &self.parser)
	self.mode = mode
	for {
		self.zero()
		self.parser.zero()
		self.scanner.zero()
		switch (self.mode) {
		case MODE_QUIT:
			os.Exit(0)
		case MODE_RUNSTREAM:
			enable_repl = false
			if ( readbuff.Scan() ) {
				line = readbuff.Text()
				// All we're doing is getting the line #
				// and storing the source line.
				self.scanner.scanTokens(line)
			} else {
				self.mode = MODE_RUN
			}
		case MODE_REPL:
			if ( enable_repl == false ) {
				self.mode = MODE_QUIT
				break
			}
			fmt.Println("READY")
			if ( readbuff.Scan() ) {
				self.scanner.scanTokens(readbuff.Text())
				leaf, err = self.parser.parse()
				if ( err != nil ) {
					self.basicError(RUNTIME, err.Error())
				}
			}
		case MODE_RUN:
			if ( self.nextline >= MAX_SOURCE_LINES ) {
				self.mode = MODE_QUIT
				continue
			}
			line = self.source[self.nextline]
			self.lineno = self.nextline
			self.nextline += 1
			if ( line == "" ) {
				continue
			}
			fmt.Println(line)
			self.scanner.scanTokens(line)
			leaf, err = self.parser.parse()
			if ( err != nil ) {
				self.basicError(RUNTIME, err.Error())
				self.mode = MODE_QUIT
			} else {
				_ = self.interpret(leaf)
			}
			if ( self.mode != MODE_RUN ) {
				break
			}
		}
	}
}
