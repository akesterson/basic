package main

import (
	"errors"
	"strings"
)

type BasicEnvironment struct {
	variables map[string]*BasicValue
	functions map[string]*BasicFunctionDef
	
	// IF variables
	ifThenLine int64
	ifElseLine int64
	ifCondition BasicASTLeaf
	
	
	// FOR variables
	forStepLeaf *BasicASTLeaf
	forStepValue BasicValue
	forToLeaf *BasicASTLeaf
	forToValue BasicValue

	// Loop variables
	loopFirstLine int64
	loopExitLine int64
	
	gosubReturnLine int64

	parent *BasicEnvironment
	runtime *BasicRuntime
}

func (self *BasicEnvironment) init(runtime *BasicRuntime, parent *BasicEnvironment) {
	self.variables = make(map[string]*BasicValue)
	self.functions = make(map[string]*BasicFunctionDef)
	self.parent = parent
	self.runtime = runtime
}

func (self *BasicEnvironment) getFunction(fname string) *BasicFunctionDef {
	if value, ok := self.functions[fname]; ok {
		return value
	} else if ( self.parent != nil ) {
		return self.parent.getFunction(fname)
	}
	return nil
}

func (self *BasicEnvironment) get(varname string) *BasicValue {
	var value *BasicValue
	var ok bool
	if value, ok = self.variables[varname]; ok {
		return value
	} else if ( self.parent != nil ) {
		value = self.parent.get(varname)
		if ( value != nil ) {
			return value
		}
	}
	// Don't automatically create variables unless we are the currently
	// active environment (parents don't create variables for their children)
	if ( self.runtime.environment == self ) {
		self.variables[varname] = &BasicValue{
			name: strings.Clone(varname),
			valuetype: TYPE_UNDEFINED,
			stringval: "",
			intval: 0,
			floatval: 0.0,
			boolvalue: BASIC_FALSE,
			runtime: self.runtime,
			mutable: true}
		return self.variables[varname]
	}
	return nil
}

func (self *BasicEnvironment) set(lval *BasicASTLeaf, rval *BasicValue) {
	self.variables[lval.identifier] = rval
}

func (self *BasicEnvironment) assign(lval *BasicASTLeaf , rval *BasicValue) (*BasicValue, error) {
	var variable *BasicValue = nil
	if ( lval == nil || rval == nil ) {
		return nil, errors.New("nil pointer")
	}
	variable = self.get(lval.identifier)
	switch(lval.leaftype) {
	case LEAF_IDENTIFIER_INT:
		if ( rval.valuetype == TYPE_INTEGER ) {
			variable.intval = rval.intval
		} else if ( rval.valuetype == TYPE_FLOAT ) {
			variable.intval = int64(rval.floatval)
		} else {
			return nil, errors.New("Incompatible types in variable assignment")
		}
	case LEAF_IDENTIFIER_FLOAT:
		if ( rval.valuetype == TYPE_INTEGER ) {
			variable.floatval = float64(rval.intval)
		} else if ( rval.valuetype == TYPE_FLOAT ) {
			variable.floatval = rval.floatval
		} else {
			return nil, errors.New("Incompatible types in variable assignment")
		}
	case LEAF_IDENTIFIER_STRING:
		if ( rval.valuetype == TYPE_STRING ) {
			variable.stringval = strings.Clone(rval.stringval)
		} else {
			return nil, errors.New("Incompatible types in variable assignment")
		}
	default:
		return nil, errors.New("Invalid assignment")		
	}
	variable.valuetype = rval.valuetype
	//fmt.Printf("Assigned %+v\n", variable)
	return variable, nil
}
