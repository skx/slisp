# Example Programs

Our [tests/](../tests) directory contains a series of test programs, some of which are adhoc and some of which are more useful.

This directory is designed to contain bigger, or more interesting examples, than those tests.



## Contents

* [brainfuck.lisp](brainfuck.lisp) - A brainfuck interpreter
  * Sample brainfuck programs located beneath [bf/](bf/)
  * Run it with no arguments to execute the "Hello world" program.
  * Or pass the path to a script to load and run instead.
* [example.lisp](example.lisp) - Our first example.
* [globals.lisp](globals.lisp) - Explicit demonstration of scopes
  * Shows that local variables always take precedence over global ones.
* [nqueens.lisp](nqueens.lisp) - Solver for [The N-queens problem](https://en.wikipedia.org/wiki/Eight_queens_puzzle)
  * Defaults to solving 8x8, but you can give another size as CLI argument.
* [packages.lisp](packages.lisp) - Demonstrate the use of our `(package ..)` special form
  * How to declare a package.
  * How it works.
  * How to refer to functions/globals in another package, via qualifiers.
