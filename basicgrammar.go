package main

import (
	"fmt"
)

/*

   expression -> literal
	       | unary
	       | binary
	       | grouping

   literal     -> INT | FLOAT | STRING
   identifier  -> IDENTIFIER NAME
   grouping    -> "(" expression ")"
   unary       -> ( "-" | "NOT" ) expression
   binary      -> expression operator expression
   operator    -> "=" | "<" | ">" | "<=" | ">=" | "<>"
                | "+" | "-" | "*" | "/"
*/

type BasicASTLeafType int
const (
	LEAF_UNDEFINED           BasicASTLeafType = iota
	LEAF_LITERAL_INT
	LEAF_LITERAL_FLOAT
	LEAF_LITERAL_STRING
	LEAF_IDENTIFIER
	LEAF_UNARY
	LEAF_BINARY
	LEAF_GROUPING
)

type BasicASTLeaf struct {
	leaftype BasicASTLeafType
	literal_int int
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
}

func (self *BasicASTLeaf) newBinary(left *BasicASTLeaf, op BasicTokenType, right *BasicASTLeaf) {
	self.init(LEAF_BINARY)
	self.left = left
	self.right = right
	self.operator = op
}

func (self *BasicASTLeaf) newUnary(op BasicTokenType, right *BasicASTLeaf) {
	self.init(LEAF_UNARY)
	self.right = right
	self.operator = op
}

func (self *BasicASTLeaf) newGrouping(expr *BasicASTLeaf) {
	self.init(LEAF_GROUPING)
	self.expr = expr
}

func (self *BasicASTLeaf) newLiteralInt(val int) {
	self.init(LEAF_LITERAL_INT)
	self.literal_int = val
}

func (self *BasicASTLeaf) newLiteralFloat(val float64) {
	self.init(LEAF_LITERAL_FLOAT)
	self.literal_float = val
}

func (self *BasicASTLeaf) newLiteralString(val string) {
	self.init(LEAF_LITERAL_STRING)
	self.literal_string = val
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
	case LEAF_IDENTIFIER:
		return fmt.Sprintf("%s", self.identifier)
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
	}
	return ""
}

