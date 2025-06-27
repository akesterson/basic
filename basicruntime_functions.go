package main

import (
	"errors"
	"math"
	"fmt"
	//"bufio"
	"strings"
)

func (self *BasicRuntime) initFunctions() {
	var funcdefs string = `
10 DEF ABS(X#) = X#
20 DEF ATN(X#) = X#
30 DEF CHR(X#) = X#
40 DEF COS(X#) = X#
50 DEF HEX(X#) = X#
60 DEF INSTR(X$, Y$) = X$
70 DEF LEFT(X$, A#) = X$
80 DEF LEN(X$) = X$
90 DEF LOG(X#) = X#
100 DEF MID(A$, S$, L#) = A$
110 DEF RIGHT(X$, A#) = X$
120 DEF RAD(X#) = X#`
	var oldmode int = self.mode
	self.run(strings.NewReader(funcdefs), MODE_RUNSTREAM)
	for _, basicfunc := range self.environment.functions {
		basicfunc.expression = nil
		self.scanner.commands[basicfunc.name] = FUNCTION
		delete(self.scanner.functions, basicfunc.name)
		//fmt.Printf("%+v\n", basicfunc)
	}
	for i, _ := range self.source {
		self.source[i].code = ""
		self.source[i].lineno = 0
	}
	self.setMode(oldmode)
}

func (self *BasicRuntime) FunctionABS(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var tval *BasicValue = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	expr = expr.firstArgument()
	if (expr != nil) {
		rval, err = self.evaluate(expr)
		if ( err != nil ) {
			return nil, err
		}
		if ( rval.valuetype != TYPE_INTEGER &&
			rval.valuetype != TYPE_FLOAT ) {
			return nil, errors.New("ABS expected INTEGER or FLOAT")
		}
		tval, err = rval.clone(tval)
		if ( tval == nil ) {
			return nil, err
		}
		tval.intval = int64(math.Abs(float64(tval.intval)))
		tval.floatval = math.Abs(tval.floatval)
		return tval, nil
	}
	return nil, errors.New("ABS expected integer or float")
}

func (self *BasicRuntime) FunctionATN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var tval *BasicValue = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	expr = expr.firstArgument()
	if (expr != nil) {
		rval, err = self.evaluate(expr)
		if ( err != nil ) {
			return nil, err
		}
		tval, err = self.newValue()
		if ( tval == nil ) {
			return nil, err
		}
		tval.valuetype = TYPE_FLOAT
		if ( rval.valuetype == TYPE_INTEGER ) {
			tval.floatval = math.Atan(float64(rval.intval))
		} else if ( rval.valuetype == TYPE_FLOAT ) {
			tval.floatval = math.Atan(rval.floatval)
		} else {
			return nil, errors.New("ATN expected INTEGER or FLOAT")
		}
		return tval, nil
	}
	return nil, errors.New("ATN expected integer or float")
}

func (self *BasicRuntime) FunctionCHR(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var tval *BasicValue = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	expr = expr.firstArgument()
	if (expr != nil) {
		rval, err = self.evaluate(expr)
		if ( err != nil ) {
			return nil, err
		}
		if ( rval.valuetype != TYPE_INTEGER ) {
			return nil, errors.New("CHR expected INTEGER")
		}
		tval, err = self.newValue()
		if ( tval == nil ) {
			return nil, err
		}
		tval.valuetype = TYPE_STRING
		tval.stringval = string(rune(rval.intval)) 
		return tval, nil
	}
	return nil, errors.New("CHR expected INTEGER")
}

func (self *BasicRuntime) FunctionCOS(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var tval *BasicValue = nil
	
	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	expr = expr.firstArgument()
	if (expr != nil) {
		rval, err = self.evaluate(expr)
		if ( err != nil ) {
			return nil, err
		}
		tval, err = self.newValue()
		if ( tval == nil ) {
			return nil, err
		}
		tval.valuetype = TYPE_FLOAT
		if ( rval.valuetype == TYPE_INTEGER ) {
			tval.floatval = math.Cos(float64(rval.intval))
		} else if ( rval.valuetype == TYPE_FLOAT ) {
			tval.floatval = math.Cos(rval.floatval)
		} else {
			return nil, errors.New("COS expected INTEGER or FLOAT")
		}
		return tval, nil
	}
	return nil, errors.New("COS expected integer or float")
}

func (self *BasicRuntime) FunctionHEX(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var tval *BasicValue = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	expr = expr.firstArgument()
	if (expr != nil) {
		rval, err = self.evaluate(expr)
		if ( err != nil ) {
			return nil, err
		}
		if ( rval.valuetype != TYPE_INTEGER ) {
			return nil, errors.New("CHR expected INTEGER")
		}
		tval, err = self.newValue()
		if ( tval == nil ) {
			return nil, err
		}
		tval.valuetype = TYPE_STRING
		tval.stringval = fmt.Sprintf("%x", rval.intval) 
		return tval, nil
	}
	return nil, errors.New("CHR expected INTEGER")
}

func (self *BasicRuntime) FunctionINSTR(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var strtarget *BasicValue = nil
	var substr *BasicValue = nil
	var curarg *BasicASTLeaf = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	curarg = expr.firstArgument()

	if ( curarg == nil ||
		( curarg.leaftype != LEAF_IDENTIFIER_STRING &&
			curarg.leaftype != LEAF_LITERAL_STRING )) {
		return nil, errors.New("Expected (STRING, STRING)")
	}
	strtarget, err = self.evaluate(curarg)
	if ( err != nil ) {
		return nil, err
	}
	
	curarg = curarg.right
	if ( curarg == nil ||
		( curarg.leaftype != LEAF_IDENTIFIER_STRING &&
			curarg.leaftype != LEAF_LITERAL_STRING )) {
		return nil, errors.New("Expected (STRING, STRING)")
	}
	substr, err = self.evaluate(curarg)
	if ( err != nil ) {
		return nil, err
	}
	rval, err = self.newValue()
	if ( err != nil ) {
		return nil, err
	}
	rval.intval = int64(strings.Index(strtarget.stringval, substr.stringval))
	rval.valuetype = TYPE_INTEGER
	return rval, nil
}

func (self *BasicRuntime) FunctionLEFT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var strtarget *BasicValue = nil
	var length *BasicValue = nil
	var curarg *BasicASTLeaf = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	curarg = expr.firstArgument()

	if ( curarg == nil ||
		( curarg.leaftype != LEAF_IDENTIFIER_STRING &&
			curarg.leaftype != LEAF_LITERAL_STRING )) {
		return nil, errors.New("Expected (STRING, INTEGER)")
	}
	strtarget, err = self.evaluate(curarg)
	if ( err != nil ) {
		return nil, err
	}
	
	curarg = curarg.right
	if ( curarg == nil ||
		( curarg.leaftype != LEAF_IDENTIFIER_INT &&
			curarg.leaftype != LEAF_LITERAL_INT )) {
		return nil, errors.New("Expected (STRING, INTEGER)")
	}
	length, err = self.evaluate(curarg)
	if ( err != nil ) {
		return nil, err
	}
	rval, err = self.newValue()
	if ( err != nil ) {
		return nil, err
	}
	if ( length.intval >= int64(len(strtarget.stringval)) ) {
		rval.stringval = strings.Clone(strtarget.stringval)
	} else {
		rval.stringval = strtarget.stringval[0:length.intval]
	}
	rval.valuetype = TYPE_STRING
	return rval, nil
}

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

func (self *BasicRuntime) FunctionLOG(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var tval *BasicValue = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	expr = expr.firstArgument()
	if (expr != nil) {
		rval, err = self.evaluate(expr)
		if ( err != nil ) {
			return nil, err
		}
		if ( rval.valuetype != TYPE_INTEGER &&
			rval.valuetype != TYPE_FLOAT ) {
			return nil, errors.New("LOG expected INTEGER or FLOAT")
		}
		tval, err = rval.clone(tval)
		if ( tval == nil ) {
			return nil, err
		}
		tval.intval = int64(math.Log(float64(tval.intval)))
		tval.floatval = math.Log(tval.floatval)
		return tval, nil
	}
	return nil, errors.New("LOG expected integer or float")
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

func (self *BasicRuntime) FunctionRAD(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var tval *BasicValue = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	expr = expr.firstArgument()
	if (expr != nil) {
		rval, err = self.evaluate(expr)
		if ( err != nil ) {
			return nil, err
		}
		tval, err = self.newValue()
		if ( tval == nil ) {
			return nil, err
		}
		tval.valuetype = TYPE_FLOAT
		if ( rval.valuetype == TYPE_INTEGER ) {
			tval.floatval = float64(rval.intval) * (math.Pi / 180)
		} else if ( rval.valuetype == TYPE_FLOAT ) {
			tval.floatval = rval.floatval * (math.Pi / 180)
		} else {
			return nil, errors.New("RAD expected INTEGER or FLOAT")
		}
		return tval, nil
	}
	return nil, errors.New("RAD expected integer or float")
}

func (self *BasicRuntime) FunctionRIGHT(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var strtarget *BasicValue = nil
	var length *BasicValue = nil
	var curarg *BasicASTLeaf = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	curarg = expr.firstArgument()

	if ( curarg == nil ||
		( curarg.leaftype != LEAF_IDENTIFIER_STRING &&
			curarg.leaftype != LEAF_LITERAL_STRING )) {
		return nil, errors.New("Expected (STRING, INTEGER)")
	}
	strtarget, err = self.evaluate(curarg)
	if ( err != nil ) {
		return nil, err
	}
	
	curarg = curarg.right
	if ( curarg == nil ||
		( curarg.leaftype != LEAF_IDENTIFIER_INT &&
			curarg.leaftype != LEAF_LITERAL_INT )) {
		return nil, errors.New("Expected (STRING, INTEGER)")
	}
	length, err = self.evaluate(curarg)
	if ( err != nil ) {
		return nil, err
	}
	rval, err = self.newValue()
	if ( err != nil ) {
		return nil, err
	}
	var maxlen = int64(len(strtarget.stringval))
	if ( length.intval >= maxlen ) {
		rval.stringval = strings.Clone(strtarget.stringval)
	} else {
		var start int64 = maxlen - length.intval
		rval.stringval = strtarget.stringval[start:maxlen]
	}
	rval.valuetype = TYPE_STRING
	return rval, nil
}
