# slisp

This repository contains `slisp` a compiler which will read lisp programs as input, and generate standalone assembly representations for Linux/AMD64.

> The project is either named for "Simple Lisp" or "Steve's lisp", take your pick.

Lisp is traditionally interactive, and provides a REPL, but having a compiled version is still useful, and still allows most common lisp-programs to execute.

Quick links:

* [INTRODUCTION.md](INTRODUCTION.md)
  * Brief high-level overview of the facilities.
* [PRIMITIVES.md](PRIMITIVES.md)
  * Detailed list of all available functions and special-forms.



## Example

```lisp
    ;; factorial.  woo.
    (defun fact (n)
      (if (<= n 1) 1 (* n (fact (- n 1)))))

    ;; entry-point
    (defun main (args)
      (println "Showing some factorials:")
      (println (fact 4))
      (println (fact 5))
      (println (fact 9))
      (println (fact 10))

      ;; exit code - use "(exit 3)" if you prefer
      0)
```

There are several examples beneath our [test/](test/) directory, including:

* [factorial.lisp](test/factorial.lisp)
* [fibonacci.lisp](test/fibonacci.lisp)
* [fizzbuzz.lisp](test/fizzbuzz.lisp)

[example.lisp](example.lisp) has other misc. snippets, and finally [brainfuck.lisp](brainfuck.lisp) contains a useful/working brainfuck interpreter.

> There are a couple of "*.bf" files present in this repository, which are brainfuck programs for the interpreter.

It should be noted that we prepend a standard library of functions to all user programs unless `-stdlib=false` is added to the command line.  That library itself is a useful reference/demonstration of functionality:

* [stdlib.slisp](stdlib.slisp) - Our standard library, written in `slisp` itself.
  * Has a good `print` definition which handles known types appropriately.
  * Has `map`, `length` and similar general-purpose functions.



## Features

* Support for bindings, functions, integers, strings, lambdas, lists, etc.
  * The lambdas have support for closures.
  * Run-time type detection via functions such as `int?`, and `cons?`.
* A rough and ready bump-allocator used for heap-allocated cons-cells.
* Mathematical operations:
  * `+`, `-`, `*`, `/`, and `%`.
* Comparison operations:
  * `=`, `<`, `<=`, `>=`, `>`, and `!` to invert a result.
* Special forms
  * `(cond ..)`
  * `(defun ..)`
  * `(do ..)`
  * `(if ..)`
  * `(lambda ..)`
  * `(let ..)`
  * `(list ..)`
  * `(set! ..)`

You can see a complete list of our primitives, and their details in [PRIMITIVES.md](PRIMITIVES.md) - documenting both the built-in special-forms, and the parts of the standard library which are implemented in assembly, or `slisp` itself.

Anti-features:

* No garbage collection.
* No macros.
  * It wouldn't be impossible to add them, but without `quote`, `quasiquote`, etc, it's a lot of work.
* No `quote`
  * Only really useful if you can call `eval` and as a compiler?  That's not going to happen easily.



## Usage

Build the compiler:

    go build .

Use it to compile and link a program:

    ./slisp example.lisp  > example.s
    nasm -f elf64 example.s
    ld -o example example.o

Finally execute your program:

    ./example

**ProTip** Any `*.lisp` file in the current directory will be compiled if you run:

    make clean all

This avoids the need to manually redirect, asssemble, or link.  It will also run the [example.lisp](example.lisp) file - though just "make clean example" will do that too, for neatness.



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
