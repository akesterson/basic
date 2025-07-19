package main

import (
	"fmt"
	"errors"
	"strings"
	"unsafe"
	//"os"
	"io"
	"github.com/veandco/go-sdl2/sdl"
	"bufio"
	"strconv"
)

func (self *BasicRuntime) CommandDEF(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandDIM(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var varref *BasicVariable
	var sizes []int64
	var err error = nil
	// DIM IDENTIFIER(LENGTH)
	// expr.right should be an identifier
	// expr.right->right should be an arglist
	if ( expr == nil ||
		expr.right == nil ||
		expr.right.right == nil ||
		expr.right.right.leaftype != LEAF_ARGUMENTLIST ||
		expr.right.right.operator != ARRAY_SUBSCRIPT ||
		expr.right.isIdentifier() == false ) {
		return nil, errors.New("Expected DIM IDENTIFIER(DIMENSIONS, ...)")
	}
	// Get the variable reference
	varref = self.environment.get(expr.right.identifier)
	if ( varref == nil ) {
		return nil, fmt.Errorf("Unable to get variable for identifier %s", expr.right.identifier)
	}
	// Evaluate the argument list and construct a list of sizes
	expr = expr.right.right.right
	for ( expr != nil ) {
		lval, err = self.evaluate(expr)
		if ( err != nil ) {
			return nil, err
		}
		if ( lval.valuetype != TYPE_INTEGER ) {
			return nil, errors.New("Array dimensions must evaluate to integer")
		}
		sizes = append(sizes, lval.intval)
		expr = expr.right
	}
	err = varref.init(self, sizes)
	if ( err != nil ) {
		return nil, err
	}
	varref.zero()
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandDLOAD(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var scanner *bufio.Scanner = nil
	var runtimemode int = self.mode
	if ( expr.right == nil ) {
		return nil, errors.New("Expected expression")
	}
	rval, err = self.evaluate(expr.right)
	if ( err != nil ) {
		return nil, err
	}
	if ( rval.valuetype != TYPE_STRING ) {
		return nil, errors.New("Expected STRING")
	}
	f := sdl.RWFromFile(rval.stringval, "r")
	if ( f == nil ) {
		return nil, sdl.GetError()
	}
	defer io.Closer.Close(f)
	scanner = bufio.NewScanner(f)
	for _, sourceline := range(self.source) {
		sourceline.code = ""
		sourceline.lineno = 0
	}
	self.environment.lineno = 0
	self.environment.nextline = 0
	// Not sure how it will work resetting the runtime's state
	// from within this function....
	
	for {
		self.zero()
		self.parser.zero()
		self.scanner.zero()
		self.processLineRunStream(scanner)
		if ( self.environment.nextline == 0 && self.mode == MODE_RUN ) {
			break
		}
	}
	self.setMode(runtimemode)
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandDSAVE(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	if ( expr.right == nil ) {
		return nil, errors.New("Expected expression")
	}
	rval, err = self.evaluate(expr.right)
	if ( err != nil ) {
		return nil, err
	}
	if ( rval.valuetype != TYPE_STRING ) {
		return nil, errors.New("Expected STRING")
	}
	f := sdl.RWFromFile(rval.stringval, "w")
	if ( f == nil ) {
		return nil, sdl.GetError()
	}
	defer io.Closer.Close(f)
	for _, sourceline := range(self.source) {
		if ( len(sourceline.code) == 0 ) {
			continue
		}
		f.Write([]byte(fmt.Sprintf("%d %s\n", sourceline.lineno, sourceline.code)))
	}
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandLABEL(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error
	// LABEL IDENTIFIER
	// expr.right should be an identifier
	if ( expr == nil ||
		expr.right == nil ||
		expr.right.isIdentifier() == false ) {
		return nil, errors.New("Expected LABEL IDENTIFIER")
	}
	err = self.environment.setLabel(expr.right.identifier, self.environment.lineno)
	if ( err != nil ) {
		return &self.staticFalseValue, err
	}
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandPRINT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	if ( expr.right == nil ) {
		return nil, errors.New("Expected expression")
	}
	rval, err = self.evaluate(expr.right)
	if ( err != nil ) {
		return nil, err
	}
	self.Println(rval.toString())
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandGOTO(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	if ( expr.right == nil ) {
		return nil, errors.New("Expected expression")
	}
	rval, err = self.evaluate(expr.right)
	if ( err != nil ) {
		return nil, err
	}
	if ( rval.valuetype != TYPE_INTEGER ) {
		return nil, errors.New("Expected integer")
	}
	self.environment.nextline = rval.intval
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandGOSUB(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	if ( expr.right == nil ) {
		return nil, errors.New("Expected expression")
	}
	rval, err = self.evaluate(expr.right)
	if ( err != nil ) {
		return nil, err
	}
	if ( rval.valuetype != TYPE_INTEGER ) {
		return nil, errors.New("Expected integer")
	}
	self.newEnvironment()
	self.environment.gosubReturnLine = self.environment.lineno + 1
	self.environment.nextline = rval.intval
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandPOKE(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var addr uintptr
	var ptr unsafe.Pointer
	var typedPtr *byte
	
	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	expr = expr.firstArgument()
	if (expr != nil) {
		self.eval_clone_identifiers = false
		lval, err = self.evaluate(expr)
		self.eval_clone_identifiers = true
		if ( err != nil ) {
			return nil, err
		}
		if ( lval.valuetype != TYPE_INTEGER ) {
			return nil, errors.New("POKE expected INTEGER, INTEGER")
		}
		if ( expr.right == nil ||
			expr.right.leaftype != LEAF_LITERAL_INT &&
			expr.right.leaftype != LEAF_IDENTIFIER_INT) {
			return nil, errors.New("POKE expected INTEGER, INTEGER")
		}
		rval, err = self.evaluate(expr.right)
		
		addr = uintptr(lval.intval)
		//fmt.Printf("addr: %v\n", addr)
		ptr = unsafe.Pointer(addr)
		typedPtr = (*byte)(ptr)
		//fmt.Printf("Before set: %d\n", *typedPtr)
		*typedPtr = byte(rval.intval)
		//fmt.Printf("After set: %d\n", *typedPtr)
		return &self.staticTrueValue, nil
	}
	return nil, errors.New("POKE expected INTEGER, INTEGER")
}


func (self *BasicRuntime) CommandRETURN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error
	if ( self.environment.isWaitingForCommand("RETURN") ) {
		// we probably got here from a DEF line and should not execute, just return
		self.environment.stopWaiting("RETURN")
		return &self.staticTrueValue, nil
	}
	if ( self.environment.gosubReturnLine == 0 ) {
		return nil, errors.New("RETURN outside the context of GOSUB")
	}
	//fmt.Printf("RETURN : %s\n", expr.toString())
	if ( expr.right != nil ) {
		rval, err = self.evaluate(expr.right)
	} else {
		rval = &self.staticTrueValue
		err = nil
	}
	self.environment.nextline = self.environment.gosubReturnLine
	self.environment = self.environment.parent
	// if ( rval != nil ) {
	// 	fmt.Printf("RETURNing %s\n", rval.toString())
	// } else {
	// 	fmt.Printf("RETURN got an expression but it evaluated to nil : %s\n", err)
	// }
	return rval, err
}


func (self *BasicRuntime) CommandDELETE(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var startidx int64 = 0
	var endidx int64 = MAX_SOURCE_LINES - 1
	var i int64

	if ( expr.right != nil ) {
		if ( expr.right.leaftype == LEAF_LITERAL_INT ) {
			rval, err = self.evaluate(expr.right)
			if ( err != nil ) {
				return nil, err
			}
			if ( rval.valuetype != TYPE_INTEGER ) {
				return nil, errors.New("Expected integer")
			}
			startidx = rval.intval
		} else if ( expr.right.leaftype == LEAF_BINARY &&
			expr.right.operator == MINUS ) {
			lval, err = self.evaluate(expr.right.left)
			if ( err != nil ) {
				return nil, err
			}
			if ( lval.valuetype != TYPE_INTEGER ) {
				return nil, errors.New("Expected integer")
			}			
			rval, err = self.evaluate(expr.right.right)
			if ( err != nil ) {
				return nil, err
			}
			if ( rval.valuetype != TYPE_INTEGER ) {
				return nil, errors.New("Expected integer")
			}
			startidx = lval.intval
			endidx = rval.intval
		} else if ( expr.right.leaftype == LEAF_UNARY &&
			expr.right.operator == MINUS ) {
			rval, err = self.evaluate(expr.right.right)
			if ( err != nil ) {
				return nil, err
			}
			if ( rval.valuetype != TYPE_INTEGER ) {
				return nil, errors.New("Expected integer")
			}
			endidx = rval.intval
		}
	}
	for i = startidx; i <= endidx; i++ {
		if ( len(self.source[i].code) > 0 ) {
			self.source[i].code = ""
		}
	}
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandLIST(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var startidx int64 = 0
	var endidx int64 = MAX_SOURCE_LINES - 1
	var i int64

	if ( expr.right != nil ) {
		if ( expr.right.leaftype == LEAF_LITERAL_INT ) {
			rval, err = self.evaluate(expr.right)
			if ( err != nil ) {
				return nil, err
			}
			if ( rval.valuetype != TYPE_INTEGER ) {
				return nil, errors.New("Expected integer")
			}
			startidx = rval.intval
		} else if ( expr.right.leaftype == LEAF_BINARY &&
			expr.right.operator == MINUS ) {
			lval, err = self.evaluate(expr.right.left)
			if ( err != nil ) {
				return nil, err
			}
			if ( lval.valuetype != TYPE_INTEGER ) {
				return nil, errors.New("Expected integer")
			}			
			rval, err = self.evaluate(expr.right.right)
			if ( err != nil ) {
				return nil, err
			}
			if ( rval.valuetype != TYPE_INTEGER ) {
				return nil, errors.New("Expected integer")
			}
			startidx = lval.intval
			endidx = rval.intval
		} else if ( expr.right.leaftype == LEAF_UNARY &&
			expr.right.operator == MINUS ) {
			rval, err = self.evaluate(expr.right.right)
			if ( err != nil ) {
				return nil, err
			}
			if ( rval.valuetype != TYPE_INTEGER ) {
				return nil, errors.New("Expected integer")
			}
			endidx = rval.intval
		}
	}
	for i = startidx; i <= endidx; i++ {
		if ( len(self.source[i].code) > 0 ) {
			self.Println(fmt.Sprintf("%d %s", self.source[i].lineno, self.source[i].code))
		}
	}
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandRUN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	//fmt.Println("Processing RUN")
	self.autoLineNumber = 0
	if ( expr.right == nil ) {
		self.environment.nextline = 0
	} else {
		rval, err = self.evaluate(expr.right)
		if ( err != nil ) {
			return nil, err
		}
		if ( rval.valuetype != TYPE_INTEGER ) {
			return nil, errors.New("Expected integer")
		}
		self.environment.nextline = rval.intval
	}
	self.setMode(MODE_RUN)
	//fmt.Printf("Set mode %d with nextline %d\n", self.mode, self.environment.nextline)
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandINPUT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error 
	var promptmsg *BasicValue = nil
	var assignment *BasicASTLeaf = nil
	var assignValue BasicASTLeaf

	if ( expr == nil || expr.right == nil || expr.right.left == nil || expr.right.isIdentifier() == false ) {
		return nil, errors.New("Expected INPUT \"PROMPT\" IDENTIFIER")
	}
	promptmsg, err = self.evaluate(expr.right.left)
	if ( err != nil ) {
		return nil, err
	}
	self.Write(promptmsg.stringval)
	self.drawPrintBuffer()
	// get the string from the user
	for ( len(self.userline) == 0 ) {
		self.sdlEvents()
	}
	
	assignment, err = self.parser.newLeaf()
	if ( err != nil ) {
		return nil, err
	}
	switch (expr.right.leaftype) {
	case LEAF_IDENTIFIER_STRING:
		assignValue.leaftype = LEAF_LITERAL_STRING
		assignValue.literal_string = self.userline
		assignValue.operator = LITERAL_STRING
	case LEAF_IDENTIFIER_INT:
		assignValue.leaftype = LEAF_LITERAL_INT
		var i int
		i, err = strconv.Atoi(self.userline)
		if ( err != nil ) {
			return nil, err
		}
		assignValue.literal_int = int64(i)
		assignValue.operator = LITERAL_INT
	case LEAF_IDENTIFIER_FLOAT:
		assignValue.leaftype = LEAF_LITERAL_FLOAT
		assignValue.literal_float, err = strconv.ParseFloat(self.userline, 64)
		if ( err != nil ) {
			return nil, err
		}
		assignValue.operator = LITERAL_FLOAT
	}
	assignment.newBinary(expr.right, ASSIGNMENT, &assignValue)
	self.evaluate(assignment)
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandAUTO(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	if ( expr.right == nil ) {
		//fmt.Println("Turned AUTO off")
		self.autoLineNumber = 0
	} else {
		rval, err = self.evaluate(expr.right)
		if ( err != nil ) {
			return nil, err
		}
		if ( rval.valuetype != TYPE_INTEGER ) {
			return nil, errors.New("Expected integer")
		}
		self.autoLineNumber = rval.intval
		//fmt.Printf("Turned AUTO on: %d\n", self.autoLineNumber)
	}
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandQUIT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	self.setMode(MODE_QUIT)
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandLET(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	// LET is not expressly required in our basic implementation or in
	// Commodore 128 BASIC 7.0. Assignments to variables are handled as
	// part of expression evaluation, LET doesn't need to manage it.
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandIF(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var actionclause *BasicASTLeaf = nil
	if ( expr.right == nil ) {
		return nil, errors.New("Expected IF ... THEN")
	}
	rval, err = self.evaluate(expr.right)
	if ( err != nil ) {
		return nil, err
	}
	if ( rval.boolvalue == BASIC_TRUE ) {
		for ( expr.right != nil ) {
			expr = expr.right
			if ( expr.leaftype == LEAF_COMMAND && strings.Compare(expr.identifier, "THEN") == 0 ) {
				actionclause = expr.right
			}
		}
		if ( expr == nil || expr.right == nil ) {
			return nil, errors.New("Malformed IF statement")
		}
		return self.evaluate(actionclause)
	}
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) evaluateForCondition(rval *BasicValue) (bool, error) {
	var truth *BasicValue = nil
	var err error = nil
	if ( rval == nil ) {
		return false, errors.New("NIL pointer for rval")
	}
	truth, err = self.environment.forStepValue.lessThan(&BasicValue{valuetype: TYPE_INTEGER, intval: 0})
	if ( err != nil ) {
		return false, err
	}
	if ( truth.isTrue() ) {
		// Our step is negative
		truth, err = self.environment.forToValue.greaterThanEqual(rval)
	} else {
		// Our step is positive
		truth, err = self.environment.forToValue.lessThanEqual(rval)
	}
	if ( err != nil ) {
		return false, err
	}

	//fmt.Printf("%s ? %s : %s\n",
	//rval.toString(),
	//self.environment.forToValue.toString(),
	//truth.toString())

	return truth.isTrue(), nil
}

func (self *BasicRuntime) CommandREAD(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	self.environment.waitForCommand("DATA")
	self.environment.readIdentifierIdx = 0
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandSTOP(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	self.setMode(MODE_REPL)
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandDATA(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var curIdentifier *BasicASTLeaf = nil
	var curAssignCommand BasicASTLeaf
	var err error
	if ( expr == nil || expr.right == nil ) {
		return nil, errors.New("NIL expression or argument list")
	}
	expr = expr.right.right
	for ( expr != nil ) {
		curIdentifier = self.environment.readIdentifierLeaves[self.environment.readIdentifierIdx]
		if ( curIdentifier == nil ) {
			break
		}
		curAssignCommand.newBinary(curIdentifier, ASSIGNMENT, expr)
		_, err = self.evaluate(&curAssignCommand)
		if ( err != nil ) {
			return nil, err
		}
		self.environment.readIdentifierIdx += 1
		expr = expr.right
	}
	if ( expr == nil &&
		self.environment.readIdentifierIdx < MAX_LEAVES &&
		self.environment.readIdentifierLeaves[self.environment.readIdentifierIdx] != nil ) {
		// We ran out of DATA and still have READ items left to fulfill, leave us in waitingFor mode
		return &self.staticTrueValue, nil
	}
	// we fulfilled all our READ items, exit waitingFor mode
	self.environment.stopWaiting("DATA")
	self.environment.lineno = self.environment.readReturnLine
	self.environment.readIdentifierIdx = 0
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandFOR(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	// At this point the assignment has already been evaluated. We need to
	// evaluate the STEP expression if there is one, and the TO
	// leaf, and then return nil, nil.
	var err error = nil
	var assignval *BasicValue = nil
	var tmpvar *BasicValue = nil
	var forConditionMet bool = false
		
	if ( self.environment.forToLeaf == nil || expr.right == nil ) {
		return nil, errors.New("Expected FOR ... TO [STEP ...]")
	}
	if ( expr.right.left == nil || (
		expr.right.left.leaftype != LEAF_IDENTIFIER_INT &&
			expr.right.left.leaftype != LEAF_IDENTIFIER_FLOAT &&
			expr.right.left.leaftype != LEAF_IDENTIFIER_STRING) ) {
		return nil, errors.New("Expected variable in FOR loop")
	}
	assignval, err = self.evaluate(expr.right)
	if ( err != nil ) {
		return nil, err
	}
	self.environment.forNextVariable = self.environment.get(expr.right.left.identifier)
	self.environment.forNextVariable.set(assignval, 0)


	tmpvar, err = self.evaluate(self.environment.forToLeaf)
	if ( err != nil ) {
		return nil, err
	}
	_, _ = tmpvar.clone(&self.environment.forToValue)
	tmpvar, err = self.evaluate(self.environment.forStepLeaf)
	if ( err != nil ) {
		return nil, err
	}
	_, _ = tmpvar.clone(&self.environment.forStepValue)
	self.environment.forToLeaf = nil
	self.environment.forStepLeaf = nil
	tmpvar, err = self.environment.forNextVariable.getSubscript(0)
	if (err != nil ) {
		return nil, err
	}
	forConditionMet, err = self.evaluateForCondition(tmpvar)
	if ( forConditionMet == true ) {
		self.environment.waitForCommand("NEXT")
	}
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandNEXT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var forConditionMet = false
	var err error = nil
	var nextvar *BasicVariable

	// if self.environment.forRelationLeaf is nil, parse error
	if ( self.environment.forNextVariable == nil ) {
		return nil, errors.New("NEXT outside the context of FOR")
	}

	if ( expr.right == nil ) {
		return nil, errors.New("Expected NEXT IDENTIFIER")
	}
	if ( expr.right.leaftype != LEAF_IDENTIFIER_INT &&
		expr.right.leaftype != LEAF_IDENTIFIER_FLOAT ) {
		return nil, errors.New("FOR ... NEXT only valid over INT and FLOAT types")
	}
	self.environment.loopExitLine = self.environment.lineno + 1

	//fmt.Println("Found NEXT %s, I'm waiting for NEXT %s\n", self.environment.forNextVariable.name, expr.right.identifier)
	if ( strings.Compare(expr.right.identifier, self.environment.forNextVariable.name) != 0 ) {
		self.environment.parent.nextline = self.environment.nextline
		self.prevEnvironment()
		return &self.staticFalseValue, nil
	}
	nextvar = self.environment.get(expr.right.identifier)
	rval, err = nextvar.getSubscript(0)
	if ( err != nil ) {
		return nil, err
	}
	forConditionMet, err = self.evaluateForCondition(rval)
	self.environment.stopWaiting("NEXT")
	if ( forConditionMet == true ) {
		//fmt.Println("Exiting loop")
		if ( self.environment.parent != nil ) {
			self.environment.parent.nextline = self.environment.nextline
			self.prevEnvironment()
		}
		return &self.staticTrueValue, nil
	}
	//fmt.Printf("Incrementing %s (%s) by %s\n", rval.name, rval.toString(), self.environment.forStepValue.toString())
	rval, err = rval.mathPlus(&self.environment.forStepValue)
	if ( err != nil ) {
		return nil, err
	}
	//fmt.Println("Continuing loop")
	self.environment.nextline = self.environment.loopFirstLine
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandEXIT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {

	if ( self.environment.forToValue.valuetype == TYPE_UNDEFINED ) {
		return nil, errors.New("EXIT outside the context of FOR")
	}

	self.environment.nextline = self.environment.loopExitLine
	self.prevEnvironment()
	return &self.staticTrueValue, nil
}

