package main

import (
	"errors"
	"strings"
	"fmt"
)

type BasicEnvironment struct {
	variables map[string]*BasicVariable
	functions map[string]*BasicFunctionDef
	labels map[string]int64
	
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

	// READ command variables
	readReturnLine int64
	readIdentifierLeaves [MAX_LEAVES]*BasicASTLeaf
	readIdentifierIdx int64

	// When this is set, no lines are executed until a COMMAND
	// matching this string is found, then execution resumes.
	// This prevents us from automatically executing things
	// inside branches and loop structures which should be
	// skipped, when the actual evaluation of conditions is
	// performed at the bottom of those structures
	waitingForCommand string
	
	parent *BasicEnvironment
	runtime *BasicRuntime

	// runtime bits
	lineno int64
	values [MAX_VALUES]BasicValue
	nextvalue int
	nextline int64
	errno BasicError
	// The default behavior for evaluate() is to clone any value that comes from
	// an identifier. This allows expressions like `I# + 1` to return a new value
	// without modifying I#. However some commands (like POINTER), when they are
	// evaluating an identifier, do not want the cloned value, they want the raw
	// source value. Those commands will temporarily set this to `false`.
	eval_clone_identifiers bool
	returnValue BasicValue

	// parser bits
	tokens [MAX_TOKENS]BasicToken	
	nexttoken int
	curtoken int
	leaves [MAX_TOKENS]BasicASTLeaf
	nextleaf int
	errorToken *BasicToken
}

func (self *BasicEnvironment) init(runtime *BasicRuntime, parent *BasicEnvironment) {
	self.variables = make(map[string]*BasicVariable)
	self.functions = make(map[string]*BasicFunctionDef)
	self.labels = make(map[string]int64)
	self.parent = parent
	self.runtime = runtime
	self.forNextVariable = nil
	self.forStepLeaf = nil
	self.forToLeaf = nil
	if ( self.parent != nil ) {
		self.lineno = self.parent.lineno
		self.nextline = self.parent.nextline
		self.eval_clone_identifiers = self.parent.eval_clone_identifiers
	} else {
		self.lineno = 0
		self.nextline = 0
		self.eval_clone_identifiers = true
	}
	self.zero_parser_variables()
}

func (self *BasicEnvironment) zero() {
	for i, _ := range self.values {
		self.values[i].init()
	}
	self.nextvalue = 0
	self.errno = 0
	self.eval_clone_identifiers = true
}

func (self *BasicEnvironment) zero_parser_variables() {
	for i, _ := range self.leaves {
		self.leaves[i].init(LEAF_UNDEFINED)
	}
	for i, _ := range self.tokens {
		self.tokens[i].init()
	}
	self.curtoken = 0
	self.nexttoken = 0
	self.nextleaf = 0
}

func (self *BasicEnvironment) newValue() (*BasicValue, error) {
       var value *BasicValue
       if ( self.nextvalue < MAX_VALUES ) {
               value = &self.values[self.nextvalue]
               self.nextvalue += 1
               value.runtime = self.runtime
               return value, nil
       }
       return nil, errors.New("Maximum values per line reached")
}

func (self *BasicEnvironment) waitForCommand(command string) {
	if ( len(self.waitingForCommand) != 0 ) {
		panic(fmt.Sprintf("Can't wait on multiple commands in the same environment : %s", self.waitingForCommand))
	}
	//fmt.Printf("Environment %p will wait for command %s\n", self, command)
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
	//fmt.Printf("Environment waiting for command %s? ", command)
	if (strings.Compare(self.waitingForCommand, command) == 0) {
		//fmt.Printf("YES\n")
		return true
	}
	if ( self.parent != nil ) {
		return self.parent.isWaitingForCommand(command)
	}
	//fmt.Printf("NO\n")
	return false
}

func (self *BasicEnvironment) stopWaiting(command string) {
	//fmt.Printf("Environment %p stopped waiting for command %s\n", self, command)
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

func (self *BasicEnvironment) getLabel(label string) (int64, error) {
	var ok bool
	var labelval int64
	var err error
	if labelval, ok = self.labels[label]; ok {
		return labelval, nil
	} else if ( self.parent != nil ) {
		labelval, err = self.parent.getLabel(label)
		if ( err != nil ) {
			return 0, err			
		}
		return labelval, nil
	}
	return 0, fmt.Errorf("Unable to find or create label %s in environment", label)
}

func (self *BasicEnvironment) setLabel(label string, value int64) error {
	// Only the toplevel environment creates labels
	if ( self.runtime.environment == self ) {
		self.labels[label] = value
		return nil
	} else if ( self.parent != nil ) {
		return self.parent.setLabel(label, value)
	}
	return errors.New("Unable to create label in orphaned environment")	
}

func (self *BasicEnvironment) get(varname string) *BasicVariable {
	var variable *BasicVariable
	var ok bool
 	sizes := []int64{1}
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
	// TODO : When the identifier has an argument list on .right, use it as
	// a subscript, flatten it to a pointer, and set the value there
	var variable *BasicVariable = nil
	var subscripts []int64
	var expr *BasicASTLeaf
	var tval *BasicValue
	var err error
	if ( lval == nil || rval == nil ) {
		return nil, errors.New("nil pointer")
	}
	variable = self.get(lval.identifier)
	// FIXME : Processing the sizes argumentlist before we validate the type of the
	// identifier leaf may lead to problems later.
	if ( lval.right != nil &&
		lval.right.leaftype == LEAF_ARGUMENTLIST &&
		lval.right.operator == ARRAY_SUBSCRIPT ) {
		expr = lval.right.right
		for ( expr != nil ) {
			tval, err = self.runtime.evaluate(expr)
			if ( err != nil ) {
				return nil, err
			}
			if ( tval.valuetype != TYPE_INTEGER ) {
				return nil, errors.New("Array dimensions must evaluate to integer (B)")
			}
			subscripts = append(subscripts, tval.intval)
			expr = expr.right
		}
	}
	if ( len(subscripts) == 0 ) {
		subscripts = append(subscripts, 0)
	}
	// FIXME : If we move this down below the switch() statement and return variable.getSusbcript(subscripts...) directly,
	// we get an arrat out of bounds error because somehow `subscripts` has been changed to an
	// array with a single entry [0] at this point. Getting a reference to the value here
	// prevents that.
	tval, err = variable.getSubscript(subscripts...)
	if ( err != nil ) {
		return nil, err
	}
	
	switch(lval.leaftype) {
	case LEAF_IDENTIFIER_INT:
		if ( rval.valuetype == TYPE_INTEGER ) {
			variable.setInteger(rval.intval, subscripts...)
		} else if ( rval.valuetype == TYPE_FLOAT ) {
			variable.setInteger(int64(rval.floatval), subscripts...)
		} else {
			return nil, errors.New("Incompatible types in variable assignment")
		}
	case LEAF_IDENTIFIER_FLOAT:
		if ( rval.valuetype == TYPE_INTEGER ) {
			variable.setFloat(float64(rval.intval), subscripts...)
		} else if ( rval.valuetype == TYPE_FLOAT ) {
			variable.setFloat(rval.floatval, subscripts...)
		} else {
			return nil, errors.New("Incompatible types in variable assignment")
		}
	case LEAF_IDENTIFIER_STRING:
		if ( rval.valuetype == TYPE_STRING ) {
			variable.setString(strings.Clone(rval.stringval), subscripts...)
		} else {
			return nil, errors.New("Incompatible types in variable assignment")
		}
	default:
		return nil, errors.New("Invalid assignment")		
	}
	variable.valuetype = rval.valuetype
	//fmt.Printf("Assigned %+v\n", variable)
	return tval, nil
}
