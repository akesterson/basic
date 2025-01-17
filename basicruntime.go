package main

import (
	"fmt"
	"errors"
	"strings"
	"io"
	"bufio"
	"os"
	"slices"
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
	run_finished_mode int
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

func (self *BasicRuntime) evaluateSome(expr *BasicASTLeaf, leaftypes ...BasicASTLeafType) (*BasicValue, error) {
	if ( slices.Contains(leaftypes, expr.leaftype)) {
		return self.evaluate(expr)
	}
	return nil, nil
}

func (self *BasicRuntime) evaluate(expr *BasicASTLeaf, leaftypes ...BasicASTLeafType) (*BasicValue, error) {
	var lval *BasicValue
	var rval *BasicValue
	var err error = nil

	lval, err = self.newValue()
	if ( err != nil ) {
		return nil, err
	}
	lval.init()

	//fmt.Printf("Evaluating leaf type %d\n", expr.leaftype)
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
	case LEAF_COMMAND_IMMEDIATE: fallthrough
	case LEAF_COMMAND:
		//fmt.Printf("Processing command %s\n", expr.identifier)
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
			//fmt.Println("Processing RUN")
			if ( rval == nil ) {
				self.nextline = 0
			} else {
				self.nextline = int(rval.intval)
			}
			self.mode = MODE_RUN
			//fmt.Printf("Set mode %d with nextline %d\n", self.mode, self.nextline)
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

func (self *BasicRuntime) interpret(expr *BasicASTLeaf) (*BasicValue, error) {
	var value *BasicValue
	var err error
	value, err = self.evaluate(expr)
	if ( err != nil ) {
		fmt.Println(err)
		return nil, err
	}
	return value, nil
}

func (self *BasicRuntime) interpretImmediate(expr *BasicASTLeaf) (*BasicValue, error) {
	var value *BasicValue
	var err error
	value, err = self.evaluateSome(expr, LEAF_COMMAND_IMMEDIATE)
	//fmt.Printf("after evaluateSome in mode %d\n", self.mode)
	if ( err != nil ) {
		fmt.Println(err)
		return nil, err
	}
	return value, nil
}


func (self *BasicRuntime) processLineRunStream(readbuff *bufio.Scanner) {
	var line string
	if ( readbuff.Scan() ) {
		line = readbuff.Text()
		// All we're doing is getting the line #
		// and storing the source line in this mode.
		self.scanner.scanTokens(line)
	} else {
		self.mode = MODE_RUN
	}
}

func (self *BasicRuntime) processLineRepl(readbuff *bufio.Scanner) {
	var leaf *BasicASTLeaf = nil
	var err error = nil
	fmt.Println("READY")
	if ( readbuff.Scan() ) {
		self.scanner.scanTokens(readbuff.Text())
		leaf, err = self.parser.parse()
		if ( err != nil ) {
			self.basicError(RUNTIME, err.Error())
			return
		}
		_, _ = self.interpretImmediate(leaf)
		//fmt.Printf("Leaving repl function in mode %d", self.mode)
	}
}

func (self *BasicRuntime) processLineRun(readbuff *bufio.Scanner) {
	var line string
	var leaf *BasicASTLeaf = nil
	var err error = nil
	//fmt.Printf("RUN line %d\n", self.nextline)
	if ( self.nextline >= MAX_SOURCE_LINES ) {
		self.mode = self.run_finished_mode
		return
	}
	line = self.source[self.nextline]
	self.lineno = self.nextline
	self.nextline += 1
	if ( line == "" ) {
		return
	}
	//fmt.Println(line)
	self.scanner.scanTokens(line)
	leaf, err = self.parser.parse()
	if ( err != nil ) {
		self.basicError(RUNTIME, err.Error())
		self.mode = MODE_QUIT
		return
	}
	_, _ = self.interpret(leaf)	
}

func (self *BasicRuntime) run(fileobj io.Reader, mode int) {
	var readbuff = bufio.NewScanner(fileobj)

	self.parser.init(self)
	self.scanner.init(self, &self.parser)
	self.mode = mode
	if ( self.mode == MODE_REPL ) {
		self.run_finished_mode = MODE_REPL
	} else {
		self.run_finished_mode = MODE_QUIT
	}
	for {
		//fmt.Printf("Starting in mode %d\n", self.mode)
		self.zero()
		self.parser.zero()
		self.scanner.zero()
		switch (self.mode) {
		case MODE_QUIT:
			os.Exit(0)
		case MODE_RUNSTREAM:
			self.processLineRunStream(readbuff)
		case MODE_REPL:
			self.processLineRepl(readbuff)
		case MODE_RUN:
			self.processLineRun(readbuff)
		}
		//fmt.Printf("Finishing in mode %d\n", self.mode)

	}
}
