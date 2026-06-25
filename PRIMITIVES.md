# Primitives

Here is a list of all the primitives which are available to `slisp`.

Note that you might need to consult the source of the standard-library to see further details.  This document is primarily intended as a quick summary, and might lag behind reality at times.



## Symbols / Types

The only notable special symbol is `nil` - the nil value.

Characters are specified via the `#\X` syntax, for escaped characters you just need to add the escape:

* `#\a` -> "a"
* `#\b` -> "b"
* ..
* `#\X` -> "X"
* `#\\n` -> newline
* `#\\t` -> tab

Finally strings are enclosed in "quotes like this", and integers are converted to numbers.



## Special Forms

Special forms are things that are built into the compiler core, and handled specially.

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

Core primitives are implemented in assembly language, and can be found within the `template.tmpl` file.

* Type checking functions:
  * `char?`, `cons?`, `int?`, `lambda?`, `nil?`, and `str?`.
* (Integer) mathematical operations:
  * `%`, `*`, `+`, `-`, and `/`.
* (Integer) comparison operations
  * `<`, `<=`, `>=`, `>`, and `=`.
* Other functions:
  * `car`
    * Return the first item of a list.
  * `cdr`
    * Return all items of the list, except the first.
  * `chr`
    * Return the ASCII character corresponding to the given integer.
  * `cons`
    * Add the element to the start of the given (potentially empty) list.
  * `exit`
    * Terminate execution.
  * `explode`
    * Convert the supplied string to a list of characters.
  * `implode`
    * Convert the given list of characters to a string.
  * `nat`
    * Return the list of natural numbers 1 to N.
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
  * `range`
    * Return a list of numbers between the given start/end, using the specified step-size.
  * `seq`
    * Return a list of numbers from 0 to N.



## Standard Library

The standard library consists of routines, and helpers, which are written in 100% `slisp` itself.

The implementation of these primitives can be found in the file [stdlib.slisp](stdlib.slisp).

* `abs`
  * Return the absolute value of the given integer.  (e.g. 3 -> 3, and -3 -> 3).
* `length`
  * Return the length of the specified list.
* `map`
  * Create a new list by calling the given function over every element of the supplied list.
* `print`
  * Print "anything".
* `println`
  * Print "anything", by invoking `print`, then outputting a newline.
* `reverse`
  * Reverse the contents of the specified list.
* `sum`
  * Sum the values in the given list.



## See Also

* [README.md](README.md)
  * The main project introduction.
* [INTRODUCTION.md](INTRODUCTION.md)
  * A brief introduction to the syntax and options.
