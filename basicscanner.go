/*
 * Scan text from the user
 */

package main

import (
	"fmt"
	"strconv"
	"io"
	"bufio"
	"unicode"
	"errors"
	"strings"
)

type BasicTokenType int
const (
	UNDEFINED     BasicTokenType = iota  // 0
	EQUAL                                // 1
	LESS_THAN                            // 2
	LESS_THAN_EQUAL                      // 3
	GREATER_THAN                         // 4
	GREATER_THAN_EQUAL                   // 5
	BANG                                 // 6
	HASH                                 // 7
	NOT_EQUAL                            // 8
	LEFT_PAREN                           // 9
	RIGHT_PAREN                          // 10
	PLUS                                 // 11
	MINUS                                // 12
	LEFT_SLASH                           // 13
	STAR                                 // 14
	UP_ARROW                             // 15
	LITERAL_STRING                       // 16
	LITERAL_INT                          // 17
	LITERAL_FLOAT                        // 18
	IDENTIFIER                           // 19
	IDENTIFIER_STRING                    // 20
	IDENTIFIER_FLOAT                     // 21
	IDENTIFIER_INT                       // 22
	AND                                  // 23
	OR                                   // 24
	NOT                                  // 25
	PRINT                                // 26
	GOTO                                 // 27
	REM                                  // 28
)

type BasicScanner struct {
	current int
	c rune
	start int
	tokentype BasicTokenType
	context BasicContext
	line string
	tokens [16]BasicToken
	nexttoken int
	hasError bool
	reservedwords map[string]BasicTokenType
}

func (self *BasicScanner) init(context BasicContext) {
	self.current = 0
	self.start = 0
	self.tokentype = UNDEFINED
	self.context = context
	self.nexttoken = 0
	self.hasError = false
	self.reservedwords = make(map[string]BasicTokenType)
	self.reservedwords["PRINT"] = PRINT
	self.reservedwords["GOTO"] = GOTO
	self.reservedwords["REM"] = REM
	self.reservedwords["AND"] = AND
	self.reservedwords["OR"] = OR
	self.reservedwords["NOT"] = NOT
}

func (self *BasicScanner) addToken(token BasicTokenType, lexeme string) {
	self.tokens[self.nexttoken] = BasicToken{
		tokentype: token,
		lineno: self.context.lineno,
		literal: "",
		lexeme: lexeme}
	fmt.Printf("%+v\n", self.tokens[self.nexttoken])
	self.nexttoken += 1
}

func (self *BasicScanner) getLexeme() string {
	if ( self.current == len(self.line) ) {
		return self.line[self.start:]
	} else {
		if ( self.start == self.current ) {
			return string(self.line[self.start])
		}
		return self.line[self.start:self.current]
	}		
}

func (self *BasicScanner) peek() (rune, error) {
	if ( self.isAtEnd() ) {
		return rune(0), errors.New("End Of Line")
	}
	return rune(self.line[self.current]), nil
}

func (self *BasicScanner) peekNext() (rune, error) {
	if ( (self.current + 1) >= len(self.line) ) {
		return rune(0), errors.New("End Of Line")
	}
	return rune(self.line[self.current+1]), nil
}

func (self *BasicScanner) advance() (rune, error) {
	if ( self.isAtEnd() ) {
		return rune(0), errors.New("End Of Line")
	}
	c := rune(self.line[self.current])
	self.current += 1
	return c, nil
}

func (self *BasicScanner) isAtEnd() bool {
	return (self.current >= len(self.line))
}

func (self *BasicScanner) matchNextChar(cm rune, truetype BasicTokenType, falsetype BasicTokenType) {
	if ( self.current == len(self.line)-1 ) {
		self.tokentype = falsetype
	} else if ( rune(self.line[self.current+1]) == cm ) {
		self.current += 1
		self.tokentype = truetype
	} else {
		self.tokentype = falsetype
	}
}

func (self *BasicScanner) matchString() {
	for !self.isAtEnd() {
		c, err := self.peek()
		if ( err != nil ) {
			basicError(self.context.lineno, PARSE, "UNTERMINATED STRING LITERAL\n")
			self.hasError = true
			return
		}
		if ( c == '"' ) {
			break
		} else {
			self.current += 1
		}
	}
	self.tokentype = LITERAL_STRING
}

func (self *BasicScanner) matchNumber() {
	var linenumber bool = (self.nexttoken == 0)
	self.tokentype = LITERAL_INT
	for !self.isAtEnd() {
		// Discard the error, we're checking isAtEnd()
		c, _ := self.peek()
		if ( ! unicode.IsDigit(c) ) {
			break
		} else if ( c == '.' ) {
			nc, err := self.peekNext()
			if ( err != nil || !unicode.IsDigit(nc) ) {
				basicError(self.context.lineno, PARSE, "INVALID FLOATING POINT LITERAL\n")
				self.hasError = true
				return
			}
			self.tokentype = LITERAL_FLOAT
		}
		self.current += 1
	}
	if ( self.tokentype == LITERAL_INT && linenumber == true ) {
		lineno, err := strconv.Atoi(self.getLexeme())
		if ( err != nil ) {
			basicError(self.context.lineno, PARSE, fmt.Sprintf("INTEGER CONVERSION ON '%s'", self.getLexeme()))
			self.hasError = true
		}
		self.context.lineno = lineno
		self.context.source[self.context.lineno] = self.line
	}
}

func (self *BasicScanner) matchIdentifier() {
	var matchedReservedWord = false
	var identifierSoFar string
	var reservedIdentifier BasicTokenType
	for !self.isAtEnd() {
		// Discard the error, we're checking isAtEnd()
		c, _ := self.peek()
		if ( unicode.IsDigit(c) || unicode.IsLetter(c) ) {
			self.current += 1
		} else {
			identifierSoFar = strings.ToUpper(self.getLexeme())
			reservedIdentifier = self.reservedwords[identifierSoFar]
			if ( reservedIdentifier != UNDEFINED ) {
				matchedReservedWord = true
			}
			
			switch (c) {
			case '$':
				self.tokentype = IDENTIFIER_STRING
				self.current += 1
			case '%':
				self.tokentype = IDENTIFIER_FLOAT
				self.current += 1
			case '#':
				self.tokentype = IDENTIFIER_INT
				self.current += 1
			default:
				self.tokentype = IDENTIFIER
			}
			
			break
		}
	}
	// Look for reserved words in variable identifiers
	if ( self.tokentype != IDENTIFIER && matchedReservedWord ) {
		basicError(self.context.lineno, SYNTAX, "Reserved word in variable name\n")
		self.hasError = true
	} else if ( reservedIdentifier != UNDEFINED ) {
		self.tokentype = reservedIdentifier
	}
}

func (self *BasicScanner) scanTokens(line string) {

	var c rune
	self.line = line
	self.nexttoken = 0
	self.current = 0
	self.start = 0
	self.hasError = false
	for !self.isAtEnd() {
		// Discard the error, we're doing our own isAtEnd()
		c, _ = self.advance()
		switch (c) {
		case '^': self.tokentype = UP_ARROW
		case '(': self.tokentype = LEFT_PAREN
		case ')': self.tokentype = RIGHT_PAREN
		case '+': self.tokentype = PLUS
		case '-': self.tokentype = MINUS
		case '/': self.tokentype = LEFT_SLASH
		case '*': self.tokentype = STAR
		case '!': self.tokentype = BANG
		case '=': self.tokentype = EQUAL
		case '<':
			// I'm being lazy here
			self.matchNextChar('=', LESS_THAN_EQUAL, LESS_THAN)
			self.matchNextChar('>', NOT_EQUAL, LESS_THAN)
		case '>': self.matchNextChar('=', GREATER_THAN_EQUAL, GREATER_THAN)
		case '"': self.matchString()
		case ' ':
			self.start = self.current
			break
		case '\t': fallthrough
		case '\r': fallthrough
		case '\n':
			return
		default:
			if ( unicode.IsDigit(c) ) {
				self.matchNumber()
			} else if ( unicode.IsLetter(c) ) {
				self.matchIdentifier()
			} else {
				basicError(self.context.lineno, PARSE, fmt.Sprintf("UKNOWN TOKEN %c\n", c))
				self.hasError = true
				self.start = self.current
			}
		}
		if ( self.tokentype != UNDEFINED && self.hasError == false ) {
			self.addToken(self.tokentype, self.getLexeme())
			if ( self.tokentype == LITERAL_STRING ) {
				// String parsing stops on the final ",
				// move past it.
				self.current += 1
			}
			self.tokentype = UNDEFINED
			self.start = self.current
		}
	}
}

func (self *BasicScanner) repl(fileobj io.Reader) {
	var readbuff = bufio.NewScanner(fileobj)
	
	fmt.Println("READY")
	for readbuff.Scan() {
		self.scanTokens(readbuff.Text())
		fmt.Println("READY")
	}	
}

