package main

import (
	"fmt"
	"strconv"
	"io"
	"bufio"
	/*"os"*/
	"unicode"
	"errors"
)

type BasicTokenType int
type BasicType int
type BasicError int

const (
	UNDEFINED     BasicTokenType = iota
	EQUAL
	LESS_THAN
	LESS_THAN_EQUAL
	GREATER_THAN
	GREATER_THAN_EQUAL
	BANG
	BANG_EQUAL
	LEFT_PAREN
	RIGHT_PAREN
	PLUS
	MINUS
	LEFT_SLASH
	STAR
	PRINT
	GOTO
	REM
	LITERAL_STRING
	LITERAL_INT
	LITERAL_FLOAT
)

const (
	INTEGER      BasicType = iota
	STRING
)

const (
	IO           BasicError = iota
	PARSE
	EXECUTE	
)

type BasicLiteral struct {
	literaltype BasicType
	stringval string
	intval int
}

type BasicToken struct {
	tokentype BasicTokenType
	lineno int
	literal string
	lexeme string
}

func (self BasicToken) toString() string {
	return fmt.Sprintf("%d %s %s", self.tokentype, self.lexeme, self.literal)
}

type BasicContext struct {
	source [9999]string
	lineno int
}

func (self BasicContext) init() {
	self.lineno = 0
}

type BasicScanner struct {
	current int
	c rune
	start int
	tokentype BasicTokenType
	context BasicContext
	line string
	tokens [16]BasicToken
	nexttoken int
}

func (self *BasicScanner) init(context BasicContext) {
	self.current = 0
	self.start = 0
	self.tokentype = UNDEFINED
	self.context = context
	self.nexttoken = 0
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
		}
		self.context.lineno = lineno
		self.context.source[self.context.lineno] = self.line
	}
}

func (self *BasicScanner) scanTokens(line string) {

	var c rune
	self.line = line
	self.nexttoken = 0
	self.current = 0
	self.start = 0
	for !self.isAtEnd() {
		// Discard the error, we're doing our own isAtEnd()
		c, _ = self.advance()
		switch (c) {
		case '(': self.tokentype = LEFT_PAREN
		case ')': self.tokentype = RIGHT_PAREN
		case '+': self.tokentype = PLUS
		case '-': self.tokentype = MINUS
		case '/': self.tokentype = LEFT_SLASH
		case '*': self.tokentype = STAR
		case '!': self.matchNextChar('=', BANG_EQUAL, BANG)
		case '<': self.matchNextChar('=', LESS_THAN_EQUAL, LESS_THAN)
		case '>': self.matchNextChar('=', GREATER_THAN_EQUAL, GREATER_THAN)
		case '"': self.matchString()
		case ' ':
			self.start = self.current+1
			break
		case '\t': fallthrough
		case '\r': fallthrough
		case '\n':
			return
		default:
			if ( unicode.IsDigit(c) ) {
				self.matchNumber()
			} else {
				basicError(self.context.lineno, PARSE, fmt.Sprintf("UKNOWN TOKEN %c\n", c))
				self.start = self.current
			}
		}
		if ( self.tokentype != UNDEFINED ) {
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

func errorCodeToString(errno BasicError) string {
	switch (errno) {
	case IO:
		return "IO"
	case PARSE:
		return "PARSE"
	case EXECUTE:
		return "EXEC"
	}
	return "UNDEF"
}

func basicError(line int, errno BasicError, message string) {
	fmt.Printf("? %s %s", errorCodeToString(errno), message)
}

func main() {
	var context BasicContext;
	var scanner BasicScanner;
	context.init()
	scanner.init(context)
	//scanner.repl(os.Stdin)
	scanner.scanTokens("10 PRINT \"HELLO\"")
	scanner.scanTokens("20 ABC=3+2")
	scanner.scanTokens("30 XYZ=(3+(4*5))")
}
