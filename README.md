# slisp

This repository contains `slisp` (either named for "Steve's Lisp Compiler", or "Simple Lisp Compiler"), which is a compiler reading Lisp programs as input, generating standalone assembly representations for Linux/AMD64 as output.

Lisp is traditionally interactive, and provides a REPL, which this project is not.  That said it's still a great way to execute programs, and a good project for learning (either using a compiled lisp, or implementing one).

Quick links:

* [INTRODUCTION.md](INTRODUCTION.md)
  * Brief high-level overview of the facilities.
* [PRIMITIVES.md](PRIMITIVES.md)
  * Detailed list of all available functions and special-forms.



## Example

This is a minimal, standalone, example of what a program might look like:

```lisp
    (defun fact (n)
      "Calculate, and return, the value of N!"
      (if (<= n 1) 1 (* n (fact (- n 1)))))

    ;; entry-point
    (defun main (args)
      "Command line arguments are available in the list ARGS."
      (println "factorial demonstration, 10!:" (fact 10))

      ;; exit code - use "(exit 0)" if you prefer
      0)
```

You can find bigger examples beneath [examples/](examples/), and our [test/](test/) directory contains a large number of programs which are used for testing purposes (they are compiled and executed, and their output compared to known-good results stored alongside them).

* Notable examples
  * [examples/brainfuck.lisp](examples/brainfuck.lisp) contains a useful/working brainfuck interpreter.
  * [examples/example.lisp](examples/example.lisp) has other misc. snippets.
  * [examples/globals.lisp](examples/globals.lisp) - Explicit demonstration of scopes, showing that local variables always take precedence over global ones.
  * [examples/nqueens.lisp](examples/nqueens.lisp) is a solver for the N Queens problem, defaults to solving the 8x8 grid but you may specify different sizes via a CLI argument.

* Notable tests:
  * [test/entries.lisp](test/entries.lisp) - Read all the files in a directory, filter them, sort them, and print their names.
  * Standard programs: [test/factorial.lisp](test/factorial.lisp), [test/fibonacci.lisp](test/fibonacci.lisp), [test/fizzbuzz.lisp](test/fizzbuzz.lisp).
  * File I/O: [test/fread.lisp](test/fread.lisp) and [test/fwrite.lisp](test/fwrite.lisp).
  * [test/sort3.lisp](test/sort3.lisp) - A mergesort implementation.
  * [test/vararg.lisp](test/vararg.lisp) - Demonstration of a function accepting a variable number of arguments.

It should be noted that we prepend a standard library of functions to all user programs unless `-stdlib=false` is added to the command line.  That library itself is a useful reference/demonstration of functionality:

* [stdlib.slisp](stdlib.slisp) - Our standard library, written in `slisp` itself.
  * Has a good `print` definition which handles known types appropriately.
  * Has `map`, `length` and similar general-purpose functions.



## Features

* Support for bindings, functions, floating-point numbers, integers, strings, lambdas, lists, etc.
  * The lambdas have support for closures.
  * Run-time type detection via functions such as `int?`, and `cons?`.
* A rough and ready bump-allocator used for heap-allocated cons-cells.
* Mathematical operations `+`, `-`, `*`, and `/`.
  * These work against integers, floating point numbers, or combination of the two.
* File I/O operations:
  * `fopen`, `fclose`, `fread`, and `fwrite`.
* Filesystem primitive:
  * `dir?`, `entries`, `exists?`, `file?`, `mkdir`, `mkdirs`, `rmdir`, `stat`, `unlink` and `which`.
* Comparison operations:
  * `=`, `<`, `<=`, `>=`, `>`, and `!` to invert a result.
* Special forms:
  * `(cond ..)`
  * `(defun ..)`
    * `defconst`, `defun`, and `defvar` are the only things that may appear at the top-level of user-scripts.
  * `(defconst ..)`
    * `defconst`, `defun`, and `defvar` are the only things that may appear at the top-level of user-scripts.
  * `(defvar ..)`
    * `defconst`, `defun`, and `defvar` are the only things that may appear at the top-level of user-scripts.
  * `(do ..)`
  * `(if ..)`
  * `(lambda ..)`
  * `(let ..)`
  * `(list ..)`
  * `(set! ..)`
  * `(while ..)`

You can see a complete list of our primitives, and their details in [PRIMITIVES.md](PRIMITIVES.md) - documenting both the built-in special-forms, and the parts of the standard library which are implemented in assembly, or `slisp` itself.

Anti-features:

* No garbage collection.
* No macros.
  * It wouldn't be impossible to add them, but without `quote`, `quasiquote`, etc, it's a lot of work.
* No `quote`
  * Only really useful if you can call `eval` and as a compiler?  That's not going to happen easily.
* We don't have "symbols" exposed to the language, but if you prefix a variable with "`:`" it will become visually distinct, and this is useful when working with alists, or plists.
  * Internally that is actually translated to a stringified version of the variable name (So `(print :name)` becomes `(print "name")` - that might seem weird but it works for alist/plist usage, etc.)



## Usage

Build the compiler:

    go build .

Then use it to compile and link a program:

    ./slisp -compile example.lisp

That will create "example.asm", and "example.o", before creating "example".  If you prefer to run the commands manually you can do it this way:

    ./slisp example.lisp  > example.s
    nasm -f elf64 example.s
    ld -o example example.o

Finally you may execute your compiled program:

    ./example



## Testing

There are some functional test programs beneath [test/](test/), which compile fixed programs and compare their output to known-good results.  You can run these tests by executing:

```sh
cd test && make test
```

Running `make clean` at the top-level will remove the test artifacts, and compiled programs.

In addition to the functional tests there are also golang tests of the internal implementation packages, these can be executed in the standard fashion:

```sh
$ go test ./...
ok      github.com/skx/slisp	0.004s
ok      github.com/skx/slisp/compiler	0.009s
ok      github.com/skx/slisp/env	(cached)
ok      github.com/skx/slisp/lexer	0.008s
ok      github.com/skx/slisp/parser	0.006s
```


There is also support for the fuzz-testing that golang provides, you can run five minutes of fuzz-testing by executing the following (remove the `-fuzztime=300s` to run _forever_, and remove `-parallel=1` to run more than a single instance at a time):

```sh
$ go test -fuzztime=300s -parallel=1 -fuzz=FuzzProject -v
```



## Motivation

I've spent a few weeks writing a compiler for a home-made language, [s-lang](https://github.com/skx/s-lang).  Initially that language only used integers, but later I added floats/strings/pointers with appropriate type-markers in the lower bits of the values.

I found the overhead of dealing with typing and syntax a bit complex, and kinda backed myself into a corner with it - I wrote a reasonably complete standard-library with File I/O, getenv, and other things.

However adding more types, and dynamic things felt like it would be too complex as it would involve ripping out so much of what I'd done.  The compiler, the standard library, and the interface between the two.

So this repository was born:

* Implement a compiler.
* With proper typing from the ground-up.  Using macros for readability and to minimize the chances of making mistakes.
* Use the well-known SysV ABI, rather than my home-grown alternative.
* Use lisp because the syntax is trivial to parse.
  * And I've written interpreters for it in the past so there are dragons, but somewhat friendly ones.

Already this compiler is more "real" and "usable", although it lacks the quality, standard-library, test-cases, and creativity of `s-lang`.  I guess at the end of the day both are toys, and both are here for my own personal learning.
