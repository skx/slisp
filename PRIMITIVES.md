# Primitives

Here is a list of all the primitives which are available to `slisp`.

Note that you might need to consult the source code to more details.  This document is primarily intended as a quick summary, and might lag behind reality at times.

* The standard library, written in slisp itself:
  * [stdlib.slisp](stdlib.slisp)
* The low-level primitives, written in Linux/AM64 assembly language:
  * [compiler/template.tmpl](compiler/template.tmpl)



## Symbols / Types

The only notable special symbols are `nil`, which is synonymous with false and the empty list, and `t` which is a true value.

We don't have symbols as a specific type, but anything prefixed with a ":" will be silently converted into a string, with the colon removed.  This is designed for visual clarity in code relating to alists, or plists.

* Comments are begun with ";" and continue until the end of the line.  (There are no multi-line/block comments.)
* We support both integers and floating point numbers for mathematical operations.
* Integers may be written in any base the golang `strconv.ParseInt` function supports:
  * `(print 3)`
  * `(print 0xff)`
  * `(print 0b10101010)`
* Floating point numbers are only supported literally, in base10.  (For example `(print 3.4)`)
* We don't have a boolean type, but `nil` (or the empty list) is false.
  * Anything else is true, and we have a `t` symbol for when you want to show that explicitly.
* Strings are encoded literally, and escaped characters are honored:
  * `(print "I say \"Hello, world\".\n")` has embedded quotes and a trailing newline, as you would expect.
* Character-literals are specified with a `#\` prefix:
  * `(print #\*)`
* Lists are written using parenthesis to group them:
  * `(print (list 1 2 3))`
* The only native data structures we support is a list.
  * But `alists` and `plists` are implemented in our standard library, and are documented below.
  * But alists and plists are implemented in our standard library, and are documented in [INTRODUCTION.md](INTRODUCTION.md).

We don't have "symbols" exposed to the language, but if you prefix a variable with "`:`" it will become visually distinct, and this is useful when working with alists, or plists.  Internally that is actually translated to a stringified version of the variable name (So `(print :name)` becomes `(print "name")` - that might seem weird but it works for alist/plist usage, etc.)



## Special Forms

Special forms are things that are built into the compiler core, and handled specially.

* `alias!`
  * Remap functions.  See [INTRODUCTION.md](INTRODUCTION.md) for details.
* `cond`
  * An efficient switch implementation.
* `defconst`
  * Define an immutable global variable.
* `defun`
  * Define a function - The function named `main` is both mandatory, and the entry-point to user-scripts.
* `defvar`
  * Define a global variable.
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
* `require`
  * Load a new package, inline.  See [INTRODUCTION.md](INTRODUCTION.md) for details.
* `set!`
  * Set the value of a variable.
* `unless`
  * Run an unlimited number of expressions when the given condition is false
  * `(unless x (expression1) (expression2) ..)` is the same as `(if x nil (do (expression1) (expression2) ..))`
* `when`
  * Run an unlimited number of expressions when the given condition is true.
  * `(when x (expression1) (expression2) ..)` is the same as `(if x (do (expression1) (expression2) ..))`
* `while`
  * Run the given body for as long as the specified condition is non-nil.



## Core Primitives

Core primitives are implemented in assembly language, and can be found within the file [compiler/template.tmpl](compiler/template.tmpl)

Note that functions have their names mangled a little bit ("!" is converted to "BANG", and
"?" to "QUESTION", for example).  All functions that are user-callable have a "fn_" prefix
added to them too.

* Type checking functions:
  * `char?`, `cons?`, `float?`, `int?`, `lambda?`, `nil?`, and `str?`.
  * The `numeric?` primitive will return true for ints, floats, and characters.
* mathematical operations  `*`, `+`, `-`, and `/`.
  * These work against integers, floating point numbers, or mixed operands.
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
  * `entries`
    * Return the names of all files in the given directory.
    * See [test/entries.lisp](test/entries.lisp) for an example
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
  * `int`
    * Convert the given float, or character, to an integer.
    * Anything else becomes zero.
  * `isqrt`
    * Integer square root, which rounds down.
  * `mkdir`
    * Create the named directory.  **NOTE**: Mode is fixed at 0755, and parent directories must exist unless you use `mkdirs`.
  * `newline`
    * Print a newline.
  * `now`
    * Get the milliseconds past the epoch.
  * `not`
    * If the supplied value is `nil` return 1, otherwise return `nil`.
  * `nth`
    * Return the Nth item from the given list.
  * `nth!`
    * Update the Nth item in the given list with the specified value, in-place.
  * `ord`
    * Return the ASCII code of the specified character.
  * `package`
     * Return the contents of one of our embedded packages at runtime.
  * `packages`
     * Return the names of our embedded packages.
  * `putc`
    * Print the given character.
  * `printfloat`
    * Print the specified floating point number.
  * `printint`
    * Print the specified integer.
  * `printstr`
    * Print the given string.
  * `random`
    * Return a random integer between zero and N.
  * `rmdir`
    * Remove the named directory.
  * `split`
    * Split a string by the given character, and return a list of "(before after)".  Return nil if the character isn't found.
  * `split-all`
    * Return a list of all parts of string, split by the character.
    * e.g. `(split-all (getenv "PATH") #\:)` to find all directories on the PATH.
  * `sqrt`
    * Square root.
  * `stat`
    * Returns file information as a list (TYPE SIZE MODE), or nil on failure.
  * `stdlib`
     * Return the contents of our slisp stdlib at runtime.
  * `strcat`
    * Join two strings together and return them.
  * `strcmp`
    * Compare the two given strings, like the C-function this returns zero on equality.
  * `string`
     * Convert characters, integers, and floats to strings.
     * Everything else returns an empty string.
  * `strlen`
    * Return the length of the given string.
  * `substr`
    * Return a substring from a given string.
  * `sys-gc`
    * Force our garbage collector to run.
  * `sys-heap-allocs`
    * Return the number of memory allocations that have been completed since our last GC.
  * `sys-heap-bytes`
    * Return the size of the heap, in bytes.
  * `sys-heap-objects`
    * Return the number of objects stored within our heap.
  * `sys_run`
    * Helper for command execution.
  * `unlink`
    * Delete the named file.



## Standard Library

The standard library consists of routines, and helpers, which are written in 100% `slisp` itself.

The implementation of these primitives can be found in the file [stdlib.slisp](stdlib.slisp).

* `abs`
  * Return the absolute value of the given integer.  (e.g. 3 -> 3, and -3 -> 3).
* `alist:new`
  * Create a new alist.
* `alist:get`
  * Get an item from an alist.
* `alist:keys`
  * Return all known keys from the given alist.
* `alist:remove`
  * Remove an item, by key, from an alist.
* `alist:set`
  * Add the given key/value to an alist.
* `alist:values`
  * Return all known values from the given alist.
* `and`
  * Test if every item in a list is true.
* `append`
  * Append the given value to the specified list.  If the list is empty just return the specified item.
* `atoi`
  * Convert a string such as `"123"` to an integer.
* `atof`
  * Convert a string such as `"123.25"` to a float.
* `dir?`
  * Does the given path exist as a directory?
* `even?`
  * Return 1 if the given number is even, nil otherwise.
* `every`
  * Does every list item return non-nil when the given function is called?
* `exists?`
  * Does the given filename exist?
* `file?`
  * Does the given path exist as a file?
* `filter`
  * Return a list consisting of all members of the input list for which the given predicate returns non-nil.
* `find`
  * Return the offset of matching items inside the given list.
* `flatten`
  * Flatten the given list of lists into a single list
* `getenv`
  * Return the value of the given environmental variable, nor NIL if not found.
  * Uses `environment`.
* `join`
  * Join all (string) items of a list into a single string.
* `join-by`
  * Join all (string) items of a list into a single string, with the given separator.
* `length`
  * Return the length of the specified list, or string.
* `lower`
  * Return the given string, converted to lower-case.
* `map`
  * Create a new list by calling the given function over every element of the supplied list.
* `max`
  * Return the highest integer in the list of numbers provided.
* `member?`
  * Tests if the given item is present in the specified list.
* `min`
  * Return the lowest integer in the list of numbers provided.
* `mkdirs`
  * Create the given directory, creating any parents as required.  (e.g. "`(mkdirs "foo/bar/baz")`".)
* `nat`
  * Return the list of natural numbers 1 to N.
* `neg?`
  * Return true if the number is negative.
* `odd?`
  * Return 1 if the given number is odd, nil otherwise.
* `one?`
  * Return true if the number is one.
* `or`
  * Is any value in the given list non-nil?
* `plist:new`
  * Create a new property-list
* `plist:get`
  * Get an item from a property-list.
* `plist:keys`
  * Return all known keys from the given plist.
* `plist:remove`
  * Remove an item, by key, from a property-list.
* `plist:set`
  * Set a given key/value in a property-list.
* `plist:values`
  * Return all known values from the given plist.
* `pos?`
  * Return true if the number is positive.
* `print`
  * Print "anything".
* `println`
  * Print "anything" by invoking `print`, then outputting a newline.
* `range`
  * Return a list of numbers between the given start/end, using the specified step-size.
* `read-line`
  * Return a single line of input from STDIN, built on top of `getc`.
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
* `some`
  * Return the first non-nil result of calling the given function against each item of the specified list.
* `strstr`
  * Find a string within another.
* `sum`
  * Sum the values in the given list.
* `system`
  * Run a command via `sys_run`,  return the output on success, and nil on failure.
* `upper`
  * Return the given string, converted to upper-case.
* `which`
  * Find the complete path to the given binary, searching each directory on the $PATH.
* `zero?`
  * Return true if the number is zero.



## See Also

* [README.md](README.md)
  * The main project introduction.
* [INTRODUCTION.md](INTRODUCTION.md)
  * A brief introduction to the syntax and options.
