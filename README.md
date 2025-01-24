This BASIC is styled after [Commodore BASIC 7.0](http://www.jbrain.com/pub/cbm/manuals/128/C128PRG.pdf). The scanner, parser and runtime were initially built with the instructions for the Java implementation of Lox in [https://craftinginterpreters.com](https://craftinginterpreters.com), but I got impatient and struck off on my own pretty much as soon as I got commands working.

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

## Expressions

* `+`
* `-`
* `^`
* `*`
* `/`
* `< <= <> == >= >` less than, less than equal, not equal, equal, greater equal, greater than

Expressions can be grouped with `()` arbitrarily deeply. Currently the interpreter has a limit of 32 tokens and leaves per line. In effect this means about 16 operations in a single line.

## Commands (Verbs)

The following commands/verbs are implemented:

* `AUTO n` : Turn automatic line numbering on/off
* `REM` : everything after this is a comment
* `DEF FN(X, ...) = expression` : Define a function with arguments that performs a given expression
* `IF (comparison) THEN (statement) [ELSE (statement)]` : Conditional branching
* `EXIT`: Exit a loop before it would normally finish
* `FOR (assignment) TO (expression) [STEP (expression)]: (statement) : NEXT (variable)`: Iterate over a range of values and perform (statement) or block each time.
* `GOTO n`: Go to line n in the program
* `GOSUB n`: Go to line n in the program and return here when `RETURN` is found
* `LIST [n-n]`: List all or a portion of the lines in the current program
* `PRINT (expression)`
* `QUIT` : Exit the interpreter
* `RETURN` : return from `GOSUB` to the point where it was called
* `RUN`: Run the program currently in memory

## Functions

The following functions are implemented

* `LEN(var$)`: Return the length of the string `var$`
* `MID(var$, start, length)` : Return a substring from `var$`
