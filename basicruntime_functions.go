package main

import (
	"errors"
)

func (self *BasicRuntime) FunctionLEN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var strval *BasicValue = nil
	var varref *BasicVariable = nil
	var firstarg *BasicASTLeaf = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	firstarg = expr.firstArgument()
	
	if ( firstarg == nil ||
		firstarg == nil ||
		firstarg.isIdentifier() == false ) {
		//fmt.Printf("%+v\n", expr);
		//fmt.Printf("%+v\n", expr.right);
		return nil, errors.New("Expected identifier or string literal")
	}
	rval, err = self.newValue()
	if ( err != nil ) {
		return nil, err
	}	
	rval.valuetype = TYPE_INTEGER
	if ( firstarg.leaftype == LEAF_LITERAL_STRING ||
		firstarg.leaftype == LEAF_IDENTIFIER_STRING ) {
		strval, err = self.evaluate(firstarg)
		if ( err != nil ) {
			return nil, err
		}
		rval.intval = int64(len(strval.stringval))
	} else {
		varref = self.environment.get(firstarg.identifier)
		rval.intval = int64(len(varref.values))
	}
	return rval, nil
}

func (self *BasicRuntime) FunctionMID(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var strtarget *BasicValue = nil
	var startpos *BasicValue = nil
	var length *BasicValue = nil
	var curarg *BasicASTLeaf = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	curarg = expr.firstArgument()

	if ( curarg == nil ||
		( curarg.leaftype != LEAF_IDENTIFIER_STRING &&
			curarg.leaftype != LEAF_LITERAL_STRING )) {
		return nil, errors.New("Expected (STRING, INTEGER[, INTEGER])")
	}
	strtarget, err = self.evaluate(curarg)
	if ( err != nil ) {
		return nil, err
	}
	
	curarg = curarg.right
	if ( curarg == nil ||
		( curarg.leaftype != LEAF_IDENTIFIER_INT &&
			curarg.leaftype != LEAF_LITERAL_INT )) {
		return nil, errors.New("Expected (STRING, INTEGER[, INTEGER])")
	}
	startpos, err = self.evaluate(curarg)
	if ( err != nil ) {
		return nil, err
	}

	curarg = curarg.right
	if ( curarg != nil ) {
		// Optional length
		if ( curarg.leaftype != LEAF_IDENTIFIER_INT &&
			curarg.leaftype != LEAF_LITERAL_INT ) {
			return nil, errors.New("Expected (STRING, INTEGER[, INTEGER])")
		}
		length, err = self.evaluate(curarg)
		if ( err != nil ) {
			return nil, err
		}
	} else {
		length, err = self.newValue()
		if ( err != nil ) {
			return nil, err
		}
		length.intval = int64(len(strtarget.stringval))
	}

	rval, err = self.newValue()
	if ( err != nil ) {
		return nil, err
	}
	rval.stringval = strtarget.stringval[startpos.intval:(startpos.intval+length.intval)]
	rval.valuetype = TYPE_STRING
	return rval, nil
}
