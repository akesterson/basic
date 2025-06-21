package main

import (
	"fmt"
	"errors"
	"io"
	"bufio"
	"os"
	"slices"
	"reflect"
)

type BasicError int
const (
	NOERROR    BasicError = iota
	IO
	PARSE
	SYNTAX
	RUNTIME
)

type BasicSourceLine struct {
	code string
	lineno int64
}

type BasicRuntime struct {
	source [MAX_SOURCE_LINES]BasicSourceLine
	lineno int64
	values [MAX_VALUES]BasicValue
	variables [MAX_VARIABLES]BasicVariable
	staticTrueValue BasicValue
	staticFalseValue BasicValue
	nextvalue int
	nextvariable int
	nextline int64
	mode int
	errno BasicError
	run_finished_mode int
	scanner BasicScanner
	parser BasicParser
	environment *BasicEnvironment
	autoLineNumber int64
}

func (self *BasicRuntime) zero() {
	for i, _ := range self.values {
		self.values[i].init()
	}
	self.errno = 0
	self.nextvalue = 0
}

func (self *BasicRuntime) init() {
	self.environment = nil
	self.lineno = 0
	self.nextline = 0
	self.autoLineNumber = 0
	self.staticTrueValue.basicBoolValue(true)
	self.staticFalseValue.basicBoolValue(false)

	self.parser.init(self)
	self.scanner.init(self)
	self.newEnvironment()

	self.environment.functions["LEN"] = &BasicFunctionDef{
		arglist: &BasicASTLeaf{
			leaftype: LEAF_ARGUMENTLIST,
			left: nil,
			parent: nil,
			expr: nil,
			identifier: "",
			operator: FUNCTION_ARGUMENT,
			right: &BasicASTLeaf{
				leaftype: LEAF_IDENTIFIER,
				left: nil,
				parent: nil,
				expr: nil,
				identifier: "X$",
			},
		},
		expression: nil,
		runtime: self,
		name: "LEN",
	}
	self.environment.functions["MID"] = &BasicFunctionDef{
		arglist: &BasicASTLeaf{
			leaftype: LEAF_ARGUMENTLIST,
			left: nil,
			parent: nil,
			expr: nil,
			identifier: "",
			operator: FUNCTION_ARGUMENT,
			right: &BasicASTLeaf{
				leaftype: LEAF_IDENTIFIER,
				left: nil,
				parent: nil,
				expr: nil,
				identifier: "STR$",
				right: &BasicASTLeaf{
					leaftype: LEAF_IDENTIFIER_INT,
					identifier: "START#",
					right: &BasicASTLeaf{
						leaftype: LEAF_IDENTIFIER_INT,
						identifier: "LENGTH#",
					},
				},
			},
		},
		expression: nil,
		runtime: self,
		name: "LEN",
	}

	self.zero()
}

func (self *BasicRuntime) newEnvironment() {
	//fmt.Println("Creating new environment")
	var env *BasicEnvironment = new(BasicEnvironment)
	env.init(self, self.environment)
	self.environment = env
}

func (self *BasicRuntime) prevEnvironment() {
	if ( self.environment.parent == nil ) {
		self.basicError(RUNTIME, "No previous environment to return to")
		return
	}
	self.environment = self.environment.parent
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

func (self *BasicRuntime) newVariable() (*BasicVariable, error) {
	var variable *BasicVariable
	if ( self.nextvariable < MAX_VARIABLES ) {
		variable = &self.variables[self.nextvariable]
		self.nextvariable += 1
		variable.runtime = self
		return variable, nil
	}
	return nil, errors.New("Maximum runtime variables reached")
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

func (self *BasicRuntime) evaluateSome(expr *BasicASTLeaf, leaftypes ...BasicASTLeafType) (*BasicValue, error) {
	if ( slices.Contains(leaftypes, expr.leaftype)) {
		return self.evaluate(expr)
	}
	return nil, nil
}

func (self *BasicRuntime) evaluate(expr *BasicASTLeaf, leaftypes ...BasicASTLeafType) (*BasicValue, error) {
	var lval *BasicValue
	var rval *BasicValue
	var texpr *BasicASTLeaf
	var tval *BasicValue
	var err error = nil
	var subscripts []int64

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
	case LEAF_IDENTIFIER_FLOAT:
		// FIXME : How do I know if expr.right is an array subscript that I should follow,
		// or some other right-joined expression (like an argument list) which I should
		// *NOT* follow?
		texpr = expr.right
		if ( texpr != nil &&
			texpr.leaftype == LEAF_ARGUMENTLIST &&
			texpr.operator == ARRAY_SUBSCRIPT ) {
			texpr = texpr.right
			for ( texpr != nil ) {
				tval, err = self.evaluate(texpr)
				if ( err != nil ) {
					return nil, err
				}
				if ( tval.valuetype != TYPE_INTEGER ) {
					return nil, errors.New("Array dimensions must evaluate to integer (C)")
				}
				subscripts = append(subscripts, tval.intval)
				texpr = texpr.right
			}
		}
		fallthrough
	case LEAF_IDENTIFIER_STRING:
		if ( len(subscripts) == 0 ) {
			subscripts = append(subscripts, 0)
		}
		lval, err = self.environment.get(expr.identifier).getSubscript(subscripts...)
		if ( err != nil ) {
			return nil, err
		}
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
	case LEAF_FUNCTION:
		//fmt.Printf("Processing command %s\n", expr.identifier)
		lval, err = self.commandByReflection("Function", expr, lval, rval)
		if ( err != nil ) {
			lval, err = self.userFunction(expr, lval, rval)
			if ( err != nil ) {
				return nil, err
			} else if ( lval != nil ) {
				return lval, nil
			}
			return nil, err
		} else if ( lval != nil ) {
			return lval, nil
		}
	case LEAF_COMMAND_IMMEDIATE: fallthrough
	case LEAF_COMMAND:
		return self.commandByReflection("Command", expr, lval, rval)
		
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

func (self *BasicRuntime) userFunction(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var fndef *BasicFunctionDef = nil
	var leafptr *BasicASTLeaf = nil
	var argptr *BasicASTLeaf = nil
	var leafvalue *BasicValue = nil
	var err error = nil
	
	fndef = self.environment.getFunction(expr.identifier)
	//fmt.Printf("Function : %+v\n", fndef)
	if ( fndef == nil ) {
		return nil, nil
	} else {
		fndef.environment.init(self, self.environment)
		leafptr = expr.right
		argptr = fndef.arglist
		//fmt.Printf("Function arglist leaf: %s (%+v)\n", argptr.toString(), argptr)
		//fmt.Printf("Calling user function %s(", fndef.name)
		for ( leafptr != nil && argptr != nil) {
			//fmt.Printf("%+v\n", leafptr)
			leafvalue, err = self.evaluate(leafptr)
			if ( err != nil ) {
				return nil, err
			}
			//fmt.Printf("%s = %s, \n", argptr.toString(), leafvalue.toString())
			fndef.environment.set(argptr, leafvalue)
			leafptr = leafptr.right
			argptr = argptr.right
		}
		//fmt.Printf(")\n")
		self.environment = &fndef.environment
		//self.environment.dumpVariables()
		leafvalue, err = self.evaluate(fndef.expression)
		self.environment = fndef.environment.parent
		return leafvalue, err
		// return the result
	}
}

func (self *BasicRuntime) commandByReflection(rootKey string, expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
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
	rmethod = reflector.MethodByName(fmt.Sprintf("%s%s", rootKey, expr.identifier))
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
	if ( self.environment.isWaitingForAnyCommand() ) {
		if ( expr.leaftype != LEAF_COMMAND || !self.environment.isWaitingForCommand(expr.identifier) ) {
			//fmt.Printf("I am not waiting for %+v\n", expr)
			return &self.staticTrueValue, nil
		}
	}
	//fmt.Printf("Interpreting %+v\n", expr)
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
		//fmt.Println(err)
		return nil, err
	}
	return value, nil
}

func (self *BasicRuntime) findPreviousLineNumber() int64 {
	var i int64
	for i = self.lineno - 1; i > 0 ; i-- {
		if ( len(self.source[i].code) > 0 ) {
			return i
		}
	}
	return self.lineno
}

func (self *BasicRuntime) processLineRunStream(readbuff *bufio.Scanner) {
	var line string
	if ( readbuff.Scan() ) {
		line = readbuff.Text()
		// All we're doing is getting the line #
		// and storing the source line in this mode.
		self.scanner.scanTokens(line)
		self.source[self.lineno] = BasicSourceLine{
			code:   line,
			lineno: self.lineno}
	} else {
		self.setMode(MODE_RUN)
	}
}

func (self *BasicRuntime) processLineRepl(readbuff *bufio.Scanner) {
	var leaf *BasicASTLeaf = nil
	var value *BasicValue = nil
	var err error = nil
	var line string
	if ( self.autoLineNumber > 0 ) {
		fmt.Printf("%d ", (self.lineno + self.autoLineNumber))
	}
	if ( readbuff.Scan() ) {
		line = readbuff.Text()
		self.lineno += self.autoLineNumber
		line = self.scanner.scanTokens(line)
		for ( !self.parser.isAtEnd() ) {
			leaf, err = self.parser.parse()
			if ( err != nil ) {
				self.basicError(PARSE, err.Error())
				return
			}
			//fmt.Printf("%+v\n", leaf)
			//fmt.Printf("%+v\n", leaf.right)
			value, err = self.interpretImmediate(leaf)
			if ( value == nil ) {
				// Only store the line and increment the line number if we didn't run an immediate command
				self.source[self.lineno] = BasicSourceLine{
					code:   line,
					lineno: self.lineno}
			} else if ( self.autoLineNumber > 0 ) {
				self.lineno = self.findPreviousLineNumber()
				//fmt.Printf("Reset line number to %d\n", self.lineno)
			}
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
	line = self.source[self.nextline].code
	self.lineno = self.nextline
	self.nextline += 1
	if ( line == "" ) {
		return
	}
	//fmt.Println(line)
	self.scanner.scanTokens(line)
	for ( !self.parser.isAtEnd() ) {
		leaf, err = self.parser.parse()
		if ( err != nil ) {
			self.basicError(PARSE, err.Error())
			self.setMode(MODE_QUIT)
			return
		}
		_, _ = self.interpret(leaf)
	}
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
