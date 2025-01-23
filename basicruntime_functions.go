package main

import (
	"errors"
	"strings"
)

func (self *BasicRuntime) CommandDEF(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	if ( expr == nil ||
		expr.left == nil ||
		expr.right == nil ||
		expr.expr == nil) {
		return nil, errors.New("Incomplete function definition")
	}
	//fmt.Printf("DEF leaf : %s\n", expr.toString())
	//fmt.Printf("DEF Name leaf : %s\n", expr.right.toString())
	//fmt.Printf("DEF Arglist leaf : %s (%+v)\n", expr.left.toString(), expr.left)
	//fmt.Printf("DEF Expression leaf : %s\n", expr.expr.toString())
	self.environment.functions[expr.right.identifier] = &BasicFunctionDef{
		arglist: expr.left.clone(),
		expression: expr.expr.clone(),
		runtime: self,
		name: strings.Clone(expr.right.identifier)}
	//fmt.Printf("Defined function %+v\n", self.environment.functions[expr.right.identifier])
	return nil, nil
}

func (self *BasicRuntime) CommandLEN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var strval *BasicValue = nil
	
	if ( expr.right == nil ||
		( expr.right.leaftype != LEAF_IDENTIFIER_STRING &&
			expr.right.leaftype != LEAF_LITERAL_STRING )) {
		//fmt.Printf("%+v\n", expr);
		//fmt.Printf("%+v\n", expr.right);
		return nil, errors.New("Expected identifier or string literal")
	}
	strval, err = self.evaluate(expr.right)
	if ( err != nil ) {
		return nil, err
	}
	rval, err = self.newValue()
	if ( err != nil ) {
		return nil, err
	}
	rval.intval = int64(len(strval.stringval))
	rval.valuetype = TYPE_INTEGER
	return rval, nil
}

func (self *BasicRuntime) CommandMID(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
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
