package main

import (
	"errors"
	"strings"
)

func (self *BasicParser) ParseCommandDEF() (*BasicASTLeaf, error) {
	// DEF    NAME       (A, ...)        =           ....
	// COMMAND IDENTIFIER ARGUMENTLIST    ASSIGNMENT  EXPRESSION
	var command *BasicASTLeaf = nil
	var identifier *BasicASTLeaf = nil
	var arglist *BasicASTLeaf = nil
	var expression *BasicASTLeaf = nil
	var err error = nil

	identifier, err = self.primary()
	if ( err != nil ) {
		return nil, err
	}
	if ( identifier.leaftype != LEAF_IDENTIFIER ) {
		return nil, errors.New("Expected identifier")
	}
	arglist, err = self.argumentList()
	if ( err != nil ) {
		return nil, errors.New("Expected argument list (identifier names)")
	}
	expression = arglist
	for ( expression.right != nil ) {
		switch (expression.right.leaftype) {
		case LEAF_IDENTIFIER_STRING: fallthrough
		case LEAF_IDENTIFIER_INT: fallthrough
		case LEAF_IDENTIFIER_FLOAT:
			break
		default:
			return nil, errors.New("Only variable identifiers are valid arguments for DEF")
		}
	}
	if self.match(ASSIGNMENT) {
		expression, err = self.expression()
		if ( err != nil ) {
			return nil, err
		}
	}
	command, err = self.newLeaf()
	if ( err != nil ) {
		return nil, err
	}
	
	self.runtime.scanner.functions[identifier.literal_string] = FUNCTION
	command.newCommand("DEF", identifier)
	command.left = arglist
	command.expr = expression
	return command, nil
}

func (self *BasicParser) ParseCommandFOR() (*BasicASTLeaf, error) {
	// FOR     ...        TO ....        [STEP    ...]
	// COMMAND ASSIGNMENT    EXPRESSION  [COMMAND EXPRESSION]
	// Set up:
	//    self.runtime.environment.forStepLeaf with the step expression
	//    self.runtime.environment.forToLeaf with the TO expression
	//    self.runtime.environment.loopFirstLine with the first line of the FOR code
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
		// Let the runtime determine the correct default step
		self.runtime.environment.forStepLeaf, err = self.newLeaf()
		self.runtime.environment.forStepLeaf.newLiteralInt("0")
	}
	self.runtime.environment.loopFirstLine = (self.runtime.lineno + 1)
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
	if ( err != nil ) {
		return nil, err
	}
	if (!self.match(COMMAND) ) {
		return nil, errors.New("Incomplete IF statement")
	}
	operator, err = self.previous()
	if ( err != nil || strings.Compare(operator.lexeme, "THEN") != 0 ) {
		return nil, errors.New("Expected IF ... THEN")
	}
	then_command, err = self.command()
	if ( self.match(COMMAND) ) {
		operator, err = self.previous()
		if ( err != nil || strings.Compare(operator.lexeme, "ELSE") != 0 ) {
			return nil, errors.New("Expected IF ... THEN ... ELSE ...")
		}
		else_command, err = self.command()
		if ( err != nil ) {
			return nil, errors.New("Expected IF ... THEN ... ELSE ...")
		}
	}
	branch, err = self.newLeaf()
	if ( err != nil ) {
		return nil, err
	}
	branch.newBranch(relation, then_command, else_command)
	return branch, nil
}
