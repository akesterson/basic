package main

import (
	"fmt"
	"errors"
	"io"
	"bufio"
	//"os"
	"slices"
	"unicode"
	"strings"
	"reflect"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"	
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
	
	lineInProgress [MAX_LINE_LENGTH]rune
	userlineIndex int
	userline string

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
	// The default behavior for evaluate() is to clone any value that comes from
	// an identifier. This allows expressions like `I# + 1` to return a new value
	// without modifying I#. However some commands (like POINTER), when they are
	// evaluating an identifier, do not want the cloned value, they want the raw
	// source value. Those commands will temporarily set this to `false`.
	eval_clone_identifiers bool
	window *sdl.Window
	printSurface *sdl.Surface
	cursorX int32
	cursorY int32
	
	font *ttf.Font
	fontWidth int
	fontHeight int
	maxCharsW int32
	maxCharsH int32

	printBuffer string
}

func (self *BasicRuntime) zero() {
	for i, _ := range self.values {
		self.values[i].init()
	}
	self.printBuffer = ""
	self.errno = 0
	self.nextvalue = 0
	self.userline = ""
	self.eval_clone_identifiers = true
}

func (self *BasicRuntime) init(window *sdl.Window, font *ttf.Font) {
	var err error = nil
	var windowSurface *sdl.Surface = nil
	
	self.environment = nil
	self.lineno = 0
	self.nextline = 0
	self.autoLineNumber = 0
	self.staticTrueValue.basicBoolValue(true)
	self.staticFalseValue.basicBoolValue(false)

	self.parser.init(self)
	self.scanner.init(self)
	self.newEnvironment()

	self.eval_clone_identifiers = true
	self.window = window
	self.font = font

	self.fontWidth, self.fontHeight, err = self.font.SizeUTF8("A")
	if ( err != nil ) {
		self.basicError(RUNTIME, "Could not get the height and width of the font")
	} else {
		windowSurface, err = self.window.GetSurface()
		if ( err != nil ) {
			self.basicError(RUNTIME, "Could not get SDL window surface")
		} else {
			self.maxCharsW = (windowSurface.W / int32(self.fontWidth))
			self.maxCharsH = (windowSurface.H / int32(self.fontHeight))-1
		}
	}
	self.printSurface, err = sdl.CreateRGBSurface(0, windowSurface.W, windowSurface.H, int32(windowSurface.Format.BitsPerPixel), 0, 0, 0, 0)
	if ( err != nil ) {
		self.basicError(RUNTIME, "Could not create the print buffer surface")
	}
	
	self.zero()
	self.parser.zero()
	self.scanner.zero()
	self.initFunctions()
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
	self.Println(fmt.Sprintf("? %d : %s %s\n", self.lineno, self.errorCodeToString(errno), message))
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
	case LEAF_IDENTIFIER_FLOAT: fallthrough
	case LEAF_IDENTIFIER_STRING:
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
		if ( self.eval_clone_identifiers == false ) {
			return lval, nil
		} else {
			return lval.clone(nil)
		}
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
			return nil, err
		} else if ( lval == nil ) {
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
		lval, err = self.commandByReflection("Command", expr, lval, rval)
		if ( err != nil ) {
			return nil, err
		} else if ( lval == nil ) {
			return nil, fmt.Errorf("Unknown command %s", expr.identifier)
		}
		return lval, err
		
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
	
	fndef = self.environment.getFunction(strings.ToUpper(expr.identifier))
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
	rmethod = reflector.MethodByName(fmt.Sprintf("%s%s", rootKey, strings.ToUpper(expr.identifier)))
	if ( !rmethod.IsValid() ) {
		return nil, nil
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
	// All we're doing is getting the line #
	// and storing the source line in this mode.
	if ( readbuff.Scan() ) {
		line = readbuff.Text()
		//fmt.Printf("processLineRunStream loaded %s\n", line)
		if ( self.mode == MODE_REPL ) {
			// DLOAD calls this method from inside of
			// MODE_REPL. In that case we want to strip the
			// line numbers off the beginning of the lines
			// the same way we do in the repl.
			line = self.scanner.scanTokens(line)
		} else {
			self.scanner.scanTokens(line)
		}
		self.source[self.lineno] = BasicSourceLine{
			code:   line,
			lineno: self.lineno}
	} else {
		//fmt.Printf("processLineRunStream exiting\n")
		self.nextline = 0
		self.setMode(MODE_RUN)
	}
}

func (self *BasicRuntime) processLineRepl(readbuff *bufio.Scanner) {
	var leaf *BasicASTLeaf = nil
	var value *BasicValue = nil
	var err error = nil
	if ( self.autoLineNumber > 0 ) {
		fmt.Printf("%d ", (self.lineno + self.autoLineNumber))
	}
	// get a new line from the keyboard
	if ( len(self.userline) > 0 ) {
		self.lineno += self.autoLineNumber
		self.userline = self.scanner.scanTokens(self.userline)
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
					code:   self.userline,
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
		self.Println("READY")
	}
}

func (self *BasicRuntime) sdlEvents() error {
	var ir rune
	var sb strings.Builder
	var i int
	var err error
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			self.setMode(MODE_QUIT)
		case *sdl.TextInputEvent:
			// This is LAZY and WRONG but it works on US ASCII keyboards so I guess
			// international users go EFF themselves? It's how we did it in the old days...
			ir = rune(t.Text[0])
			if ( unicode.IsPrint(ir) ) {
				self.lineInProgress[self.userlineIndex] = ir
				self.userlineIndex += 1
				err = self.drawText(
					(self.cursorX * int32(self.fontWidth)),
					(self.cursorY * int32(self.fontHeight)),
					string(ir),
					true)
				if ( err != nil ) {
					fmt.Println(err)
					return err
				}
				self.advanceCursor(1, 0)
			}
		case *sdl.KeyboardEvent:
			if ( t.Type == sdl.KEYUP ) {
				//fmt.Printf("Key released: %s (Scancode: %d, Keycode: %d)\n", sdl.GetKeyName(t.Keysym.Sym), t.Keysym.Scancode, t.Keysym.Sym)
				ir = self.runeForSDLScancode(t.Keysym)
				//fmt.Printf("Rune: %c", ir)
				if ( ir == sdl.K_BACKSPACE ) {
					if ( self.userlineIndex == 0 ) {
						return nil
					}
					self.lineInProgress[self.userlineIndex-1] = 0
					self.userlineIndex -= 1
					err = self.drawText(
						(self.cursorX * int32(self.fontWidth)),
						(self.cursorY * int32(self.fontHeight)),
						" ",
						true)
					if ( err != nil ) {
						return err
					}
					self.advanceCursor(-1, 0)
					err = self.drawText(
						(self.cursorX * int32(self.fontWidth)),
						(self.cursorY * int32(self.fontHeight)),
						" ",
						true)
					if ( err != nil ) {
						return err
					}
				} else if ( ir == sdl.K_RETURN || ir == '\n' ) {
					self.userline = ""
					for i = 0; i <= self.userlineIndex; i++  {
						if ( self.lineInProgress[i] == 0 ) {
							break
						}
						sb.WriteRune(self.lineInProgress[i])
						self.lineInProgress[i] = 0
					}
					//fmt.Printf("\n")
					self.userline = sb.String()
					self.userlineIndex = 0
					//fmt.Println(self.userline)
					//self.Println(self.userline)
					self.advanceCursor(-(self.cursorX), 1)
				}
			}
		}
	}
	return nil
}

func (self *BasicRuntime) runeForSDLScancode(keysym sdl.Keysym) rune {
	var rc rune = 0
	var keyboardstate []uint8
	rc = rune(keysym.Sym)
	keyboardstate = sdl.GetKeyboardState()
	if ( keyboardstate[sdl.SCANCODE_LSHIFT] != 0 ||
		keyboardstate[sdl.SCANCODE_RSHIFT] != 0 ) {
		if ( unicode.IsUpper(rc) ) {
			return unicode.ToLower(rc)
		}
		return unicode.ToUpper(rc)
	}
	return rc
}

func (self *BasicRuntime) run(fileobj io.Reader, mode int) {
	var readbuff = bufio.NewScanner(fileobj)
	var err error

	self.setMode(mode)
	if ( self.mode == MODE_REPL ) {
		self.run_finished_mode = MODE_REPL
		sdl.StartTextInput()
	} else {
		self.run_finished_mode = MODE_QUIT
	}
	for {
		//fmt.Printf("Starting in mode %d\n", self.mode)
		self.drawPrintBuffer()
		self.zero()
		self.parser.zero()
		self.scanner.zero()
		switch (self.mode) {
		case MODE_QUIT:
			return
		case MODE_RUNSTREAM:
			self.processLineRunStream(readbuff)
		case MODE_REPL:
			err = self.sdlEvents()
			if ( err != nil ) {
				self.basicError(RUNTIME, err.Error())
			}
			err = self.drawCursor()
			if ( err != nil ) {
				self.basicError(RUNTIME, err.Error())
			}
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
