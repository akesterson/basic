package main

import (
	"errors"
	"strings"
)

func (self *BasicParser) ParseCommandFOR() (*BasicASTLeaf, error) {
	// FOR     ...        TO ....        [STEP    ...]
	// COMMAND ASSIGNMENT    EXPRESSION  [COMMAND EXPRESSION]
	// Set up:
	//    self.runtime.environment.forStepLeaf with the step expression
	//    self.runtime.environment.forToLeaf with the TO expression
	//    self.runtime.environment.forFirstLine with the first line of the FOR code
	// Return the FOR +assignment
	
	var assignment *BasicASTLeaf = nil
	var operator *BasicToken = nil
	var err error = nil
	var expr *BasicASTLeaf = nil
	
	assignment, err = self.assignment()
	if ( err != nil || !self.match(COMMAND) ) {
		goto _basicparser_parsecommandfor_error
	}
	operator, err = self.previous()
	if ( err != nil || strings.Compare(operator.lexeme, "TO") != 0 ) {
		goto _basicparser_parsecommandfor_error
	}
	self.runtime.environment.forToLeaf, err = self.expression()
	if ( err != nil ) {
		return nil, err
	}
	if ( self.match(COMMAND) ) {
		operator, err = self.previous()
		if ( err != nil || strings.Compare(operator.lexeme, "STEP") != 0) {
			goto _basicparser_parsecommandfor_error
		}
		self.runtime.environment.forStepLeaf, err = self.expression()
		if ( err != nil ) {
			return nil, err
		}
	} else {
		// Use a default step of 1
		self.runtime.environment.forStepLeaf, err = self.newLeaf()
		self.runtime.environment.forStepLeaf.newLiteralInt("1")
	}
	self.runtime.environment.forFirstLine = (self.runtime.lineno + 1)
	expr, err = self.newLeaf()
	if ( err != nil ) {
		return nil, err
	}
	expr.newCommand("FOR", assignment)
	return expr, nil
	
_basicparser_parsecommandfor_error:
	return nil, errors.New("Expected FOR (assignment) TO (expression) [STEP (expression)]")	
}

func (self *BasicParser) ParseCommandIF() (*BasicASTLeaf, error) {
	// IF      ...          THEN      ....                [ : ELSE    .... ]
	// COMMAND RELATION     COMMAND   COMMAND EXPRESSION  [ : COMMAND EXPRESSION ]
	//  
	// IF      1 == 1       THEN      PRINT "HELLO"         : ELSE PRINT "GOODBYE"
	//
	// BRANCH(THEN_COMMAND, RELATION, ELSE_COMMAND)
	
	var then_command *BasicASTLeaf = nil;
	var else_command *BasicASTLeaf = nil;
	var relation *BasicASTLeaf = nil;
	var branch *BasicASTLeaf = nil;
	var operator *BasicToken = nil;
	var err error = nil;

	relation, err = self.relation()
	if ( err != nil || !self.match(COMMAND) ) {
		return nil, errors.New("Expected IF ... THEN")
	}
	operator, err = self.previous()
	if ( err != nil || strings.Compare(operator.lexeme, "THEN") != 0 ) {
		return nil, errors.New("Expected IF ... THEN")
	}
	then_command, err = self.command()
	if ( err != nil || self.match(COLON) ) {
		if ( ! self.match(COMMAND) ) {
			return nil, errors.New("Expected IF ... THEN ... :ELSE ...")
		}
		operator, err = self.previous()
		if ( err != nil || strings.Compare(operator.lexeme, "ELSE") != 0 ) {
			return nil, errors.New("Expected IF ... THEN ... :ELSE ...")
		}
		else_command, err = self.command()
		if ( err != nil ) {
			return nil, errors.New("Expected IF ... THEN ... :ELSE ...")
		}
	}
	branch, err = self.newLeaf()
	if ( err != nil ) {
		return nil, err
	}
	branch.newBranch(relation, then_command, else_command)
	return branch, nil
}
