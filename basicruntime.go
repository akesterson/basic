package main

import (
	"fmt"
	"errors"
	"io"
	"bufio"
	"os"
	"slices"
	"reflect"
	"unicode"
)

type BasicError int
const (
	NOERROR    BasicError = iota
	IO
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
	errno BasicError
	run_finished_mode int
	scanner BasicScanner
	parser BasicParser
	environment BasicEnvironment
}

func (self *BasicRuntime) zero() {
	for i, _ := range self.values {
		self.values[i].init()
	}
	self.errno = 0
	self.nextvalue = 0
}

func (self *BasicRuntime) init() {
	self.lineno = 0
	self.nextline = 0

	self.parser.init(self)
	self.scanner.init(self)
	self.environment.init(self)
	
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
	self.errno = errno
	fmt.Printf("? %d : %s %s\n", self.lineno, self.errorCodeToString(errno), message)
}

func (self *BasicRuntime) newValue() (*BasicValue, error) {
	var value *BasicValue
	if ( self.nextvalue < MAX_VALUES ) {
		value = &self.values[self.nextvalue]
		self.nextvalue += 1
		value.runtime = self
		return value, nil
	}
	return nil, errors.New("Maximum values per line reached")
}

func (self *BasicRuntime) isTrue(value *BasicValue) (bool, error) {
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
	case LEAF_BRANCH:
		rval, err = self.evaluate(expr.expr)
		if ( err != nil ) {
			self.basicError(RUNTIME, err.Error())
			return nil, err
			
		}
		if ( rval.boolvalue == BASIC_TRUE ) {
			return self.evaluate(expr.left)
		}
		if ( expr.right != nil ) {
			// For some branching operations, a false
			// branch is optional.
			return self.evaluate(expr.right)
		}
	case LEAF_IDENTIFIER_INT: fallthrough
	case LEAF_IDENTIFIER_FLOAT: fallthrough
	case LEAF_IDENTIFIER_STRING:
		lval = self.environment.get(expr.identifier)
		if ( lval == nil ) {
			return nil, fmt.Errorf("Identifier %s is undefined", expr.identifier)
		}
		return lval, nil
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
			return rval.invert()
		case NOT:
			return rval.bitwiseNot()
		default:
			return nil, errors.New(fmt.Sprintf("Don't know how to perform operation %d on unary type %d", expr.operator, rval.valuetype))
		}
	case LEAF_COMMAND_IMMEDIATE: fallthrough
	case LEAF_COMMAND:
		//fmt.Printf("Processing command %s\n", expr.identifier)
		return self.commandByReflection(expr, lval, rval)
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
			return self.environment.assign(expr.left, rval)
		case MINUS:
			return lval.mathMinus(rval)
		case PLUS:
			return lval.mathPlus(rval)
		case LEFT_SLASH:
			return lval.mathDivide(rval)
		case STAR:
			return lval.mathMultiply(rval)
		case AND:
			return lval.bitwiseAnd(rval)
		case OR:
			return lval.bitwiseOr(rval)
		case LESS_THAN:
			return lval.lessThan(rval)
		case LESS_THAN_EQUAL:
			return lval.lessThanEqual(rval)
		case EQUAL:
			return lval.isEqual(rval)
		case NOT_EQUAL:
			return lval.isNotEqual(rval)
		case GREATER_THAN:
			return lval.greaterThan(rval)
		case GREATER_THAN_EQUAL:
			return lval.greaterThanEqual(rval)
		}
		if ( err != nil ) {
			return nil, err
		}
	}
	return lval, nil
}

func (self *BasicRuntime) commandByReflection(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
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
	rmethod = reflector.MethodByName(fmt.Sprintf("Command%s", expr.identifier))
	if ( !rmethod.IsValid() ) {
		return nil, fmt.Errorf("Unknown command %s", expr.identifier)
	}
	if ( !rmethod.CanInterface() ) {
		return nil, fmt.Errorf("Unable to execute command %s", expr.identifier)
	}
	methodiface = rmethod.Interface()
	
	methodfunc, ok := methodiface.(func(*BasicASTLeaf, *BasicValue, *BasicValue) (*BasicValue, error))
	if ( !ok ) {
		return nil, fmt.Errorf("Command %s has an invalid function signature", expr.identifier)
	}
	return methodfunc(expr, lval, rval)
}

func (self *BasicRuntime) interpret(expr *BasicASTLeaf) (*BasicValue, error) {
	var value *BasicValue
	var err error
	value, err = self.evaluate(expr)
	if ( err != nil ) {
		self.basicError(RUNTIME, err.Error())
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
		self.setMode(MODE_RUN)
	}
}

func (self *BasicRuntime) processLineRepl(readbuff *bufio.Scanner) {
	var leaf *BasicASTLeaf = nil
	var err error = nil
	var line string
	if ( readbuff.Scan() ) {
		line = readbuff.Text()
		self.scanner.scanTokens(line)
		leaf, err = self.parser.parse()
		if ( err != nil ) {
			self.basicError(PARSE, err.Error())
			return
		}
		if ( !unicode.IsDigit(rune(line[0])) ) {
			_, _ = self.interpretImmediate(leaf)
		}
		//fmt.Printf("Leaving repl function in mode %d", self.mode)
	}
}

func (self *BasicRuntime) processLineRun(readbuff *bufio.Scanner) {
	var line string
	var leaf *BasicASTLeaf = nil
	var err error = nil
	//fmt.Printf("RUN line %d\n", self.nextline)
	if ( self.nextline >= MAX_SOURCE_LINES ) {
		self.setMode(self.run_finished_mode)
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
		self.basicError(PARSE, err.Error())
		self.setMode(MODE_QUIT)
		return
	}
	_, _ = self.interpret(leaf)
}

func (self *BasicRuntime) setMode(mode int) {
	self.mode = mode
	if ( self.mode == MODE_REPL ) {
		fmt.Println("READY")
	}
}

func (self *BasicRuntime) run(fileobj io.Reader, mode int) {
	var readbuff = bufio.NewScanner(fileobj)

	self.setMode(mode)
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
		if ( self.errno != 0 ) {
			self.setMode(self.run_finished_mode)
		}
		//fmt.Printf("Finishing in mode %d\n", self.mode)

	}
}
