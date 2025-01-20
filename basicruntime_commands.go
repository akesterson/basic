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

func (self *BasicRuntime) CommandLIST(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var startidx int64 = 0
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
		startidx = rval.intval
	}
	for _, value := range(self.source[startidx:]) {
		if ( len(value) > 0 ) {
			fmt.Println(value)
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
	var tmpvar *BasicValue = nil
		
	if ( self.environment.forToLeaf == nil || expr.right == nil ) {
		return nil, errors.New("Expected FOR ... TO [STEP ...]")
	}
	tmpvar, err = self.evaluate(expr.right)
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
	return nil, nil
}

func (self *BasicRuntime) CommandNEXT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var curValue float64 = 0.0
	var maxValue float64 = 0.0

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
	rval = self.environment.get(expr.right.identifier)
	
	if ( self.environment.forToValue.valuetype == TYPE_FLOAT ) {
		maxValue = self.environment.forToValue.floatval
	} else {
		maxValue = float64(self.environment.forToValue.intval)
	}	
	if ( self.environment.forStepValue.valuetype == TYPE_FLOAT ) {
		curValue = rval.floatval
	} else {
		curValue = float64(rval.intval)
	}

	if ( curValue == maxValue ) {
		self.environment.forStepValue.zero()
		self.environment.forToValue.zero()
		self.environment.forFirstLine = 0
		return nil, nil
	}
	if ( self.environment.forStepValue.valuetype == TYPE_FLOAT ) {
		rval.floatval += self.environment.forStepValue.floatval
	} else {
		rval.intval += self.environment.forStepValue.intval
	}
	self.nextline = self.environment.forFirstLine
	return nil, nil
}
