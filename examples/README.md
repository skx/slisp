# Example Programs

Our [tests/](../tests) directory contains a series of test programs, some of which are adhoc and some of which are more useful.

This directory is designed to contain bigger, or more interesting examples, than those tests.



## Contents

* [args.lisp](args.lisp) - Demonstrate our argument-parsing.
  * Demonstrates the use of the [../packages/arg-parser.lisp](../packages/arg-parser.lisp) package.
* [brainfuck.lisp](brainfuck.lisp) - A brainfuck interpreter
  * Sample brainfuck programs located beneath [bf/](bf/)
  * Run it with no arguments to execute the "Hello world" program.
  * Or pass the path to a script to load and run instead.
* [example.lisp](example.lisp) - Our first example.
* [globals.lisp](globals.lisp) - Explicit demonstration of scopes
  * Shows that local variables always take precedence over global ones.
* [inception.lisp](inception.lisp) - A **lisp interpreter** written in slisp
  * This can load and evaluate named files, and optionally give a REPL mode too.
  * Run "./inception ./inception.in" to see it in operation.
* [life.lisp](life.lisp) - Conway's game of life.
  * Runs in real-time, and will terminate after 100 generations.
  * Run with/without an argument to see the two modes.
* [lisp-reader.lisp](lisp-reader.lisp) - Demonstration of our lisp-reader
  * The [lisp-reader](../packages/lisp-reader.lisp) is what we use to parse lisp expressions in the inception interpreter.
* [nqueens.lisp](nqueens.lisp) - Solver for [The N-queens problem](https://en.wikipedia.org/wiki/Eight_queens_puzzle)
  * Defaults to solving 8x8, but you can give another size as CLI argument.
* [tree.lisp](tree.lisp) - Example of using our AVL-tree package.
  * [../packages/tree.lisp](../packages/tree.lisp) contains the source to the package itself.
* [wc.lisp](wc.lisp) - Clone of the Unix "wc" command.
  * Shows lines, characters, and words for named files.
  * Demonstrates the use of our CLI argument parsing package; [../packages/arg-parser.lisp](../packages/arg-parser.lisp)
* [z-combinator.lisp](z-combinator.lisp) - The less famous cousin to the Y-combinator.
  * Using that we calculate fibonacci and factorial series.
