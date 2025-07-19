package main

import (
	"fmt"
	"errors"
	"strings"
)

type BasicType int
const (
	TYPE_UNDEFINED    BasicType = iota
	TYPE_INTEGER      // 1
	TYPE_FLOAT        // 2
	TYPE_STRING       // 3
	TYPE_BOOLEAN      // 4
)

type BasicValue struct {
	name string
	valuetype BasicType
	stringval string
	intval int64
	floatval float64
	boolvalue int64
	runtime *BasicRuntime
	mutable bool
}

func (self *BasicValue) init() {
}

func (self *BasicValue) zero() {
	self.valuetype = TYPE_UNDEFINED
	self.stringval = ""
	self.mutable = false
	self.name = ""
	self.intval = 0
	self.floatval = 0.0
	self.boolvalue = BASIC_FALSE
}

func (self *BasicValue) clone(dest *BasicValue) (*BasicValue, error) {
	var err error
	if ( dest == nil ) {
		dest, err = self.runtime.newValue()
		if ( err != nil ) {
			return nil, err
		}
	}
	dest.name = strings.Clone(self.name)
	dest.runtime = self.runtime
	dest.valuetype = self.valuetype
	dest.stringval = strings.Clone(self.stringval)
	dest.intval = self.intval
	dest.floatval = self.floatval
	dest.boolvalue = self.boolvalue
	return dest, nil
}

func (self *BasicValue) toString() string {
	switch ( self.valuetype ) {
	case TYPE_STRING: return self.stringval
	case TYPE_INTEGER: return fmt.Sprintf("%d", self.intval)
	case TYPE_FLOAT: return fmt.Sprintf("%f", self.floatval)
	case TYPE_BOOLEAN: return fmt.Sprintf("%t", (self.boolvalue == BASIC_TRUE))
	}
	return fmt.Sprintf("(UNDEFINED STRING REPRESENTATION FOR %d)", self.valuetype)
}


func (self *BasicValue) cloneIfNotMutable() (*BasicValue, error) {
	if ( !self.mutable ) {
		return self.clone(nil)
	}
	return self, nil
}


func (self *BasicValue) invert() (*BasicValue, error) {
	if ( self.valuetype == TYPE_STRING ) {
		return nil, errors.New("Cannot invert a string")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	dest.intval = -(self.intval)
	dest.floatval = -(self.floatval)
	return dest, nil
}

func (self *BasicValue) bitwiseShiftLeft(bits int64) (*BasicValue, error) {
	if ( self.valuetype != TYPE_INTEGER ) {
		return nil, errors.New("Only integer datatypes can be bit-shifted")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	dest.intval = dest.intval << bits
	return dest, nil
}

func (self *BasicValue) bitwiseShiftRight(bits int64) (*BasicValue, error) {
	if ( self.valuetype != TYPE_INTEGER) {
		return nil, errors.New("Only integer datatypes can be bit-shifted")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	dest.intval = dest.intval >> bits
	return dest, nil
	
}

func (self *BasicValue) bitwiseNot() (*BasicValue, error) {
	if ( self.valuetype != TYPE_INTEGER ) {
		return nil, errors.New("Cannot only perform bitwise operations on integers")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}	
	dest.intval = ^self.intval
	return dest, nil
}

func (self *BasicValue) bitwiseAnd(rval *BasicValue) (*BasicValue, error) {
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	if ( self.valuetype != TYPE_INTEGER ) {
		return nil, errors.New("Cannot perform bitwise operations on string or float")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	dest.intval = self.intval & rval.intval
	return dest, nil
}

func (self *BasicValue) bitwiseOr(rval *BasicValue) (*BasicValue, error) {
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	if ( self.valuetype != TYPE_INTEGER ) {
		return nil, errors.New("Can only perform bitwise operations on integers")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	dest.intval = self.intval | rval.intval
	return dest, nil
}

func (self *BasicValue) bitwiseXor(rval *BasicValue) (*BasicValue, error) {
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	if ( self.valuetype != TYPE_INTEGER || rval.valuetype != TYPE_INTEGER ) {
		return nil, errors.New("Can only perform bitwise operations on integers")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	dest.intval = self.intval ^ rval.intval
	return dest, nil
}

// TODO: Implement - (remove) * (duplicate) and / (split) on string types, that would be cool

func (self *BasicValue) mathPlus(rval *BasicValue) (*BasicValue, error) {
	var dest *BasicValue
	var err error
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	if ( self.mutable == false ) {
		dest, err = self.clone(nil)
		if ( err != nil ) {
			return nil, err
		}
	} else {
		dest = self
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		dest.intval = self.intval + (rval.intval + int64(rval.floatval))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		dest.floatval = self.floatval + (rval.floatval + float64(rval.intval))
	} else if ( self.valuetype == TYPE_STRING && rval.valuetype == TYPE_STRING ){
		dest.stringval = self.stringval + rval.stringval
	} else if ( self.valuetype == TYPE_STRING && rval.valuetype == TYPE_INTEGER ) {
		dest.stringval = fmt.Sprintf("%s%d", self.stringval, rval.intval)
	} else if ( self.valuetype == TYPE_STRING && rval.valuetype == TYPE_FLOAT ) {
		dest.stringval = fmt.Sprintf("%s%f", self.stringval, rval.floatval)
	} else {
		//fmt.Printf("%+v + %+v\n", self, rval)	
		return nil, errors.New("Invalid arithmetic operation")
	}
	return dest, nil
}


func (self *BasicValue) mathMinus(rval *BasicValue) (*BasicValue, error) {
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	if ( self.valuetype == TYPE_STRING || rval.valuetype == TYPE_STRING ) {
		return nil, errors.New("Cannot perform subtraction on strings")
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		dest.intval = self.intval - (rval.intval + int64(rval.floatval))
	} else {
		dest.floatval = self.floatval - (rval.floatval + float64(rval.intval))
	}
	return dest, nil
}

func (self *BasicValue) mathDivide(rval *BasicValue) (*BasicValue, error) {
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}	
	if ( self.valuetype == TYPE_STRING || rval.valuetype == TYPE_STRING ) {
		return nil, errors.New("Cannot perform division on strings")
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		dest.intval = self.intval / (rval.intval + int64(rval.floatval))
	} else {
		dest.floatval = self.floatval / (rval.floatval + float64(rval.intval))
	}
	return dest, nil
}

func (self *BasicValue) mathMultiply(rval *BasicValue) (*BasicValue, error) {
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	if ( self.valuetype == TYPE_STRING ) {
		if ( rval.valuetype == TYPE_STRING ) {
			return nil, errors.New("String multiplication requires an integer multiple")
		}
		dest.stringval = strings.Repeat(dest.stringval, int(rval.intval))
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		dest.intval = self.intval * (rval.intval + int64(rval.floatval))
	} else {
		dest.floatval = self.floatval * (rval.floatval + float64(rval.intval))
	}
	return dest, nil
}

func (self *BasicValue) lessThan(rval *BasicValue) (*BasicValue, error) {
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		dest.basicBoolValue(self.intval < (rval.intval + int64(rval.floatval)))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		dest.basicBoolValue(self.floatval < (rval.floatval + float64(rval.intval)))
	} else {
		dest.basicBoolValue(strings.Compare(self.stringval, rval.stringval) < 0)
	}
	return dest, nil
}

func (self *BasicValue) lessThanEqual(rval *BasicValue) (*BasicValue, error) {
	var result int
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		dest.basicBoolValue(self.intval <= (rval.intval + int64(rval.floatval)))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		dest.basicBoolValue(self.floatval <= (rval.floatval + float64(rval.intval)))
	} else {
		result = strings.Compare(self.stringval, rval.stringval)
		dest.basicBoolValue(result < 0 || result == 0)
	}
	return dest, nil
}

func (self *BasicValue) greaterThan(rval *BasicValue) (*BasicValue, error) {
	var result int
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		dest.basicBoolValue(self.intval > (rval.intval + int64(rval.floatval)))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		dest.basicBoolValue(self.floatval > (rval.floatval + float64(rval.intval)))
	} else {
		result = strings.Compare(self.stringval, rval.stringval)
		dest.basicBoolValue((result > 0))
	}
	return dest, nil
}

func (self *BasicValue) greaterThanEqual(rval *BasicValue) (*BasicValue, error) {
	var result int
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		dest.basicBoolValue(self.intval >= (rval.intval + int64(rval.floatval)))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		dest.basicBoolValue(self.floatval >= (rval.floatval + float64(rval.intval)))
	} else {
		result = strings.Compare(self.stringval, rval.stringval)
		dest.basicBoolValue(result > 0 || result == 0)
	}
	return dest, nil
}

func (self *BasicValue) isEqual(rval *BasicValue) (*BasicValue, error) {
	var result int
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		dest.basicBoolValue(self.intval == (rval.intval + int64(rval.floatval)))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		dest.basicBoolValue(self.floatval == (rval.floatval + float64(rval.intval)))
	} else {
		result = strings.Compare(self.stringval, rval.stringval)
		dest.basicBoolValue(result == 0)
	}
	//fmt.Printf("isEqual %+v ? %+v\n : %+v", self, rval, dest)
	return dest, nil
}

func (self *BasicValue) isNotEqual(rval *BasicValue) (*BasicValue, error) {
	var result int
	if ( rval == nil ) {
		return nil, errors.New("nil rval")
	}
	dest, err := self.clone(nil)
	if ( err != nil ) {
		return nil, err
	}
	if ( self.valuetype == TYPE_INTEGER ) {
		dest.basicBoolValue(self.intval != (rval.intval + int64(rval.floatval)))
	} else if ( self.valuetype == TYPE_FLOAT ) {
		dest.basicBoolValue(self.floatval != (rval.floatval + float64(rval.intval)))
	} else {
		result = strings.Compare(self.stringval, rval.stringval)
		dest.basicBoolValue(result != 0)
	}
	return dest, nil
}

func (self *BasicValue) isTrue() bool {
	if ( self.valuetype != TYPE_BOOLEAN ) {
		return false
	}
	return (self.boolvalue == BASIC_TRUE)
}

func (self *BasicValue) basicBoolValue(result bool) {
	self.valuetype = TYPE_BOOLEAN
	if ( result == true ) {
		self.boolvalue = BASIC_TRUE
		return
	}
	self.boolvalue = BASIC_FALSE
}


