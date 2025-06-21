/*
 * Scan text from the user
 */

package main

import (
	"fmt"
	"strconv"
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
	COMMA                                // 6
	HASH                                 // 7
	NOT_EQUAL                            // 8
	LEFT_PAREN                           // 9
	RIGHT_PAREN                          // 10
	PLUS                                 // 11
	MINUS                                // 12
	LEFT_SLASH                           // 13
	STAR                                 // 14
	CARAT                                // 15
	LITERAL_STRING                       // 16
	LITERAL_INT                          // 17
	LITERAL_FLOAT                        // 18
	IDENTIFIER                           // 19
	IDENTIFIER_STRING                    // 20
	IDENTIFIER_FLOAT                     // 21
	IDENTIFIER_INT                       // 22
	COLON // 23 (:)
	AND // 24
	NOT // 25
	OR // 26
	REM // 27
	EOL // 28 (End of Line)
	EOF // 29 (End of File)
	LINE_NUMBER // 30 (a literal integer that was encountered at the beginning of the line and thus is a line number)
	COMMAND // 31
	COMMAND_IMMEDIATE // 32
	FUNCTION // 33
	ASSIGNMENT // 34
	LEFT_SQUAREBRACKET // 35
	RIGHT_SQUAREBRACKET // 36
)

type BasicScanner struct {
	current int
	c rune
	start int
	tokentype BasicTokenType
	runtime *BasicRuntime
	parser *BasicParser
	line string
	hasError bool
	reservedwords map[string]BasicTokenType
	commands map[string]BasicTokenType
	functions map[string]BasicTokenType
}

func (self *BasicScanner) zero() {
	self.current = 0
	self.start = 0
	self.hasError = false
}

func (self *BasicScanner) init(runtime *BasicRuntime) error {
	if ( runtime == nil ) {
		return errors.New("nil pointer argument")
	}
	self.zero()
	self.runtime = runtime
	if len(self.reservedwords) == 0 {
		self.reservedwords = make(map[string]BasicTokenType)
		self.reservedwords["REM"] = REM
		self.reservedwords["AND"] = AND
		self.reservedwords["OR"] = OR
		self.reservedwords["NOT"] = NOT
	}
	if len(self.commands) == 0 {
		self.commands = make(map[string]BasicTokenType)
		// self.commands["APPEND"] =  COMMAND
		// self.commands["ATN"] =  COMMAND
		self.commands["AUTO"] =  COMMAND_IMMEDIATE
		// self.commands["BACKUP"] =  COMMAND
		// self.commands["BANK"] =  COMMAND
		// self.commands["BEGIN"] =  COMMAND
		// self.commands["BEND"] =  COMMAND
		// self.commands["BLOAD"] =  COMMAND
		// self.commands["BOOT"] =  COMMAND
		// self.commands["BOX"] =  COMMAND
		// self.commands["BSAVE"] =  COMMAND
		// self.commands["CALLFN"] =  COMMAND
		// self.commands["CATALOG"] =  COMMAND
		// self.commands["CHAR"] =  COMMAND
		// self.commands["CHARCIRCLE"] =  COMMAND
		// self.commands["CLOSE"] =  COMMAND
		// self.commands["CLR"] =  COMMAND
		// self.commands["CMD"] =  COMMAND
		// self.commands["COLLECT"] =  COMMAND
		// self.commands["COLLISION"] =  COMMAND
		// self.commands["COLOR"] =  COMMAND
		// self.commands["CONCAT"] =  COMMAND
		// self.commands["CONT"] =  COMMAND
		// self.commands["COPY"] =  COMMAND
		// self.commands["DATA"] =  COMMAND
		// self.commands["DCLEAR"] =  COMMAND
		// self.commands["DCLOSE"] =  COMMAND
		self.commands["DEF"] =  COMMAND
		// self.commands["DELETE"] =  COMMAND
		self.commands["DIM"] =  COMMAND
		// self.commands["DIRECTORY"] =  COMMAND
		// self.commands["DLOAD"] =  COMMAND
		// self.commands["DO"] =  COMMAND
		// self.commands["DOPEN"] =  COMMAND
		// self.commands["DRAW"] =  COMMAND
		// self.commands["DSAVE"] =  COMMAND
		// self.commands["DVERIFY"] =  COMMAND
		self.commands["ELSE"] =  COMMAND
		// self.commands["END"] =  COMMAND
		// self.commands["ENVELOPE"] =  COMMAND
		// self.commands["ER"] =  COMMAND
		// self.commands["ERR"] =  COMMAND
		self.commands["EXIT"] =  COMMAND
		// self.commands["FAST"] =  COMMAND
		// self.commands["FETCH"] =  COMMAND
		// self.commands["FILTER"] =  COMMAND
		self.commands["FOR"] =  COMMAND
		// self.commands["GET"] =  COMMAND
		// self.commands["GETIO"] =  COMMAND
		// self.commands["GETKEY"] =  COMMAND
		self.commands["GOSUB"] =  COMMAND
		self.commands["GOTO"] =  COMMAND
		// self.commands["GRAPHIC"] =  COMMAND
		// self.commands["GSHAPE"] =  COMMAND
		// self.commands["HEADER"] =  COMMAND
		// self.commands["HELP"] =  COMMAND
		self.commands["IF"] =  COMMAND
		// self.commands["INPUT"] =  COMMAND
		// self.commands["INPUTIO"] =  COMMAND
		// self.commands["KEY"] =  COMMAND
		// self.commands["ABS"] =  COMMAND
		self.commands["LET"] =  COMMAND
		self.commands["LIST"] =  COMMAND_IMMEDIATE
		// self.commands["LOAD"] =  COMMAND
		// self.commands["LOCATE"] =  COMMAND
		// self.commands["LOOP"] =  COMMAND
		// self.commands["MONITOR"] =  COMMAND
		// self.commands["MOVSPR"] =  COMMAND
		// self.commands["NEW"] =  COMMAND
		self.commands["NEXT"] =  COMMAND
		// self.commands["ON"] =  COMMAND
		// self.commands["OPENIO"] =  COMMAND
		// self.commands["PAINT"] =  COMMAND
		// self.commands["PLAY"] =  COMMAND
		// self.commands["POKE"] =  COMMAND
		self.commands["PRINT"] =  COMMAND
		// self.commands["PRINTIO"] =  COMMAND
		// self.commands["PUDEF"] =  COMMAND
		self.commands["QUIT"] = COMMAND_IMMEDIATE
		// self.commands["READ"] =  COMMAND
		// self.commands["RECORDIO"] =  COMMAND
		// self.commands["RENAME"] =  COMMAND
		// self.commands["RENUMBER"] =  COMMAND
		// self.commands["RESTORE"] =  COMMAND
		// self.commands["RESUME"] =  COMMAND
		self.commands["RETURN"] =  COMMAND
		self.commands["RUN"] =  COMMAND_IMMEDIATE
		// self.commands["SAVE"] =  COMMAND
		// self.commands["SCALE"] =  COMMAND
		// self.commands["SCNCLR"] =  COMMAND
		// self.commands["SCRATCH"] =  COMMAND
		// self.commands["SLEEP"] =  COMMAND
		// self.commands["SOUND"] =  COMMAND
		// self.commands["SPRCOLOR"] =  COMMAND
		// self.commands["SPRDEF"] =  COMMAND
		// self.commands["SPRITE"] =  COMMAND
		// self.commands["SPRSAV"] =  COMMAND
		// self.commands["SSHAPE"] =  COMMAND
		// self.commands["STASH"] =  COMMAND
		self.commands["STEP"] =  COMMAND
		// self.commands["STOP"] =  COMMAND
		// self.commands["SWAP"] =  COMMAND
		// self.commands["SYS"] =  COMMAND
		// self.commands["TEMPO"] =  COMMAND
		self.commands["THEN"] =  COMMAND
		// self.commands["TI"] =  COMMAND
		self.commands["TO"] =  COMMAND
		// self.commands["TRAP"] =  COMMAND
		// self.commands["TROFF"] =  COMMAND
		// self.commands["TRON"] =  COMMAND
		// self.commands["UNTIL"] =  COMMAND
		// self.commands["USING"] =  COMMAND
		// self.commands["VERIFY"] =  COMMAND
		// self.commands["VOL"] =  COMMAND
		// self.commands["WAIT"] =  COMMAND
		// self.commands["WAIT"] =  COMMAND
		// self.commands["WHILE"] =  COMMAND
		// self.commands["WIDTH"] =  COMMAND
		// self.commands["WINDOW"] =  COMMAND
	}
	if len(self.functions) == 0 {
		self.functions = make(map[string]BasicTokenType)
		// self.functions["ASC"] =  FUNCTION
		// self.functions["BUMP"] =  FUNCTION
		// self.functions["CHR"] =  FUNCTION
	        // self.functions["COS"] =  FUNCTION
		// self.functions["FRE"] =  FUNCTION
		// self.functions["HEX"] =  FUNCTION
		// self.functions["INSTR"] =  FUNCTION
		// self.functions["INT"] =  FUNCTION
		// self.functions["JOY"] =  FUNCTION
		self.commands["LEN"] =  FUNCTION
		// self.functions["LEFT"] =  FUNCTION
		// self.functions["LOG"] =  FUNCTION
		self.commands["MID"] =  FUNCTION
		// self.functions["PEEK"] =  FUNCTION
		// self.functions["POINTER"] =  FUNCTION
		// self.functions["POS"] =  FUNCTION
		// self.functions["POT"] =  FUNCTION
		// self.functions["RCLR"] =  FUNCTION
		// self.functions["RDOT"] =  FUNCTION
		// self.functions["RGR"] =  FUNCTION
		// self.functions["RIGHT"] =  FUNCTION
		// self.functions["RND"] =  FUNCTION
		// self.functions["RSPCOLOR"] =  FUNCTION
		// self.functions["RSPPOS"] =  FUNCTION
		// self.functions["RSPRITE"] =  FUNCTION
		// self.functions["RWINDOW"] =  FUNCTION
		// self.functions["SGN"] =  FUNCTION
		// self.functions["SIN"] =  FUNCTION
		// self.functions["SPC"] =  FUNCTION
		// self.functions["SQR"] =  FUNCTION
		// self.functions["STR"] =  FUNCTION
		// self.functions["TAB"] =  FUNCTION
		// self.functions["TAN"] =  FUNCTION
		// self.functions["USR"] =  FUNCTION
		// self.functions["VAL"] =  FUNCTION
		// self.functions["XOR"] =  FUNCTION
	}
	return nil
}

func (self *BasicScanner) addToken(token BasicTokenType, lexeme string) {
	self.runtime.parser.tokens[self.runtime.parser.nexttoken].tokentype = token
	self.runtime.parser.tokens[self.runtime.parser.nexttoken].lineno = self.runtime.lineno
	self.runtime.parser.tokens[self.runtime.parser.nexttoken].lexeme = lexeme
	
	//fmt.Printf("%+v\n", self.runtime.parser.tokens[self.runtime.parser.nexttoken])
	self.runtime.parser.nexttoken += 1
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

func (self *BasicScanner) matchNextChar(cm rune, truetype BasicTokenType, falsetype BasicTokenType) bool {
	var nc rune
	var err error
	nc, err = self.peek()
	if ( err != nil ) {
		return false
	}
	if ( nc == cm ) {
		self.current += 1
		self.tokentype = truetype
		return true
	} else {
		self.tokentype = falsetype
		return false
	}
}

func (self *BasicScanner) matchString() {
	for !self.isAtEnd() {
		c, err := self.peek()
		if ( err != nil ) {
			self.runtime.basicError(PARSE, "UNTERMINATED STRING LITERAL\n")
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
	var linenumber bool = (self.runtime.parser.nexttoken == 0)
	self.tokentype = LITERAL_INT
	for !self.isAtEnd() {
		// Discard the error, we're checking isAtEnd()
		c, _ := self.peek()
		// We support hex so allow 'x' as a valid part of a number and let
		// the parser detect invalid number formats
		if ( c == '.' ) {
			nc, err := self.peekNext()
			if ( err != nil || !unicode.IsDigit(nc) ) {
				self.runtime.basicError(PARSE, "INVALID FLOATING POINT LITERAL\n")
				self.hasError = true
				return
			}
			self.tokentype = LITERAL_FLOAT
		} else if ( !unicode.IsDigit(c) && c != 'x' ) {
			break
		}
		self.current += 1
	}
	if ( self.tokentype == LITERAL_INT && linenumber == true ) {
		lineno, err := strconv.Atoi(self.getLexeme())
		if ( err != nil ) {
			self.runtime.basicError(PARSE, fmt.Sprintf("INTEGER CONVERSION ON '%s'", self.getLexeme()))
			self.hasError = true
		}
		self.runtime.lineno = int64(lineno)
		self.tokentype = LINE_NUMBER
	}
}

func (self *BasicScanner) matchIdentifier() {
	var identifier string
	self.tokentype = IDENTIFIER
	for !self.isAtEnd() {
		// Discard the error, we're checking isAtEnd()
		c, _ := self.peek()
		if ( unicode.IsDigit(c) || unicode.IsLetter(c) ) {
			self.current += 1
		} else {
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
			}
			break
		}
	}
	identifier = strings.ToUpper(self.getLexeme())

	// Look for reserved words (command and function names) in variable identifiers
	reservedtype, resexists := self.reservedwords[identifier]
	commandtype, cmdexists := self.commands[identifier]
	functiontype, fexists := self.functions[identifier]
	_, ufexists := self.runtime.environment.functions[identifier]
	if ( self.tokentype == IDENTIFIER ) {
		if resexists {
			self.tokentype = reservedtype
		} else if cmdexists {
			self.tokentype = commandtype
		} else if fexists {
			self.tokentype = functiontype
		} else if ufexists {
			self.tokentype = FUNCTION
		}
	} else if ( self.tokentype != IDENTIFIER ) {
		if ( resexists || cmdexists || fexists ) {
			self.runtime.basicError(SYNTAX, "Reserved word in variable name\n")
			self.hasError = true
		}
	}
}

func (self *BasicScanner) scanTokens(line string) string {

	var c rune
	self.line = line
	self.runtime.parser.zero()
	self.current = 0
	self.start = 0
	self.hasError = false
	for !self.isAtEnd() {
		// Discard the error, we're doing our own isAtEnd()
		c, _ = self.advance()
		switch (c) {
		case '^': self.tokentype = CARAT
		case '(': self.tokentype = LEFT_PAREN
		case ')': self.tokentype = RIGHT_PAREN
		case '+': self.tokentype = PLUS
		case '-': self.tokentype = MINUS
		case '/': self.tokentype = LEFT_SLASH
		case '*': self.tokentype = STAR
		case ',': self.tokentype = COMMA
		case '=': self.matchNextChar('=', EQUAL, ASSIGNMENT)
		case '<':
			if ( ! self.matchNextChar('=', LESS_THAN_EQUAL, LESS_THAN) ) {
				self.matchNextChar('>', NOT_EQUAL, LESS_THAN)
			}
		case '>': self.matchNextChar('=', GREATER_THAN_EQUAL, GREATER_THAN)
		case '[': self.tokentype = LEFT_SQUAREBRACKET
		case ']': self.tokentype = RIGHT_SQUAREBRACKET
		case '"':
			self.start = self.current
			self.matchString()
		case '\t': fallthrough
		case ' ':
			self.start = self.current
			break
		case '\r': fallthrough
		case '\n':
			return self.line
		default:
			if ( unicode.IsDigit(c) ) {
				self.matchNumber()
			} else if ( unicode.IsLetter(c) ) {
				self.matchIdentifier()
			} else {
				self.runtime.basicError(PARSE, fmt.Sprintf("UNKNOWN TOKEN %c\n", c))
				self.hasError = true
				self.start = self.current
			}
		}
		if ( self.tokentype != UNDEFINED && self.hasError == false ) {
			switch ( self.tokentype ) {
			case REM: return self.line
			case LINE_NUMBER:
				// We don't keep the line number token, move along
				//fmt.Printf("Shortening line by %d characters\n", self.current)
				self.line = strings.TrimLeft(self.line[self.current:], " ")
				//fmt.Printf("New line : %s\n", self.line)
				self.current = 0
			default:
				self.addToken(self.tokentype, self.getLexeme())
				switch ( self.tokentype ) {
				case LITERAL_STRING:
					// String parsing stops on the final ",
					// move past it.
					self.current += 1
				}
			}
			self.tokentype = UNDEFINED
			self.start = self.current
		}
	}
	return self.line
}
