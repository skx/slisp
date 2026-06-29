# Primitives

Here is a list of all the primitives which are available to `slisp`.

Note that you might need to consult the source of the standard-library to see further details.  This document is primarily intended as a quick summary, and might lag behind reality at times.



## Symbols / Types

The only notable special symbol is `nil` - the nil value.

* Comments are begun with ";" and continue until the end of the line.
  * There are no block comments.
* We only support integers, but they may be written in any base the golang `strconv.ParseInt` function supports:
  * `(print 3)`
  * `(print 0xff)`
  * `(print 0b10101010)`
* Floating point numbers are not supported, so this is an error:
  * `(print 3.4)`
* Strings are just encoded literally, and escaped characters are honored:
  * `(print "Hello, world\n")`
* Characters are written with a `#\` prefix:
  * `(print #\*)`
* Lists are written using parenthesis to group them:
  * `(print (list 1 2 3))`



## Special Forms

Special forms are things that are built into the compiler core, and handled specially.

* `cond`
  * An efficient switch implementation.
* `defun`
  * Define a function.
  * We only allow functions at the top-level, and the function named `main` is both mandatory, and our entry-point.
* `do`
  * Execute each statement in the list.
* `if`
  * Our conditional operation.
* `lambda`
  * Creates a lambda function.
* `let`
  * Create a new scope, with locally bound variables.
* `list`
  * Create a list.
* `set!`
  * Set the value of a variable.



## Core Primitives

Core primitives are implemented in assembly language, and can be found within the file [compiler/template.tmpl](compiler/template.tmpl)

* Type checking functions:
  * `char?`, `cons?`, `int?`, `lambda?`, `nil?`, and `str?`.
* (Integer) mathematical operations:
  * `%`, `*`, `+`, `-`, and `/`.
* (Integer) comparison operations
  * `<`, `<=`, `>=`, `>`, and `=`.
* Other functions implemented in assembly:
  * `car`
    * Return the first item of a list.
  * `cdr`
    * Return all items of the list, except the first.
  * `chr`
    * Return the ASCII character corresponding to the given integer.
  * `cons`
    * Add the element to the start of the given (potentially empty) list.
  * `environment`
    * Return a list of all environmental variables.
  * `exit`
    * Terminate execution.
  * `explode`
    * Convert the supplied string to a list of characters.
  * `fclose`
    * Close the given file-handle, and always return nil.
    * To simplify usage `fclose` will accept a nil-filehandle.
  * `fopen`
    * Open the given filename, for read/write, and return a handle.
  * `fread`
    * Read ALL available data from the given handle.
    * To simplify usage `fread` will accept a nil-filehandle, and return nil.
  * `fwrite`
    * Write the given data, with length, to the open file handle.
    * To simplify usage `fwrite` will accept a nil-filehandle, and return nil.
  * `getc`
    * Read a single character from STDIN, return NIL on failure.
  * `implode`
    * Convert the given list of characters to a string.
  * `newline`
    * Print a newline.
  * `not`
    * If the value is `nil` return 1, otherwise return `nil`.
  * `ord`
    * Return the ASCII code of the specified character.
  * `putc`
    * Print the given character.
  * `printint`
    * Print the specified integer.
  * `printstr`
    * Print the given string.
  * `random`
    * Return a random integer between zero and N.
  * `split`
    * Split a string by the given character, and return a list of "(before after)".  Return nil if the character isn't found.
  * `strcmp`
    * Compare the two given strings, like the C-function this returns zero on equality.
  * `strlen`
    * Return the length of the given string.



## Standard Library

The standard library consists of routines, and helpers, which are written in 100% `slisp` itself.

The implementation of these primitives can be found in the file [stdlib.slisp](stdlib.slisp).

* `abs`
  * Return the absolute value of the given integer.  (e.g. 3 -> 3, and -3 -> 3).
* `append`
  * Append the given value to the specified list.  If the list is empty just return the specified item.
* `even?`
  * Return 1 if the given number is even, nil otherwise.
* `filter`
  * Return a list consisting of all members of the input list for which the given predicate returns non-nil.
* `flatten`
  * Flatten the given list of lists into a single list
* `getenv`
  * Return the value of the given environmental variable, nor NIL if not found.
  * Uses `environment`.
* `length`
  * Return the length of the specified list, or string.
* `map`
  * Create a new list by calling the given function over every element of the supplied list.
* `max`
  * Return the highest integer in the list of numbers provided.
* `min`
  * Return the lowest integer in the list of numbers provided.
* `nat`
  * Return the list of natural numbers 1 to N.
* `neg?`
  * Return true if the number is negative.
* `odd?`
  * Return 1 if the given number is odd, nil otherwise.
* `one?`
  * Return true if the number is one.
* `pos?`
  * Return true if the number is positive.
* `print`
  * Print "anything".
* `println`
  * Print "anything", by invoking `print`, then outputting a newline.
* `range`
  * Return a list of numbers between the given start/end, using the specified step-size.
* `reduce`
  * Reduce combines all elements of a list with a function and accumulator.
* `repeated`
  * Create a list with the given value repeated the specified number of times.
* `repeat`
  * Call the given function N times.
* `reverse`
  * Reverse the contents of the specified list.
* `seq`
  * Return a list of numbers from 0 to N.
* `sum`
  * Sum the values in the given list.
* `zero?`
  * Return true if the number is zero.



## See Also

* [README.md](README.md)
  * The main project introduction.
* [INTRODUCTION.md](INTRODUCTION.md)
  * A brief introduction to the syntax and options.
