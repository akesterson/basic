package main

import (
	"errors"
	"strings"
)

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
