package main

import (
	"errors"
	"math"
	"fmt"
	//"bufio"
	"strings"
	"strconv"
	"slices"
	"unsafe"
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
101 DEF MOD(X%, Y%) = X% - (Y% * (X% / Y%))
104 DEF PEEK(X#) = X#
105 DEF POINTERVAR(X#) = X#
106 DEF POINTER(X#) = X#
110 DEF RIGHT(X$, A#) = X$
120 DEF RAD(X#) = X#
130 DEF SGN(X#) = X#
135 DEF SHL(X#, Y#) = X#
136 DEF SHR(X#, Y#) = X#
140 DEF SIN(X#) = X#
150 DEF SPC(X#) = " " * X#
160 DEF STR(X#) = "" + X#
170 DEF TAN(X#) = X#
180 DEF VAL(X$) = X#
190 DEF XOR(X#, Y#) = X#`
	var freeStandingFunctions = []string{
		"MOD",
		"SPC",
		"STR"}
	var oldmode int = self.mode
	self.run(strings.NewReader(funcdefs), MODE_RUNSTREAM)
	for _, basicfunc := range self.environment.functions {
		if ( slices.Contains(freeStandingFunctions, basicfunc.name) == false ) {
			basicfunc.expression = nil
		}
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
		(firstarg.isIdentifier() == false &&
			firstarg.isLiteral() == false)) {
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

func (self *BasicRuntime) FunctionPEEK(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var tval *BasicValue = nil
	var addr uintptr
	var ptr unsafe.Pointer
	var typedPtr *byte
	
	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	expr = expr.firstArgument()
	if (expr != nil) {
		if ( expr.leaftype != LEAF_LITERAL_INT &&
			expr.leaftype != LEAF_IDENTIFIER_INT) {
			return nil, errors.New("PEEK expected INTEGER or INTEGER VARIABLE")
		}
		rval, err = self.evaluate(expr)
		if ( err != nil ) {
			return nil, err
		}
		tval, err = self.newValue()
		if ( err != nil ) {
			return nil, err
		}
		if ( rval.valuetype != TYPE_INTEGER || rval.intval == 0 ) {
			return nil, errors.New("PEEK got NIL pointer or uninitialized variable")
		}
		addr = uintptr(rval.intval)
		ptr = unsafe.Pointer(addr)
		typedPtr = (*byte)(ptr)
		tval.valuetype = TYPE_INTEGER
		tval.intval = int64(*typedPtr)
		return tval, nil
	}
	return nil, errors.New("PEEK expected integer or float")
}

func (self *BasicRuntime) FunctionPOINTERVAR(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var tvar *BasicVariable = nil
	var tval *BasicValue = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	expr = expr.firstArgument()
	if (expr != nil) {
		if ( expr.isIdentifier() == false ) {
			return nil, errors.New("POINTERVAR expected IDENTIFIER")
		}
		tvar = self.environment.get(expr.identifier)
		tval, err = self.newValue()
		if ( err != nil ) {
			return nil, err
		}
		tval.valuetype = TYPE_INTEGER
		tval.intval = int64(uintptr(unsafe.Pointer(tvar)))
		return tval, nil
	}
	return nil, errors.New("POINTERVAR expected integer or float")
}

func (self *BasicRuntime) FunctionPOINTER(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var tval *BasicValue = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	expr = expr.firstArgument()
	if (expr != nil) {
		if ( expr.isIdentifier() == false ) {
			return nil, errors.New("POINTER expected IDENTIFIER")
		}
		rval, err = self.evaluate(expr)
		if ( err != nil ) {
			return nil, err
		}
		tval, err = self.newValue()
		if ( err != nil ) {
			return nil, err
		}
		tval.valuetype = TYPE_INTEGER
		switch (rval.valuetype) {
		case TYPE_INTEGER:
			tval.intval = int64(uintptr(unsafe.Pointer(&rval.intval)))
		case TYPE_FLOAT:
			tval.intval = int64(uintptr(unsafe.Pointer(&rval.floatval)))
		case TYPE_STRING:
			tval.intval = int64(uintptr(unsafe.Pointer(&rval.stringval)))
		default:
			return nil, errors.New("POINTER expects a INT, FLOAT or STRING variable")
		}
		return tval, nil
	}
	return nil, errors.New("POINTER expected integer or float")
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

func (self *BasicRuntime) FunctionSGN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
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
			return nil, errors.New("SGN expected INTEGER or FLOAT")
		}
		tval, err = self.newValue()
		if ( tval == nil ) {
			return nil, err
		}
		tval.zero()
		tval.valuetype = TYPE_INTEGER
		if ( rval.intval < 0 || rval.floatval < 0 ) {
			tval.intval = -1
		} else if ( rval.intval > 0 || rval.floatval > 0 ) {
			tval.intval = 1
		} else {
			tval.intval = 0
		}
		return tval, nil
	}
	return nil, errors.New("ABS expected integer or float")
}

func (self *BasicRuntime) FunctionSHL(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var sval *BasicValue = nil

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
			return nil, errors.New("SHL expected NUMERIC, INTEGER")
		}
		sval, err = self.evaluate(expr.right)
		if ( err != nil ) {
			return nil, err
		}
		if ( rval.valuetype != TYPE_INTEGER) {
			return nil, errors.New("SHL expected NUMERIC, INTEGER")
		}
		return rval.bitwiseShiftLeft(sval.intval)
	}
	return nil, errors.New("SHL expected NUMERIC, NUMERIC")
}

func (self *BasicRuntime) FunctionSHR(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var sval *BasicValue = nil

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
			return nil, errors.New("SHR expected NUMERIC, INTEGER")
		}
		sval, err = self.evaluate(expr.right)
		if ( err != nil ) {
			return nil, err
		}
		if ( rval.valuetype != TYPE_INTEGER) {
			return nil, errors.New("SHR expected NUMERIC, INTEGER")
		}
		return rval.bitwiseShiftRight(sval.intval)
	}
	return nil, errors.New("SHR expected NUMERIC, NUMERIC")
}

func (self *BasicRuntime) FunctionSIN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
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
			tval.floatval = math.Sin(float64(rval.intval))
		} else if ( rval.valuetype == TYPE_FLOAT ) {
			tval.floatval = math.Sin(rval.floatval)
		} else {
			return nil, errors.New("SIN expected INTEGER or FLOAT")
		}
		return tval, nil
	}
	return nil, errors.New("SIN expected integer or float")
}

func (self *BasicRuntime) FunctionTAN(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
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
			tval.floatval = math.Tan(float64(rval.intval))
		} else if ( rval.valuetype == TYPE_FLOAT ) {
			tval.floatval = math.Tan(rval.floatval)
		} else {
			return nil, errors.New("TAN expected INTEGER or FLOAT")
		}
		return tval, nil
	}
	return nil, errors.New("TAN expected integer or float")
}

func (self *BasicRuntime) FunctionVAL(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	var strval *BasicValue = nil
	var firstarg *BasicASTLeaf = nil

	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	firstarg = expr.firstArgument()
	
	if ( firstarg == nil ||
		firstarg == nil ||
		(firstarg.isIdentifier() == false &&
			firstarg.isLiteral() == false)) {
		//fmt.Printf("%+v\n", expr);
		//fmt.Printf("%+v\n", expr.right);
		return nil, errors.New("Expected identifier or string literal")
	}
	rval, err = self.newValue()
	if ( err != nil ) {
		return nil, err
	}	
	rval.valuetype = TYPE_FLOAT
	if ( firstarg.leaftype == LEAF_LITERAL_STRING ||
		firstarg.leaftype == LEAF_IDENTIFIER_STRING ) {
		strval, err = self.evaluate(firstarg)
		if ( err != nil ) {
			return nil, err
		}
		rval.floatval, err = strconv.ParseFloat(strval.stringval, 64)
		if ( err != nil ) {
			return nil, err
		}
	} else {
		return nil, errors.New("Expected identifier or string literal")
	}
	return rval, nil
}

func (self *BasicRuntime) FunctionXOR(expr *BasicASTLeaf, lval *BasicValue, rval *BasicValue) (*BasicValue, error) {
	var err error = nil
	
	if ( expr == nil ) {
		return nil, errors.New("NIL leaf")
	}
	expr = expr.firstArgument()
	if (expr != nil) {
		lval, err = self.evaluate(expr)
		if ( err != nil ) {
			return nil, err
		}
		rval, err = self.evaluate(expr.right)
		if ( err != nil ) {
			return nil, err
		}
		return lval.bitwiseXor(rval)
	}
	return nil, errors.New("COS expected integer or float")
}
