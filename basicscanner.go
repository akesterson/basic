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
	// -------- FUNCTIONS AND OPERATORS ------
	ABS // 23
	AND // 24
	APPEND // 25
	ASC // 26
	ATN // 27
	AUTO // 28
	BACKUP // 29
	BANK // 30
	BEGIN // 31
	BEND // 32
	BLOAD // 33
	BOOT // 34
	BOX // 35
	BSAVE // 36
	BUMP // 37
	CALLFN // 38
	CATALOG // 39
	CHAR // 40
	CHARCIRCLE // 41
	CHR // 42
	CLOSE // 43
	CLR // 44
	CMD // 45
	COLLECT // 46
	COLLISION // 47
	COLOR // 48
	CONCAT // 49
	CONT // 50
	COPY // 51
	COS // 52
	DATA // 53
	DCLEAR // 54
	DCLOSE // 55
	DEFFN // 56
	DELETE // 57
	DIM // 58
	DIRECTORY // 59
	DLOAD // 60
	DO // 61
	DOPEN // 62
	DRAW // 63
	DSAVE // 64
	DVERIFY // 65
	ELSE // 66
	END // 67
	ENVELOPE // 68
	ER // 69
	ERR // 70
	EXIT // 71
	FAST // 72
	FETCH // 73
	FILTER // 74
	FOR // 75
	FRE // 76
	GET // 77
	GETIO // 78
	GETKEY // 79
	GOSUB // 80
	GOTO // 81
	GRAPHIC // 82
	GSHAPE // 83
	HEADER // 84
	HELP // 85
	HEX // 86
	IF // 87
	INPUT // 88
	INPUTIO // 89
	INSTR // 90
	INT // 91
	JOY // 92
	KEY // 93
	LEFT // 94
	LEN // 95
	LET // 96
	LIST // 97
	LOAD // 98
	LOCATE // 99
	LOG // 100
	LOOP // 101
	MID // 102
	MONITOR // 103
	MOVSPR // 104
	NEW // 105
	NEXT // 106
	NOT // 107
	ON // 108
	OPENIO // 109
	OR // 110
	PAINT // 111
	PEEK // 112
	PLAY // 113
	POINTER // 114
	POKE // 115
	POS // 116
	POT // 117
	PRINT // 118
	PRINTIO // 119
	PUDEF // 120
	RCLR // 121
	RDOT // 122
	READ // 123
	RECORDIO // 124
	REM // 125
	RENAME // 126
	RENUMBER // 127
	RESTORE // 128
	RESUME // 129
	RETURN // 130
	RGR // 131
	RIGHT // 132
	RND // 133
	RSPCOLOR // 134
	RSPPOS // 135
	RSPRITE // 136
	RUN // 137
	RWINDOW // 138
	SAVE // 139
	SCALE // 140
	SCNCLR // 141
	SCRATCH // 142
	SGN // 143
	SIN // 144
	SLEEP // 145
	SOUND // 146
	SPC // 147
	SPRCOLOR // 148
	SPRDEF // 149
	SPRITE // 150
	SPRSAV // 151
	SQR // 152
	SSHAPE // 153
	STASH // 154
	STEP // 155
	STOP // 156
	STR // 157
	SWAP // 158
	SYS // 159
	TAB // 160
	TAN // 161
	TEMPO // 162
	THEN // 163
	TI // 164
	TO // 165
	TRAP // 166
	TROFF // 167
	TRON // 168
	UNTIL // 169
	USING // 170
	USR // 171
	VAL // 172
	VERIFY // 173
	VOL // 174
	WAIT // 175
	WHILE // 176
	WIDTH // 177
	WINDOW // 178
	XOR // 179
	COLON // 180 (:)
	EOL // 181 (End of Line)
	EOF // 182 (End of File)
	LINE_NUMBER // 183 (a literal integer that was encountered at the beginning of the line and thus is a line number)
)

type BasicScanner struct {
	current int
	c rune
	start int
	tokentype BasicTokenType
	context *BasicContext
	parser *BasicParser
	line string
	hasError bool
	reservedwords map[string]BasicTokenType
}

func (self *BasicScanner) init(context *BasicContext, parser *BasicParser) error {
	if ( context == nil || parser == nil ) {
		return errors.New("nil pointer argument")
	}
	self.current = 0
	self.start = 0
	self.tokentype = UNDEFINED
	self.context = context
	self.parser = parser
	self.parser.nexttoken = 0
	self.hasError = false
	if len(self.reservedwords) == 0 {
		self.reservedwords = make(map[string]BasicTokenType)
		self.reservedwords["REM"] = REM
		self.reservedwords["AND"] = AND
		self.reservedwords["OR"] = OR
		self.reservedwords["NOT"] = NOT
		self.reservedwords["PRINT"] = PRINT
		self.reservedwords["GOTO"] = GOTO
		self.reservedwords["ABS"] = ABS
		self.reservedwords["APPEND"] = APPEND
		self.reservedwords["ASC"] = ASC
		self.reservedwords["ATN"] = ATN
		self.reservedwords["AUTO"] = AUTO
		self.reservedwords["BACKUP"] = BACKUP
		self.reservedwords["BANK"] = BANK
		self.reservedwords["BEGIN"] = BEGIN
		self.reservedwords["BEND"] = BEND
		self.reservedwords["BLOAD"] = BLOAD
		self.reservedwords["BOOT"] = BOOT
		self.reservedwords["BOX"] = BOX
		self.reservedwords["BSAVE"] = BSAVE
		self.reservedwords["BUMP"] = BUMP
		self.reservedwords["CALLFN"] = CALLFN
		self.reservedwords["CATALOG"] = CATALOG
		self.reservedwords["CHAR"] = CHAR
		self.reservedwords["CHARCIRCLE"] = CHARCIRCLE
		self.reservedwords["CHR"] = CHR
		self.reservedwords["CLOSE"] = CLOSE
		self.reservedwords["CLR"] = CLR
		self.reservedwords["CMD"] = CMD
		self.reservedwords["COLLECT"] = COLLECT
		self.reservedwords["COLLISION"] = COLLISION
		self.reservedwords["COLOR"] = COLOR
		self.reservedwords["CONCAT"] = CONCAT
		self.reservedwords["CONT"] = CONT
		self.reservedwords["COPY"] = COPY
		self.reservedwords["COS"] = COS
		self.reservedwords["DATA"] = DATA
		self.reservedwords["DCLEAR"] = DCLEAR
		self.reservedwords["DCLOSE"] = DCLOSE
		self.reservedwords["DEFFN"] = DEFFN
		self.reservedwords["DELETE"] = DELETE
		self.reservedwords["DIM"] = DIM
		self.reservedwords["DIRECTORY"] = DIRECTORY
		self.reservedwords["DLOAD"] = DLOAD
		self.reservedwords["DO"] = DO
		self.reservedwords["DOPEN"] = DOPEN
		self.reservedwords["DRAW"] = DRAW
		self.reservedwords["DSAVE"] = DSAVE
		self.reservedwords["DVERIFY"] = DVERIFY
		self.reservedwords["ELSE"] = ELSE
		self.reservedwords["END"] = END
		self.reservedwords["ENVELOPE"] = ENVELOPE
		self.reservedwords["ER"] = ER
		self.reservedwords["ERR"] = ERR
		self.reservedwords["EXIT"] = EXIT
		self.reservedwords["FAST"] = FAST
		self.reservedwords["FETCH"] = FETCH
		self.reservedwords["FILTER"] = FILTER
		self.reservedwords["FOR"] = FOR
		self.reservedwords["FRE"] = FRE
		self.reservedwords["GET"] = GET
		self.reservedwords["GETIO"] = GETIO
		self.reservedwords["GETKEY"] = GETKEY
		self.reservedwords["GOSUB"] = GOSUB
		self.reservedwords["GOTO"] = GOTO
		self.reservedwords["GRAPHIC"] = GRAPHIC
		self.reservedwords["GSHAPE"] = GSHAPE
		self.reservedwords["HEADER"] = HEADER
		self.reservedwords["HELP"] = HELP
		self.reservedwords["HEX"] = HEX
		self.reservedwords["IF"] = IF
		self.reservedwords["INPUT"] = INPUT
		self.reservedwords["INPUTIO"] = INPUTIO
		self.reservedwords["INSTR"] = INSTR
		self.reservedwords["INT"] = INT
		self.reservedwords["JOY"] = JOY
		self.reservedwords["KEY"] = KEY
		self.reservedwords["LEFT"] = LEFT
		self.reservedwords["LEN"] = LEN
		self.reservedwords["LET"] = LET
		self.reservedwords["LIST"] = LIST
		self.reservedwords["LOAD"] = LOAD
		self.reservedwords["LOCATE"] = LOCATE
		self.reservedwords["LOG"] = LOG
		self.reservedwords["LOOP"] = LOOP
		self.reservedwords["MID"] = MID
		self.reservedwords["MONITOR"] = MONITOR
		self.reservedwords["MOVSPR"] = MOVSPR
		self.reservedwords["NEW"] = NEW
		self.reservedwords["NEXT"] = NEXT
		self.reservedwords["ON"] = ON
		self.reservedwords["OPENIO"] = OPENIO
		self.reservedwords["PAINT"] = PAINT
		self.reservedwords["PEEK"] = PEEK
		self.reservedwords["PLAY"] = PLAY
		self.reservedwords["POINTER"] = POINTER
		self.reservedwords["POKE"] = POKE
		self.reservedwords["POS"] = POS
		self.reservedwords["POT"] = POT
		self.reservedwords["PRINT"] = PRINT
		self.reservedwords["PRINTIO"] = PRINTIO
		self.reservedwords["PUDEF"] = PUDEF
		self.reservedwords["RCLR"] = RCLR
		self.reservedwords["RDOT"] = RDOT
		self.reservedwords["READ"] = READ
		self.reservedwords["RECORDIO"] = RECORDIO
		self.reservedwords["RENAME"] = RENAME
		self.reservedwords["RENUMBER"] = RENUMBER
		self.reservedwords["RESTORE"] = RESTORE
		self.reservedwords["RESUME"] = RESUME
		self.reservedwords["RETURN"] = RETURN
		self.reservedwords["RGR"] = RGR
		self.reservedwords["RIGHT"] = RIGHT
		self.reservedwords["RND"] = RND
		self.reservedwords["RSPCOLOR"] = RSPCOLOR
		self.reservedwords["RSPPOS"] = RSPPOS
		self.reservedwords["RSPRITE"] = RSPRITE
		self.reservedwords["RUN"] = RUN
		self.reservedwords["RWINDOW"] = RWINDOW
		self.reservedwords["SAVE"] = SAVE
		self.reservedwords["SCALE"] = SCALE
		self.reservedwords["SCNCLR"] = SCNCLR
		self.reservedwords["SCRATCH"] = SCRATCH
		self.reservedwords["SGN"] = SGN
		self.reservedwords["SIN"] = SIN
		self.reservedwords["SLEEP"] = SLEEP
		self.reservedwords["SOUND"] = SOUND
		self.reservedwords["SPC"] = SPC
		self.reservedwords["SPRCOLOR"] = SPRCOLOR
		self.reservedwords["SPRDEF"] = SPRDEF
		self.reservedwords["SPRITE"] = SPRITE
		self.reservedwords["SPRSAV"] = SPRSAV
		self.reservedwords["SQR"] = SQR
		self.reservedwords["SSHAPE"] = SSHAPE
		self.reservedwords["STASH"] = STASH
		self.reservedwords["STEP"] = STEP
		self.reservedwords["STOP"] = STOP
		self.reservedwords["STR"] = STR
		self.reservedwords["SWAP"] = SWAP
		self.reservedwords["SYS"] = SYS
		self.reservedwords["TAB"] = TAB
		self.reservedwords["TAN"] = TAN
		self.reservedwords["TEMPO"] = TEMPO
		self.reservedwords["THEN"] = THEN
		self.reservedwords["TI"] = TI
		self.reservedwords["TO"] = TO
		self.reservedwords["TRAP"] = TRAP
		self.reservedwords["TROFF"] = TROFF
		self.reservedwords["TRON"] = TRON
		self.reservedwords["UNTIL"] = UNTIL
		self.reservedwords["USING"] = USING
		self.reservedwords["USR"] = USR
		self.reservedwords["VAL"] = VAL
		self.reservedwords["VERIFY"] = VERIFY
		self.reservedwords["VOL"] = VOL
		self.reservedwords["WAIT"] = WAIT
		self.reservedwords["WAIT"] = WAIT
		self.reservedwords["WHILE"] = WHILE
		self.reservedwords["WIDTH"] = WIDTH
		self.reservedwords["WINDOW"] = WINDOW
		self.reservedwords["XOR"] = XOR
	}
	return nil
}

func (self *BasicScanner) addToken(token BasicTokenType, lexeme string) {
	self.parser.tokens[self.parser.nexttoken].tokentype = token
	self.parser.tokens[self.parser.nexttoken].lineno = self.context.lineno
	self.parser.tokens[self.parser.nexttoken].lexeme = lexeme
	
	fmt.Printf("%+v\n", self.parser.tokens[self.parser.nexttoken])
	self.parser.nexttoken += 1
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
	var linenumber bool = (self.parser.nexttoken == 0)
	self.tokentype = LITERAL_INT
	for !self.isAtEnd() {
		// Discard the error, we're checking isAtEnd()
		c, _ := self.peek()
		// We support hex so allow 'x' as a valid part of a number and let
		// the parser detect invalid number formats
		if ( c == '.' ) {
			nc, err := self.peekNext()
			if ( err != nil || !unicode.IsDigit(nc) ) {
				basicError(self.context.lineno, PARSE, "INVALID FLOATING POINT LITERAL\n")
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
			basicError(self.context.lineno, PARSE, fmt.Sprintf("INTEGER CONVERSION ON '%s'", self.getLexeme()))
			self.hasError = true
		}
		self.context.lineno = lineno
		self.context.source[self.context.lineno] = self.line
		self.tokentype = LINE_NUMBER
	}
}

func (self *BasicScanner) matchIdentifier() {
	var identifier string
	var reservedIdentifier BasicTokenType
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
	reservedIdentifier = self.reservedwords[identifier]
	// Look for reserved words in variable identifiers
	if ( self.tokentype != IDENTIFIER && reservedIdentifier != UNDEFINED ) {
		basicError(self.context.lineno, SYNTAX, "Reserved word in variable name\n")
		self.hasError = true
		return
	}
}

func (self *BasicScanner) scanTokens(line string) {

	var c rune
	self.line = line
	self.parser.nexttoken = 0
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
		case '=': self.tokentype = EQUAL
		case ':': self.tokentype = COLON
		case '<':
			if ( ! self.matchNextChar('=', LESS_THAN_EQUAL, LESS_THAN) ) {
				self.matchNextChar('>', NOT_EQUAL, LESS_THAN)
			}
		case '>': self.matchNextChar('=', GREATER_THAN_EQUAL, GREATER_THAN)
		case '"':
			self.start = self.current
			self.matchString()
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
				basicError(self.context.lineno, PARSE, fmt.Sprintf("UNKNOWN TOKEN %c\n", c))
				self.hasError = true
				self.start = self.current
			}
		}
		if ( self.tokentype != UNDEFINED && self.hasError == false ) {
			if ( self.tokentype == REM ) {
				return
			} else {
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
}

func (self *BasicScanner) repl(fileobj io.Reader) {
	var readbuff = bufio.NewScanner(fileobj)
	var leaf *BasicASTLeaf = nil
	var err error = nil
	
	fmt.Println("READY")
	for readbuff.Scan() {
		self.scanTokens(readbuff.Text())
		leaf, err = self.parser.parse()
		if ( err != nil ) {
			fmt.Println(fmt.Sprintf("? %s", err))
		}
		if ( leaf != nil ) {
			fmt.Println(fmt.Sprintf("? %s", leaf.toString()))
		}
		fmt.Println("READY")
	}	
}
