package main

import (
	"fmt"
	"errors"
	"strings"
)

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
	return nil, nil	
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
	self.nextline = int(rval.intval)
	return nil, nil
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
	self.nextline = int(rval.intval)
	return nil, nil
}

func (self *BasicRuntime) CommandRETURN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	if ( self.environment.gosubReturnLine == 0 ) {
		return nil, errors.New("RETURN outside the context of GOSUB")
	}
	self.nextline = self.environment.gosubReturnLine
	return nil, nil
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
	return nil, nil
}

func (self *BasicRuntime) CommandRUN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	//fmt.Println("Processing RUN")
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
		self.nextline = int(rval.intval)
	}
	self.setMode(MODE_RUN)
	//fmt.Printf("Set mode %d with nextline %d\n", self.mode, self.nextline)
	return nil, nil
}

func (self *BasicRuntime) CommandQUIT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	self.setMode(MODE_QUIT)
	return nil, nil
}

func (self *BasicRuntime) CommandLET(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	// LET is not expressly required in our basic implementation or in
	// Commodore 128 BASIC 7.0. Assignments to variables are handled as
	// part of expression evaluation, LET doesn't need to manage it.
	return nil, nil
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
	return nil, nil
}

func (self *BasicRuntime) CommandFOR(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	// At this point the assignment has already been evaluated. We need to
	// evaluate the STEP expression if there is one, and the TO
	// leaf, and then return nil, nil.
	var err error = nil
	var assignvar *BasicValue = nil
	var tmpvar *BasicValue = nil
	var truth *BasicValue = nil
		
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
	if ( self.environment.forStepValue.intval == 0 && self.environment.forStepValue.floatval == 0 ) {
		// Set a default step
		truth, err = self.environment.forToValue.greaterThan(assignvar)
		if ( err != nil ) {
			return nil, err
		}
		if ( truth.isTrue() ) {
			self.environment.forStepValue.intval = 1
		} else {
			self.environment.forStepValue.intval = -1	
		}
	}
	self.environment.forToLeaf = nil
	self.environment.forStepLeaf = nil
	return nil, nil
}

func (self *BasicRuntime) CommandNEXT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var truth *BasicValue = nil
	var err error = nil

	// if self.environment.forRelationLeaf is nil, parse error
	if ( self.environment.forToValue.valuetype == TYPE_UNDEFINED ) {
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
	
	rval, err = self.environment.get(expr.right.identifier).mathPlus(&self.environment.forStepValue)
	if ( err != nil ) {
		return nil, err
	}
	truth, err = self.environment.forStepValue.lessThan(&BasicValue{valuetype: TYPE_INTEGER, intval: 0})
	if ( err != nil ) {
		return nil, err
	}
	if ( truth.isTrue() ) {
		// Our step is negative
		truth, err = self.environment.forToValue.greaterThan(rval)
	} else {
		// Our step is positive
		truth, err = self.environment.forToValue.lessThan(rval)
	}
	if ( err != nil ) {
		return nil, err
	}
	if ( truth.isTrue() ) {
		self.environment.forStepValue.zero()
		self.environment.forToValue.zero()
		self.environment.loopFirstLine = 0
		return nil, nil
	}
	self.nextline = self.environment.loopFirstLine
	return rval, nil
}

func (self *BasicRuntime) CommandEXIT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {

	if ( self.environment.forToValue.valuetype == TYPE_UNDEFINED ) {
		return nil, errors.New("EXIT outside the context of FOR")
	}

	self.environment.forStepValue.zero()
	self.environment.forToValue.zero()
	self.environment.loopFirstLine = 0
	self.nextline = self.environment.loopExitLine
	self.environment.loopExitLine = 0	
	return nil, nil
}
