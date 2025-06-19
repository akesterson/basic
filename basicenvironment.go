package main

import (
	"errors"
	"strings"
	"fmt"
)

type BasicEnvironment struct {
	variables map[string]*BasicVariable
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
	forNextValue *BasicValue
	forNextVariable *BasicVariable

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
	self.variables = make(map[string]*BasicVariable)
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
	if (len(self.waitingForCommand) != 0) {
		return true
	}
	if ( self.parent != nil ) {
		return self.parent.isWaitingForAnyCommand()
	}
	return false
}

func (self *BasicEnvironment) isWaitingForCommand(command string) bool {
	if (strings.Compare(self.waitingForCommand, command) == 0) {
		return true
	}
	if ( self.parent != nil ) {
		return self.parent.isWaitingForCommand(command)
	}
	return false
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

func (self *BasicEnvironment) get(varname string) *BasicVariable {
	var variable *BasicVariable
	var ok bool
 	sizes := []int64{10}
	if variable, ok = self.variables[varname]; ok {
		return variable
	} else if ( self.parent != nil ) {
		variable = self.parent.get(varname)
		if ( variable != nil ) {
			return variable
		}
	}
	// Don't automatically create variables unless we are the currently
	// active environment (parents don't create variables for their children)
	if ( self.runtime.environment == self ) {
		self.variables[varname] = &BasicVariable{
			name: strings.Clone(varname),
			valuetype: TYPE_UNDEFINED,
			runtime: self.runtime,
			mutable: true,
		}
		self.variables[varname].init(self.runtime, sizes)
		return self.variables[varname]
	}
	return nil
}

func (self *BasicEnvironment) set(lval *BasicASTLeaf, rval *BasicValue) {
	//fmt.Printf("Setting variable in environment: [%s] = %s\n", lval.toString(), rval.toString())
	self.get(lval.identifier).set(rval, 0)
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
	var variable *BasicVariable = nil
	if ( lval == nil || rval == nil ) {
		return nil, errors.New("nil pointer")
	}
	variable = self.get(lval.identifier)
	switch(lval.leaftype) {
	case LEAF_IDENTIFIER_INT:
		if ( rval.valuetype == TYPE_INTEGER ) {
			variable.setInteger(rval.intval, 0)
		} else if ( rval.valuetype == TYPE_FLOAT ) {
			variable.setInteger(int64(rval.floatval), 0)
		} else {
			return nil, errors.New("Incompatible types in variable assignment")
		}
	case LEAF_IDENTIFIER_FLOAT:
		if ( rval.valuetype == TYPE_INTEGER ) {
			variable.setFloat(float64(rval.intval), 0)
		} else if ( rval.valuetype == TYPE_FLOAT ) {
			variable.setFloat(rval.floatval, 0)
		} else {
			return nil, errors.New("Incompatible types in variable assignment")
		}
	case LEAF_IDENTIFIER_STRING:
		if ( rval.valuetype == TYPE_STRING ) {
			variable.setString(strings.Clone(rval.stringval), 0)
		} else {
			return nil, errors.New("Incompatible types in variable assignment")
		}
	default:
		return nil, errors.New("Invalid assignment")		
	}
	variable.valuetype = rval.valuetype
	//fmt.Printf("Assigned %+v\n", variable)
	return variable.getSubscript(0)
}
