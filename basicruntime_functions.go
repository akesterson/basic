package main

import (
	"errors"
)

func (self *BasicRuntime) FunctionLEN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var strval *BasicValue = nil
	var varref *BasicVariable = nil
	
	if ( expr.right == nil ||
		( expr.right.leaftype != LEAF_IDENTIFIER_STRING &&
			expr.right.leaftype != LEAF_IDENTIFIER_INT &&
			expr.right.leaftype != LEAF_IDENTIFIER_FLOAT &&
			expr.right.leaftype != LEAF_LITERAL_STRING )) {
		//fmt.Printf("%+v\n", expr);
		//fmt.Printf("%+v\n", expr.right);
		return nil, errors.New("Expected identifier or string literal")
	}
	rval, err = self.newValue()
	if ( err != nil ) {
		return nil, err
	}	
	rval.valuetype = TYPE_INTEGER
	if ( expr.right.leaftype == LEAF_LITERAL_STRING ) {
		strval, err = self.evaluate(expr.right)
		if ( err != nil ) {
			return nil, err
		}
		rval.intval = int64(len(strval.stringval))
	} else {
		varref = self.environment.get(expr.right.identifier)
		rval.intval = int64(len(varref.values))
	}
	return rval, nil
}

func (self *BasicRuntime) FunctionMID(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var strtarget *BasicValue = nil
	var startpos *BasicValue = nil
	var length *BasicValue = nil

	expr = expr.right
	if ( expr == nil ||
		( expr.leaftype != LEAF_IDENTIFIER_STRING &&
			expr.leaftype != LEAF_LITERAL_STRING )) {
		return nil, errors.New("Expected (STRING, INTEGER[, INTEGER])")
	}
	strtarget, err = self.evaluate(expr)
	if ( err != nil ) {
		return nil, err
	}
	
	expr = expr.right
	if ( expr == nil ||
		( expr.leaftype != LEAF_IDENTIFIER_INT &&
			expr.leaftype != LEAF_LITERAL_INT )) {
		return nil, errors.New("Expected (STRING, INTEGER[, INTEGER])")
	}
	startpos, err = self.evaluate(expr)
	if ( err != nil ) {
		return nil, err
	}

	expr = expr.right
	if ( expr != nil ) {
		// Optional length
		if ( expr.leaftype != LEAF_IDENTIFIER_INT &&
			expr.leaftype != LEAF_LITERAL_INT ) {
			return nil, errors.New("Expected (STRING, INTEGER[, INTEGER])")
		}
		length, err = self.evaluate(expr)
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
	rval.stringval = strtarget.stringval[startpos.intval:length.intval]
	rval.valuetype = TYPE_STRING
	return rval, nil
}
