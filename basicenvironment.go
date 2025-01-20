package main

import (
	"fmt"
	"errors"
	"strings"
)

type BasicEnvironment struct {
	variables map[string]*BasicValue
	runtime *BasicRuntime
}

func (self *BasicEnvironment) init(runtime *BasicRuntime) {
	self.variables = make(map[string]*BasicValue)
	self.runtime = runtime
}

func (self *BasicEnvironment) get(varname string) *BasicValue {
	var value *BasicValue
	var ok bool
	if value, ok = self.variables[varname]; ok {
		return value
	}
	self.variables[varname] = &BasicValue{
		name: strings.Clone(varname),
		valuetype: TYPE_UNDEFINED,
		stringval: "",
		intval: 0,
		floatval: 0.0,
		boolvalue: BASIC_FALSE,
		runtime: self.runtime}
	return self.variables[varname]
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
	fmt.Printf("Assigned variable %s\n", variable.name)
	return variable, nil
}
