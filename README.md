This BASIC is styled after [Commodore BASIC 7.0](http://www.jbrain.com/pub/cbm/manuals/128/C128PRG.pdf) and the [Dartmouth BASIC from 1964](https://www.dartmouth.edu/basicfifty/basic.html). The scanner, parser and runtime were initially built with the instructions for the Java implementation of Lox in [https://craftinginterpreters.com](https://craftinginterpreters.com), but I got impatient and struck off on my own pretty much as soon as I got commands working.

```
make basic.exe

# To use the interactive REPL
./basic.exe

# To run a basic file from the command line
./basic ./scripts/functions.bas
```

# What Works?

This implementation is significantly more complete than my last stab at a BASIC, in my [piquant bootloader project](https://github.com/akesterson/piquant). This one may actually get finished. If it does, I'll rewrite the piquant bootloader in Rust and move this interpreter in there. It will be a glorious abomination.

## Variables

* `A#` Integer variables
* `A%` Float variables
* `A$` String variables. Strings support addition operations with other types.
* `LET` is supported but optional
* Variables are strongly typed

## Arrays

* `DIM(IDENTIFIER, DIMENSION[, ...])` allows for provisioning of multiple dimensional arrays
* `DIM A$(3)` results in a single dimensional array of strings with 3 elements
* `PRINT A$(2)` accesses the last element in an array and returns it to the verb
* Arrays are strongly typed

## Expressions

* `+`
* `-`
* `^`
* `*` (also works on strings)
* `/`
* `< <= <> == >= >` less than, less than equal, not equal, equal, greater equal, greater than

Expressions can be grouped with `()` arbitrarily deeply. Currently the interpreter has a limit of 32 tokens and leaves per line. In effect this means about 16 operations in a single line.

## Commands (Verbs)

The following commands/verbs are implemented:

* `AUTO n` : Turn automatic line numbering on/off at increments of `n`
* `REM` : everything after this is a comment
* `DEF FN(X, ...) = expression` : Define a function with arguments that performs a given expression
* `IF (comparison) THEN (statement) [ELSE (statement)]` : Conditional branching
* `EXIT`: Exit a loop before it would normally finish
* `FOR` : Iterate over a range of values and perform (statement) or block each time.

```
10 FOR I# = 1 TO 5
20 REM Do some stuff in here
30 NEXT I#

10 FOR I# = 1 TO 5 STEP 2
20 REM Do some stuff here
30 NEXT I#
```

* `GOTO n`: Go to line n in the program
* `GOSUB n`: Go to line n in the program and return here when `RETURN` is found
* `LIST [n-n]`: List all or a portion of the lines in the current program
  * `LIST`: List all lines
  * `LIST n-n`: List lines between `n` and `n` (inclusive)
  * `LIST -n`: List lines from 0 to `n`
  * `LIST n`: List lines from `n` to the end of the program
* `DELETE [n-n]`: Delete some portion of the lines in the current program
  * `DELETE`: Delete ALL lines in the program
  * `DELETE n-n`: List lines between `n` and `n` (inclusive)
  * `DELETE -n`: List lines from 0 to `n`
  * `DELETE n`: Delete lines from `n` to the end of the program
* `POKE ADDRESS, VALUE`: Poke the single byte VALUE (may be an integer literal or an integer variable - only the first 8 bits are used) into the ADDRESS (which may be an integer literal or an integer variable holding a memory address).
* `PRINT (expression)`
* `QUIT` : Exit the interpreter
* `RETURN` : return from `GOSUB` to the point where it was called
* `RUN`: Run the program currently in memory

## Functions

The following functions are implemented

* `ABS(x#|x%)`: Return the absolute value of the float or integer argument
* `ATN(x#|x%)`: Return the arctangent of the float or integer argument. Input and output are in radians.
* `CHR(x#)`: Return the character value of the UTF-8 unicode codepoint in x#. Returns as a string.
* `COS(x#|x%)`: Return the cosine of the float or integer argument. Input and output are in radians.
* `HEX(x#)`: Return the string representation of the integer number in x#
* `INSTR(X$, Y$)`: Return the index of `Y$` within `X$` (-1 if not present)
* `LEN(var$)`: Return the length of the object `var$` (either a string or an array)
* `LEFT(X$, Y#)`: Return the leftmost Y# characters of the string in X$. Y# is clamped to LEN(X$).
* `LOG(X#|X%)`: Return the natural logarithm of X#|X%
* `MID(var$, start, length)` : Return a substring from `var$`
* `MOD(x%, y%)`: Return the modulus of ( x / y). Only works on integers, produces unreliable results with floating points.
* `PEEK(X)`: Return the value of the BYTE at the memory location of integer X and return it as an integer
* `POINTER(X)`: Return the address in memory for the value of the variable identified in X. This is the direct integer, float or string value stored, it is not a reference to a `BasicVariable` or `BasicValue` structure.
* `POINTERVAR(X)` : Return the address in memory of the variable X. This is the address of the internal `BasicVariable` structure, which includes additional metadata about the variable, in addition to the value. For a pointer directly to the value, use `POINTERVAL`.
* `RIGHT(X$, Y#)`: Return the rightmost Y# characters of the string in X$. Y# is clamped to LEN(X$).
* `SGN(X#)`: Returns the sign of X# (-1 for negative, 1 for positive, 0 if 0).
* `SHL(X#, Y#)`: Returns the value of X# shifted left Y# bits
* `SHR(X#, Y#)`: Returns the value of X# shifted right Y# bits
* `SIN(X#|X%)`: Returns the sine of the float or integer argument. Input and output are radians.
* `SPC(X#)`: Returns a string of X# spaces. This is included for compatibility, you can also use `(" " * X)` to multiply strings.
* `STR(X#)`: Returns the string representation of X (string or float).
* `TAN(X#|X%)`: Returns the tangent of the float or integer variable X. Input and output are in radians.
* `VAL(X$)`: Returns the float value of the number in X$
* `XOR(X#, Y#)`: Performs a bitwise exclusive OR on the two integer arguments
