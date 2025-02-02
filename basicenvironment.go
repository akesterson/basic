package main

import (
	"errors"
	"strings"
	"fmt"
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
	forNextVariable *BasicValue

	// Loop variables
	loopFirstLine int64
	loopExitLine int64
	
	gosubReturnLine int64

	// When this is set, no lines are executed until a COMMAND
	// matching this string is found, then execution resumes.
	// This prevents us from automatically executing things
	// inside branches and loop structures which should be
	// skipped, when the actual evaluation of conditions is
	// performed at the bottom of those structures
	waitingForCommand string
	
	parent *BasicEnvironment
	runtime *BasicRuntime
}

func (self *BasicEnvironment) init(runtime *BasicRuntime, parent *BasicEnvironment) {
	self.variables = make(map[string]*BasicValue)
	self.functions = make(map[string]*BasicFunctionDef)
	self.parent = parent
	self.runtime = runtime
	self.forNextVariable = nil
	self.forStepLeaf = nil
	self.forToLeaf = nil
}

func (self *BasicEnvironment) waitForCommand(command string) {
	if ( len(self.waitingForCommand) != 0 ) {
		panic("Can't wait on multiple commands in the same environment")
	}
	//fmt.Printf("Environment will wait for command %s\n", command)
	self.waitingForCommand = command
}

func (self *BasicEnvironment) isWaitingForAnyCommand() bool {
	return (len(self.waitingForCommand) != 0)
}

func (self *BasicEnvironment) isWaitingForCommand(command string) bool {
	return (strings.Compare(self.waitingForCommand, command) == 0)
}

func (self *BasicEnvironment) stopWaiting(command string) {
	//fmt.Printf("Environment stopped waiting for command %s\n", command)
	self.waitingForCommand = ""	
}


func (self *BasicEnvironment) dumpVariables() {
	for key, value := range self.variables {
		fmt.Printf("variables[%s] = %s\n", key, value.toString())
	}
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
	//fmt.Printf("Setting variable in environment: [%s] = %s\n", lval.toString(), rval.toString())
	self.variables[lval.identifier] = rval
}

func (self *BasicEnvironment) update(rval *BasicValue) (*BasicValue, error){
	var leaf BasicASTLeaf
	switch (rval.valuetype) {
	case TYPE_INTEGER: leaf.leaftype = LEAF_IDENTIFIER_INT
	case TYPE_FLOAT: leaf.leaftype = LEAF_IDENTIFIER_FLOAT
	case TYPE_STRING: leaf.leaftype = LEAF_IDENTIFIER_STRING
	}
	leaf.identifier = rval.name
	return self.assign(&leaf, rval)
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
