package main

import (
	"fmt"
	"errors"
	"strings"
)

type BasicError int
const (
	IO           BasicError = iota
	PARSE
	SYNTAX
	EXECUTE	
)

type BasicType int
const (
	TYPE_UNDEFINED    BasicType = iota
	TYPE_INTEGER  
	TYPE_FLOAT
	TYPE_STRING
	TYPE_BOOLEAN
)

type BasicValue struct {
	valuetype BasicType
	stringval string
	intval int64
	floatval float64
	boolvalue int64
}

func (self *BasicValue) init() {
	self.valuetype = TYPE_UNDEFINED
	self.stringval = ""
	self.intval = 0
	self.floatval = 0.0
	self.boolvalue = BASIC_FALSE
}

func (self *BasicValue) toString() string {
	switch ( self.valuetype ) {
	case TYPE_STRING: return self.stringval
	case TYPE_INTEGER: return fmt.Sprintf("%d", self.intval)
	case TYPE_FLOAT: return fmt.Sprintf("%f", self.floatval)
	case TYPE_BOOLEAN: return fmt.Sprintf("%t", (self.boolvalue == BASIC_TRUE))
	}
	return "(UNDEFINED STRING REPRESENTATION)"
}

func (self *BasicValue) invert() error {
	if ( self.valuetype == TYPE_STRING ) {
		return errors.New("Cannot invert a string")
	}
	self.intval = -(self.intval)
	self.floatval = -(self.floatval)
	return nil
}

func (self *BasicValue) bitwiseNot() error {
	if ( self.valuetype != TYPE_INTEGER ) {
		return errors.New("Cannot only perform bitwise operations on integers")
	}
	self.intval = ^self.intval
	return nil
}

func (self *BasicValue) bitwiseAnd(rval *BasicValue) error {
	if ( rval == nil ) {
		return errors.New("nil rval")
	}
	if ( self.valuetype != TYPE_INTEGER ) {
		return errors.New("Cannot perform bitwise operations on string or float")
	}
	self.intval &= rval.intval
	return nil
}

func (self *BasicValue) bitwiseOr(rval *BasicValue) error {
	if ( rval == nil ) {
		return errors.New("nil rval")
	}
	if ( self.valuetype != TYPE_INTEGER ) {
		return errors.New("Cannot only perform bitwise operations on integers")
	}
	self.intval |= rval.intval
	return nil
}

// TODO: Implement - (remove) * (duplicate) and / (split) on string types, that would be cool

func (self *BasicValue) mathPlus(rval *BasicValue) error {
	if ( rval == nil ) {
		return errors.New("nil rval")
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		self.intval += (rval.intval + int64(rval.floatval))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		self.floatval += (rval.floatval + float64(rval.intval))
	} else if ( self.valuetype == TYPE_STRING && rval.valuetype == TYPE_STRING ){
		self.stringval += rval.stringval
	} else {
		return errors.New("Invalid arithmetic operation")
	}
	return nil
}


func (self *BasicValue) mathMinus(rval *BasicValue) error {
	if ( rval == nil ) {
		return errors.New("nil rval")
	}
	if ( self.valuetype == TYPE_STRING || rval.valuetype == TYPE_STRING ) {
		return errors.New("Cannot perform subtraction on strings")
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		self.intval -= (rval.intval + int64(rval.floatval))
	} else {
		self.floatval -= (rval.floatval + float64(rval.intval))
	}
	return nil
}

func (self *BasicValue) mathDivide(rval *BasicValue) error {
	if ( rval == nil ) {
		return errors.New("nil rval")
	}
	if ( self.valuetype == TYPE_STRING || rval.valuetype == TYPE_STRING ) {
		return errors.New("Cannot perform division on strings")
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		self.intval = self.intval / (rval.intval + int64(rval.floatval))
	} else {
		self.floatval = self.floatval / (rval.floatval + float64(rval.intval))
	}
	return nil
}

func (self *BasicValue) mathMultiply(rval *BasicValue) error {
	if ( rval == nil ) {
		return errors.New("nil rval")
	}
	if ( self.valuetype == TYPE_STRING || rval.valuetype == TYPE_STRING ) {
		return errors.New("Cannot perform multiplication on strings")
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		self.intval = self.intval * (rval.intval + int64(rval.floatval))
	} else {
		self.floatval = self.floatval * (rval.floatval + float64(rval.intval))
	}
	return nil
}

func (self *BasicValue) lessThan(rval *BasicValue) error {
	if ( rval == nil ) {
		return errors.New("nil rval")
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		self.basicBoolValue(self.intval < (rval.intval + int64(rval.floatval)))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		self.basicBoolValue(self.floatval < (rval.floatval + float64(rval.intval)))
	} else {
		self.basicBoolValue(strings.Compare(self.stringval, rval.stringval) < 0)
	}
	return nil
}

func (self *BasicValue) lessThanEqual(rval *BasicValue) error {
	var result int
	if ( rval == nil ) {
		return errors.New("nil rval")
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		self.basicBoolValue(self.intval <= (rval.intval + int64(rval.floatval)))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		self.basicBoolValue(self.floatval <= (rval.floatval + float64(rval.intval)))
	} else {
		result = strings.Compare(self.stringval, rval.stringval)
		self.basicBoolValue(result < 0 || result == 0)
	}
	return nil
}

func (self *BasicValue) greaterThan(rval *BasicValue) error {
	var result int
	if ( rval == nil ) {
		return errors.New("nil rval")
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		self.basicBoolValue(self.intval > (rval.intval + int64(rval.floatval)))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		self.basicBoolValue(self.floatval > (rval.floatval + float64(rval.intval)))
	} else {
		result = strings.Compare(self.stringval, rval.stringval)
		self.basicBoolValue((result > 0))
	}
	return nil
}

func (self *BasicValue) greaterThanEqual(rval *BasicValue) error {
	var result int
	if ( rval == nil ) {
		return errors.New("nil rval")
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		self.basicBoolValue(self.intval >= (rval.intval + int64(rval.floatval)))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		self.basicBoolValue(self.floatval >= (rval.floatval + float64(rval.intval)))
	} else {
		result = strings.Compare(self.stringval, rval.stringval)
		self.basicBoolValue(result > 0 || result == 0)
	}
	return nil
}

func (self *BasicValue) isEqual(rval *BasicValue) error {
	var result int
	if ( rval == nil ) {
		return errors.New("nil rval")
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		self.basicBoolValue(self.intval == (rval.intval + int64(rval.floatval)))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		self.basicBoolValue(self.floatval == (rval.floatval + float64(rval.intval)))
	} else {
		result = strings.Compare(self.stringval, rval.stringval)
		self.basicBoolValue(result == 0)
	}
	return nil
}

func (self *BasicValue) isNotEqual(rval *BasicValue) error {
	var result int
	if ( rval == nil ) {
		return errors.New("nil rval")
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		self.basicBoolValue(self.intval != (rval.intval + int64(rval.floatval)))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		self.basicBoolValue(self.floatval != (rval.floatval + float64(rval.intval)))
	} else {
		result = strings.Compare(self.stringval, rval.stringval)
		self.basicBoolValue(result != 0)
	}
	return nil
}

func (self *BasicValue) basicBoolValue(result bool) {
	self.valuetype = TYPE_BOOLEAN
	if ( result == true ) {
		self.boolvalue = BASIC_TRUE
		return
	}
	self.boolvalue = BASIC_FALSE
}

type BasicRuntime struct {
	source [9999]string
	lineno int
	values [MAX_VALUES]BasicValue
	nextvalue int
}

func (self BasicRuntime) init() {
	self.lineno = 0
	self.nextvalue = 0
}

func (self BasicRuntime) newValue() (*BasicValue, error) {
	var value *BasicValue
	if ( self.nextvalue < MAX_VALUES ) {
		value = &self.values[self.nextvalue]
		self.nextvalue += 1
		return value, nil
	}
	return nil, errors.New("Maximum values per line reached")
}

func (self BasicRuntime) isTrue(value *BasicValue) (bool, error) {
	if ( value.valuetype == TYPE_STRING ) {
		return false, errors.New("strings cannot evaluate to true (-1) or false (0)")
	}
	if ( value.intval == BASIC_TRUE || value.floatval == BASIC_TRUE ) {
		return true, nil
	}
	return false, nil
}

func (self BasicRuntime) evaluate(expr *BasicASTLeaf) (*BasicValue, error) {
	var lval *BasicValue
	var rval *BasicValue
	var err error = nil
	
	lval, err = self.newValue()
	if ( err != nil ) {
		return nil, err
	}
	lval.init()
	
	switch (expr.leaftype) {
	case LEAF_GROUPING: return self.evaluate(expr.expr)
	case LEAF_LITERAL_INT:
		lval.valuetype = TYPE_INTEGER
		lval.intval = expr.literal_int
	case LEAF_LITERAL_FLOAT:
		lval.valuetype = TYPE_FLOAT
		lval.floatval = expr.literal_float
	case LEAF_LITERAL_STRING:
		lval.valuetype = TYPE_STRING
		lval.stringval = expr.literal_string
	case LEAF_UNARY:
		rval, err = self.evaluate(expr.right)
		if ( err != nil ) {
			return nil, err
		}
		switch (expr.operator) {
		case MINUS:
			err = rval.invert()
			if ( err != nil ) {
				return nil, err
			}
			return rval, nil
		case NOT:
			err = rval.bitwiseNot()
			if ( err != nil ) {
				return nil, err
			}
			return rval, nil
		default:
			return nil, errors.New(fmt.Sprintf("Don't know how to perform operation %d on unary type %d", expr.operator, rval.valuetype))
		}
	case LEAF_BINARY:
		lval, err = self.evaluate(expr.left)
		if ( err != nil ) {
			return nil, err
		}
		rval, err = self.evaluate(expr.right)
		if ( err != nil ) {
			return nil, err
		}
		switch (expr.operator) {
		case MINUS:
			err = lval.mathMinus(rval)
		case PLUS:
			err = lval.mathPlus(rval)
		case LEFT_SLASH:
			err = lval.mathDivide(rval)
		case STAR:
			err = lval.mathMultiply(rval)
		case AND:
			err = lval.bitwiseAnd(rval)
		case OR:
			err = lval.bitwiseOr(rval)
		case LESS_THAN:
			err = lval.lessThan(rval)
		case LESS_THAN_EQUAL:
			err = lval.lessThanEqual(rval)
		case EQUAL:
			err = lval.isEqual(rval)
		case NOT_EQUAL:
			err = lval.isNotEqual(rval)
		case GREATER_THAN:
			err = lval.greaterThan(rval)
		case GREATER_THAN_EQUAL:
			err = lval.greaterThanEqual(rval)
		}
		if ( err != nil ) {
			return nil, err
		}
	}
	return lval, nil
}

func (self *BasicRuntime) interpret(expr *BasicASTLeaf) {
	var value *BasicValue
	var err error
	value, err = self.evaluate(expr)
	if ( err != nil ) {
		fmt.Println(err)
		return
	}
	fmt.Println(value.toString())
}
