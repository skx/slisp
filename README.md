# slisp

This repository contains `slisp` a compiler which will read lisp programs as input, and generate standalone assembly representations for Linux/AMD64.

> The project is either named for "Simple Lisp" or "Steve's lisp", take your pick.

Lisp is traditionally interactive, and provides a REPL, but having a compiled version is still useful, and still allows most common lisp-programs to execute.



## Example

```lisp
    ;; factorial.  woo.
    (defun fact (n)
      (if (<= n 1) 1 (* n (fact (- n 1)))))

    ;; entry-point
    (defun main ()
      ;; now some factorials.
      (print "Showing results of factorial")
      (print (fact 1))
      (print (fact 2))
      (print (fact 3))
      (print (fact 4))
      (print (fact 5))
      ;; ..
      (print (fact 10))

      ;; exit code - use "(exit 3)" if you prefer
      0)
```

See [example.lisp](example.lisp) for a genuine/bigger example.

We prepend a standard library of functions, implemented in `slisp` itself of course, to all user programs unless `-stdlib=false` is added to the command line.  That library itself is a useful reference/demonstration of functionality:

* [stdlib.slisp](stdlib.slisp) - Our standard library, written in `slisp` itself.
  * Has a flexible `print` definition.
  * Has `map`, `length` and similar general-purpose functions.

Additionally our [test/](test/) directory contains test-cases which demonstrate specific things.



## Features

* Support for bindings, functions, integers, strings, lambdas, lists, etc.
  * The lambdas have support for closures.
  * Run-time type detection via functions such as `int?`, and `cons?`.
* A rough and ready bump-allocator used for heap-allocated cons-cells.
* Mathematical operations:
  * `+`, `-`, `*`, `/`, and `%`.
* Comparision operations:
  * `=`, `<`, `<=`, `>=`, `>`, and `!` to invert a result.
* Special forms
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

`make clean example` will ensure everything happens correctly for our [example.lisp](example.lisp) file.



## Testing

There are some test programs beneath `test/`.  To compile them all:

```sh
cd test && make
```

To run the tests:

```sh
cd test && make test
```

Finally `make clean` will remove the test artifacts, and compiled programs.



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
