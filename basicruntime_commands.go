package main

import (
	"fmt"
	"errors"
	"strings"
)

func (self *BasicRuntime) CommandDEF(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
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
	fmt.Println(rval.toString())
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
	self.nextline = rval.intval
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
	self.environment.gosubReturnLine = self.lineno + 1
	self.nextline = rval.intval
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandRETURN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	if ( self.environment.gosubReturnLine == 0 ) {
		return nil, errors.New("RETURN outside the context of GOSUB")
	}
	self.nextline = self.environment.gosubReturnLine
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
			fmt.Printf("%d %s\n", self.source[i].lineno, self.source[i].code)
		}
	}
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandRUN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	//fmt.Println("Processing RUN")
	self.autoLineNumber = 0
	if ( expr.right == nil ) {
		self.nextline = 0
	} else {
		rval, err = self.evaluate(expr.right)
		if ( err != nil ) {
			return nil, err
		}
		if ( rval.valuetype != TYPE_INTEGER ) {
			return nil, errors.New("Expected integer")
		}
		self.nextline = rval.intval
	}
	self.setMode(MODE_RUN)
	//fmt.Printf("Set mode %d with nextline %d\n", self.mode, self.nextline)
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

func (self *BasicRuntime) CommandFOR(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	// At this point the assignment has already been evaluated. We need to
	// evaluate the STEP expression if there is one, and the TO
	// leaf, and then return nil, nil.
	var err error = nil
	var assignvar *BasicValue = nil
	var tmpvar *BasicValue = nil
	var forConditionMet bool = false
		
	if ( self.environment.forToLeaf == nil || expr.right == nil ) {
		return nil, errors.New("Expected FOR ... TO [STEP ...]")
	}
	assignvar, err = self.evaluate(expr.right)
	if ( err != nil ) {
		return nil, err
	}
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
	forConditionMet, err = self.evaluateForCondition(assignvar)
	if ( forConditionMet == true ) {
		self.environment.waitForCommand("NEXT")
	}
	self.environment.forNextVariable = assignvar
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandNEXT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var forConditionMet = false
	var err error = nil

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
	self.environment.loopExitLine = self.lineno + 1

	//fmt.Println("Found NEXT %s, I'm waiting for NEXT %s\n", self.environment.forNextVariable.name, expr.right.identifier)
	if ( strings.Compare(expr.right.identifier, self.environment.forNextVariable.name) != 0 ) {
		self.prevEnvironment()
		return &self.staticFalseValue, nil
	}
	rval = self.environment.get(expr.right.identifier)
	forConditionMet, err = self.evaluateForCondition(rval)
	self.environment.stopWaiting("NEXT")
	if ( forConditionMet == true ) {
		//fmt.Println("Exiting loop")
		if ( self.environment.parent != nil ) {
			self.prevEnvironment()
		}
		return &self.staticTrueValue, nil
	}
	//fmt.Printf("Incrementing %s (%s) by %s\n", rval.name, rval.toString(), self.environment.forStepValue.toString())
	rval, err = rval.mathPlus(&self.environment.forStepValue)
	if ( err != nil ) {
		return nil, err
	}
	rval, err = self.environment.update(rval)
	if ( err != nil ) {
		return nil, err
	}
	//fmt.Println("Continuing loop")
	self.nextline = self.environment.loopFirstLine
	return &self.staticTrueValue, nil
}

func (self *BasicRuntime) CommandEXIT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {

	if ( self.environment.forToValue.valuetype == TYPE_UNDEFINED ) {
		return nil, errors.New("EXIT outside the context of FOR")
	}

	self.nextline = self.environment.loopExitLine
	self.prevEnvironment()
	return &self.staticTrueValue, nil
}
