package main

import (
	"fmt"
	"errors"
)

type BasicVariable struct {
	name string
	valuetype BasicType
	values []BasicValue
	dimensions []int64
	runtime *BasicRuntime
	mutable bool
}

func (self *BasicVariable) init(runtime *BasicRuntime, sizes []int64) error {
	var totalSize int64 = 1
	var i int64 = 0
	var runes = []rune(self.name)
	var value *BasicValue = nil
	//fmt.Printf("Initializing %s\n", self.name)
	if ( runtime == nil ) {
		return errors.New("NIL runtime provided to BasicVariable.init")
	}
	if len(runes) > 0 {
		lastRune := runes[len(runes)-1]
		switch(lastRune) {
			case '$':
			self.valuetype = TYPE_STRING
			case '#':
			self.valuetype = TYPE_INTEGER
			case '%':
			self.valuetype = TYPE_FLOAT
		}
	} else {
		return errors.New("Invalid variable name")
	}
	//fmt.Printf("Setting type to %d from name\n", self.valuetype)
	//if ( len(sizes) == 0 ) {
	//	sizes = make([]int64, 1)
	//	sizes[0] = 10
	//}
	self.runtime = runtime
	self.dimensions = make([]int64, len(sizes))
	copy(self.dimensions, sizes)
	//fmt.Printf("Setting variable dimensions (%+v)\n", self.dimensions)
	for _, size := range sizes {
		//fmt.Printf("Dimension %d is %d\n", i, size)
		if ( size <= 0 )  {
			return errors.New("Array dimensions must be positive integers")
		}
		totalSize *= size
	}
	//fmt.Printf("%s has %d dimensions with %d total objects\n", self.name, len(sizes), totalSize)
	self.values = make([]BasicValue, totalSize)
	for i = 0; i < totalSize ; i++ {
		value = &self.values[i]
		value.init()
		value.zero()
		value.runtime = runtime
		value.valuetype = self.valuetype
		value.mutable = true
	}
	return nil
}

func (self *BasicVariable) set(value *BasicValue, subscripts ...int64) (error){
	return self.setSubscript(value, subscripts...)
}

func (self *BasicVariable) setInteger(value int64, subscripts ...int64) (error) {
	return self.setSubscript(&BasicValue{
		stringval: "",
		intval: value,
		floatval: 0.0,
		boolvalue: BASIC_FALSE,
		runtime: self.runtime,
		mutable: false,
		valuetype: TYPE_INTEGER},
		subscripts...)
}

func (self *BasicVariable) setFloat(value float64, subscripts ...int64) (error) {
	return self.setSubscript(&BasicValue{
		stringval: "",
		intval: 0,
		floatval: value,
		boolvalue: BASIC_FALSE,
		runtime: self.runtime,
		mutable: false,
		valuetype: TYPE_FLOAT},
		subscripts...)
}

func (self *BasicVariable) setString(value string, subscripts ...int64) (error) {
	return self.setSubscript(&BasicValue{
		stringval: value,
		intval: 0,
		floatval: 0.0,
		boolvalue: BASIC_FALSE,
		runtime: self.runtime,
		mutable: false,
		valuetype: TYPE_STRING},
		subscripts...)
}

func (self *BasicVariable) setBoolean(value bool, subscripts ...int64) (error) {
	var boolvalue int64
	if ( value == true ) {
		boolvalue = BASIC_TRUE
	} else {
		boolvalue = BASIC_FALSE
	}
			
	return self.setSubscript(&BasicValue{
		stringval: "",
		intval: 0,
		floatval: 0.0,
		boolvalue: boolvalue,
		runtime: self.runtime,
		mutable: false,
		valuetype: TYPE_STRING},
		subscripts...)
}

func (self *BasicVariable) zero() {
	self.valuetype = TYPE_UNDEFINED
	self.mutable = true
}


func (self *BasicVariable) getSubscript(subscripts ...int64) (*BasicValue, error) {
	var index int64
	var err error = nil
	if ( len(subscripts) != len(self.dimensions) ) {
		return nil, fmt.Errorf("Variable %s has %d dimensions, received %d", self.name, len(self.dimensions), len(subscripts))
	}
	index, err = self.flattenIndexSubscripts(subscripts)
	if ( err != nil ) {
		return nil, err
	}
	return &self.values[index], nil
}

func (self *BasicVariable) setSubscript(value *BasicValue, subscripts ...int64) error {
	var index int64
	var err error = nil
	if ( len(subscripts) != len(self.dimensions) ) {
		return fmt.Errorf("Variable %s has %d dimensions, received %d", self.name, len(self.dimensions), len(subscripts))
	}
	index, err = self.flattenIndexSubscripts(subscripts)
	if ( err != nil ) {
		return err
	}
	value.clone(&self.values[index])
	return nil
}

func (self *BasicVariable) flattenIndexSubscripts(subscripts []int64) (int64, error) {
	var flatIndex int64 = 0
	var multiplier int64 = 1
	var i int = 0

	for i = len(subscripts) - 1; i >= 0 ; i-- {
		if ( subscripts[i] < 0 || subscripts[i] >= self.dimensions[i] ) {
			return 0, fmt.Errorf("Variable index access out of bounds at dimension %d: %d (max %d)", i, subscripts[i], self.dimensions[i]-1)
		}
		flatIndex += subscripts[i] * multiplier
		multiplier *= self.dimensions[i]
	}
	return flatIndex, nil
}

func (self *BasicVariable) toString() (string) {
	if ( len(self.values) == 0 ) {
		return self.values[0].toString()
	} else {
		return "toString() not implemented for arrays"
	}
}
