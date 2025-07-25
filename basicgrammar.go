package main

import (
	"fmt"
	"strconv"
	"errors"
	"strings"
)


type BasicASTLeafType int
const (
	LEAF_UNDEFINED           BasicASTLeafType = iota
	LEAF_LITERAL_INT // 1
	LEAF_LITERAL_FLOAT // 2
	LEAF_LITERAL_STRING // 3
	LEAF_IDENTIFIER // 4
	LEAF_IDENTIFIER_INT // 5
	LEAF_IDENTIFIER_FLOAT // 6
	LEAF_IDENTIFIER_STRING // 7
	LEAF_UNARY // 8
	LEAF_BINARY // 9
	LEAF_GROUPING // 10
	LEAF_EQUALITY // 11
	LEAF_COMPARISON // 12
	LEAF_TERM // 13
	LEAF_PRIMARY // 14
	LEAF_COMMAND // 15
	LEAF_COMMAND_IMMEDIATE // 16
	LEAF_FUNCTION // 17
	LEAF_BRANCH // 18
	LEAF_ARGUMENTLIST // 19
	LEAF_IDENTIFIER_STRUCT // 20
)

type BasicASTLeaf struct {
	leaftype BasicASTLeafType
	literal_int int64
	literal_string string
	literal_float float64
	identifier string
	operator BasicTokenType
	parent *BasicASTLeaf
	left *BasicASTLeaf
	right *BasicASTLeaf
	expr *BasicASTLeaf
}

func (self *BasicASTLeaf) init(leaftype BasicASTLeafType) {
	self.leaftype = leaftype
	self.parent = nil
	self.left = nil
	self.right = nil
	self.expr = nil
	self.identifier = ""
	self.literal_int = 0
	self.literal_float = 0.0
	self.literal_string = ""
	self.operator = UNDEFINED
}

func (self *BasicASTLeaf) clone() *BasicASTLeaf {
	var left *BasicASTLeaf = self.left
	var expr *BasicASTLeaf = self.expr
	var right *BasicASTLeaf = self.right
	if left != nil {
		left = left.clone()
	}
	if right != nil {
		right = right.clone()
	}
	if expr != nil {
		expr = expr.clone()
	}
	return &BasicASTLeaf{
		leaftype: self.leaftype,
		parent: self.parent,
		left: left,
		right: right,
		expr: expr,
		identifier: strings.Clone(self.identifier),
		literal_int: self.literal_int,
		literal_float: self.literal_float,
		literal_string: strings.Clone(self.literal_string),
		operator: self.operator}
}

func (self *BasicASTLeaf) firstArgument() *BasicASTLeaf {
	if ( self.right == nil ||
		self.right.leaftype != LEAF_ARGUMENTLIST ||
		self.right.operator != FUNCTION_ARGUMENT ) {
		return nil
	}
	return self.right.right
}

func (self *BasicASTLeaf) firstSubscript() *BasicASTLeaf {
	if ( self.right == nil ||
		self.right.leaftype != LEAF_ARGUMENTLIST ||
		self.right.operator != ARRAY_SUBSCRIPT ) {
		return nil
	}
	return self.right.right
}

func (self *BasicASTLeaf) isIdentifier() bool {
	return ( self != nil &&
		( self.leaftype == LEAF_IDENTIFIER ||
			self.leaftype == LEAF_IDENTIFIER_INT ||
			self.leaftype == LEAF_IDENTIFIER_FLOAT ||
			self.leaftype == LEAF_IDENTIFIER_STRING ))
}

func (self *BasicASTLeaf) isLiteral() bool {
	return ( self != nil &&
		( self.leaftype == LEAF_LITERAL_INT ||
			self.leaftype == LEAF_LITERAL_FLOAT ||
			self.leaftype == LEAF_LITERAL_STRING ))
}

func (self *BasicASTLeaf) newPrimary(group *BasicASTLeaf, literal_string *string, literal_int *int64, literal_float *float64) error {
	self.init(LEAF_PRIMARY)
	if ( group != nil ) {
		self.expr = group
		return nil
	} else if ( literal_string != nil ) {
		self.literal_string = *literal_string
		return nil
	} else if ( literal_int != nil ) {
		self.literal_int = *literal_int
		return nil
	} else if ( literal_float != nil ) {
		self.literal_float = *literal_float
		return nil
	}
	return errors.New("Gramattically incorrect primary leaf")
}

func (self *BasicASTLeaf) newComparison(left *BasicASTLeaf, op BasicTokenType, right *BasicASTLeaf) error {
	if ( left == nil || right == nil ) {
		return errors.New("nil pointer arguments")
	}
	self.init(LEAF_COMPARISON)
	self.left = left
	self.right = right
	switch (op) {
	case LESS_THAN: fallthrough
	case LESS_THAN_EQUAL: fallthrough
	case NOT_EQUAL: fallthrough
	case GREATER_THAN: fallthrough
	case GREATER_THAN_EQUAL:
		return nil
	default:
		return errors.New(fmt.Sprintf("Invalid operator %d for comparison", op))
	}
}

func (self *BasicASTLeaf) newBinary(left *BasicASTLeaf, op BasicTokenType, right *BasicASTLeaf) error {
	if ( left == nil || right == nil ) {
		return errors.New("nil pointer arguments")
	}
	self.init(LEAF_BINARY)
	self.left = left
	self.right = right
	self.operator = op
	return nil
}

func (self *BasicASTLeaf) newFunction(fname string, right *BasicASTLeaf) error {
	self.init(LEAF_FUNCTION)
	self.right = right
	self.operator = COMMAND
	self.identifier = fname
	return nil
}

func (self *BasicASTLeaf) newCommand(cmdname string, right *BasicASTLeaf) error {
	self.init(LEAF_COMMAND)
	self.right = right
	self.operator = COMMAND
	self.identifier = cmdname
	return nil
}

func (self *BasicASTLeaf) newImmediateCommand(cmdname string, right *BasicASTLeaf) error {
	//fmt.Println("Creating new immediate command leaf")
	self.init(LEAF_COMMAND_IMMEDIATE)
	self.right = right
	self.operator = COMMAND_IMMEDIATE
	self.identifier = cmdname
	return nil
}

func (self *BasicASTLeaf) newUnary(op BasicTokenType, right *BasicASTLeaf) error {
	if ( right == nil ) {
		return errors.New("nil pointer arguments")
	}
	self.init(LEAF_UNARY)
	self.right = right
	self.operator = op
	return nil
}

func (self *BasicASTLeaf) newBranch(expr *BasicASTLeaf, trueleaf *BasicASTLeaf, falseleaf *BasicASTLeaf) error {
	if ( expr == nil ) {
		return errors.New("nil pointer arguments")
	}
	self.init(LEAF_BRANCH)
	self.expr = expr
	self.left = trueleaf
	self.right = falseleaf
	return nil
}

func (self *BasicASTLeaf) newGrouping(expr *BasicASTLeaf) error {
	if ( expr == nil ) {
		return errors.New("nil pointer arguments")
	}
	self.init(LEAF_GROUPING)
	self.expr = expr
	return nil
}

func (self *BasicASTLeaf) newLiteralInt(lexeme string) error {
	var base int = 10
	var err error = nil
	self.init(LEAF_LITERAL_INT)
	if ( len(lexeme) > 2 && lexeme[0:2] == "0x" ) {
		base = 16
	} else if ( lexeme[0] == '0' ) {
		base = 8
	}
	self.literal_int, err = strconv.ParseInt(lexeme, base, 64)
	return err
}

func (self *BasicASTLeaf) newLiteralFloat(lexeme string) error {
	var err error = nil
	self.init(LEAF_LITERAL_FLOAT)
	self.literal_float, err = strconv.ParseFloat(lexeme, 64)
	return err
}

func (self *BasicASTLeaf) newLiteralString(lexeme string) error {
	self.init(LEAF_LITERAL_STRING)
	self.literal_string = lexeme
	return nil
}

func (self *BasicASTLeaf) newIdentifier(leaftype BasicASTLeafType, lexeme string) error {
	self.init(leaftype)
	self.identifier = lexeme
	return nil
}

func (self *BasicASTLeaf) toString() string {
	operatorToStr := func() string {
		switch (self.operator) {
		case EQUAL: return "="
		case LESS_THAN: return "<"
		case GREATER_THAN: return ">"
		case LESS_THAN_EQUAL: return "<="
		case GREATER_THAN_EQUAL: return ">="
		case NOT_EQUAL: return "<>"
		case PLUS: return "+"
		case MINUS: return "-"
		case STAR: return "*"
		case LEFT_SLASH: return "/"
		case CARAT: return "^"
		case NOT: return "NOT"
		case AND: return "AND"
		case OR: return "OR"
		
		}
		return ""
	}
	switch (self.leaftype) {
	case LEAF_LITERAL_INT:
		return fmt.Sprintf("%d", self.literal_int)
	case LEAF_LITERAL_FLOAT:
		return fmt.Sprintf("%f", self.literal_float)
	case LEAF_LITERAL_STRING:
 		return fmt.Sprintf("%s", self.literal_string)
	case LEAF_IDENTIFIER_INT: fallthrough
	case LEAF_IDENTIFIER_FLOAT: fallthrough
	case LEAF_IDENTIFIER_STRING: fallthrough
	case LEAF_IDENTIFIER:
		return fmt.Sprintf("%s", self.identifier)
	case LEAF_IDENTIFIER_STRUCT:
		return fmt.Sprintf("NOT IMPLEMENTED")
	case LEAF_UNARY:
		return fmt.Sprintf(
			"(%s %s)",
			operatorToStr(),
			self.right.toString())
	case LEAF_BINARY:
		return fmt.Sprintf(
			"(%s %s %s)",
			operatorToStr(),
			self.left.toString(),
			self.right.toString())
	case LEAF_GROUPING:
		return fmt.Sprintf(
			"(group %s)",
			self.expr.toString())
	default:
		return fmt.Sprintf("%+v", self)
	}
	return ""
}

