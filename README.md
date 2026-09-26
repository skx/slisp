# slisp

This repository contains `slisp` a simple Lisp compiler which generates static/standalone binaries for Linux AMD64 systems.

The repository _also_ contains a lisp _interpreter_, `inception`, which gives you the ability to write code in a lisp REPL interactively.

Quick links:

* [INTRODUCTION.md](INTRODUCTION.md)
  * Brief high-level overview of the facilities.
* [INCEPTION.md](INCEPTION.md)
  * The lisp interpreter which can interpret itself, hence the name.
* [GC](GC.md)
  * Notes on our garbage collector, and how to retrieve stats from it or tweak the behaviour at run-time.
* [PRIMITIVES.md](PRIMITIVES.md)
  * Detailed list of all available functions and special-forms.



## Example

This is a simple example of what a program might look like:

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

You can find bigger examples beneath [examples/](examples/), and our [test/](test/) directory also contains a large number of programs which are used for validation purposes (they are compiled and executed, their output compared to known-good results stored alongside them).

* Notable examples
  * [examples/brainfuck.lisp](examples/brainfuck.lisp) contains a useful/working brainfuck interpreter.
  * [examples/life.lisp](examples/life.lisp) - Game of Life.
  * [examples/nqueens.lisp](examples/nqueens.lisp) is a solver for the N Queens problem, defaults to solving the 8x8 grid but you may specify different sizes via a CLI argument.
  * [examples/wc.lisp](examples/wc.lisp) is a clone of the standard `wc` utility, which demonstrates the argument-parser we include in our internal [packages/](packages/) directory.

* Notable tests:
  * [test/entries.lisp](test/entries.lisp) - Read all the files in a directory, filter them, sort them, and print their names.
  * Standard programs: [test/factorial.lisp](test/factorial.lisp), [test/fibonacci.lisp](test/fibonacci.lisp), & [test/fizzbuzz.lisp](test/fizzbuzz.lisp).
  * File I/O: [test/fread.lisp](test/fread.lisp) and [test/fwrite.lisp](test/fwrite.lisp).
  * [test/sort3.lisp](test/sort3.lisp) - A mergesort implementation.
  * [test/vararg.lisp](test/vararg.lisp) - Demonstration of a function accepting a variable number of arguments.

It should be noted that we prepend a "standard library" of functions to all user programs unless `-stdlib=false` is added to the compiler command line.  That library itself is a useful reference/demonstration of functionality:

* [stdlib.slisp](stdlib.slisp)
  * Our "standard library" written in `slisp` itself.
  * Has `map`, `length` and similar general-purpose functions.



## Features

* Support for bindings, functions, floating-point numbers, integers, strings, lambdas, lists, etc.
  * The lambdas have support for closures.
* Run-time type detection via functions such as `int?`, and `cons?`.
* A rough and ready bump-allocator for memory-allocation.
  * Floats, Lambdas, Lists, and Strings all live on the heap.
  * This is supported by a stop&copy garbage collector, using [Cheney's algorithm](https://en.wikipedia.org/wiki/Cheney%27s_algorithm), named after it's inventor Chris J. Cheney.
  * See the [Garbage Collection](GC.md) file for further details on how the GC process works, and can be inspected/modified.
* Mathematical operations `+`, `-`, `*`, and `/`.
  * These work against integers, floating point numbers, or combination of the two.
* File I/O operations:
  * `fopen`, `fclose`, `fread`, and `fwrite`.
* Filesystem primitive:
  * `dir?`, `entries`, `exists?`, `file?`, `mkdir`, `mkdirs`, `rmdir`, `stat`, `unlink` and `which`.
* Comparison operations:
  * `=`, `<`, `<=`, `>=`, `>`, and `!` to invert a result.
* Special forms (only some of which are valid at the top-level, those are marked with `*` - all forms are valid in our interpreter):
  * `(alias! ..)` - `*` - Alias/overwrite a function.
  * `(defmacro ..)` - `*` - declare a macro.
  * `(defun ..)` - `*` - declare a function.
  * `(defconst ..)` - `*` - declare a global constant.
  * `(defvar ..)`- `*` - declare a global variable.
  * `(do ..)`
  * `(if ..)`
  * `(lambda ..)`
  * `(let ..)`
  * `(require ..)` - `*` - Include other source files.
  * `(set! ..)`
  * `(while ..)`
* Support for _simple_ macros.
  * For example our standard functions `and`, `cond`, `list`, `or`, `unless`, and `when` are implemented as macros.
  * Our interpreter has full macro-support, but the compiler is limited.
* Tail call optimization.
* Support for making arbitrary calls to Linux syscalls, which can be used to implement networking & similar functions.

You can see a complete list of our primitives, and their details in [PRIMITIVES.md](PRIMITIVES.md).  The primitives are grouped by their implementation location (some things are implemented in assembly, and some things are implemented in our own `slisp` language, as part of the embedded [stdlib.slisp](stdlib.slisp).)

**NOTE**: Our lisp _interpreter_, inception, allows all forms at the top-level and has more complete macro support.


### Anti-features

* Macros (`defmacro`) in our compiler are a deliberately restricted, non-hygienic, compile-time expansion mechanism.
  * A macro body may use bound parameters, literals, `quote`/`quasiquote` templates, a compile-time `if`, and `car`/`cdr`/`nil?` (for recursing over a variadic parameter) to construct its expansion - but not arbitrary compile-time computation (e.g. calling `+` directly against a parameter)
  * A macro can't substitute into "raw name" slots - the target of `set!`, `let`-binding names, or `lambda`/`defun` parameter names - since those are parsed as literal tokens, not expressions.
  * So you cannot write a decent `dolist` macro that inserts a named variable in the callee scope, however you can use an anaphoric approach.
* We don't have "symbols" exposed to the language.
  * You may prefix a variable with "`:`" to make it visually distinct.
  * Quoting a bare symbol, e.g. `'foo`, produces the same kind of string.
  * So both `:foo` and `'foo` are treated as the string `"foo"`.



## Usage

Build the compiler and interpreter by running `make`.  If you just want the compiler you can build that in the standard fashion:

    go build .

Using the compiler you can then compile, assemble and link a program like so:

    ./slisp -compile examples/example.lisp

If you prefer to run the commands manually you may do it this way:

    ./slisp examples/example.lisp  > example.s
    nasm -f elf64 example.s
    ld -o example -s --gc-sections example.o

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

I guess at the end of the day both of these languages are toys, and both are here for my own personal learning.
