# slisp

This repository contains `slisp` (either named for "Steve's Lisp Compiler", or "Simple Lisp Compiler"), which is a compiler reading Lisp programs as input, generating standalone assembly representations for Linux/AMD64 as output.

Lisp is traditionally used in an interactive way, via a REPL.  By contrast this repository allows you to turn a lisp program into a compiled executable which will run non-interactively.

> But note that I did write a **Lisp Interpreter**, complete with a REPL, which you can see described below in the [INCEPTION](#inception) section.

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
  * [examples/life.lisp](examples/life.lisp) - Game of Life.
  * [examples/globals.lisp](examples/globals.lisp) - Explicit demonstration of scopes, showing that local variables always take precedence over global ones.
  * [examples/nqueens.lisp](examples/nqueens.lisp) is a solver for the N Queens problem, defaults to solving the 8x8 grid but you may specify different sizes via a CLI argument.
  * [examples/wc.lisp](examples/wc.lisp) is a clone of the standard `wc` utility, which demonstrates our included argument-parser [packages/](package/).

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
* A rough and ready bump-allocator used heap to allocate memory for heap-allocated objects.
  * This is supported by a stop&copy garbage collector, using [Cheney's algorithm](https://en.wikipedia.org/wiki/Cheney%27s_algorithm) (which is named after it's inventor Chris J. Cheney).
  * See the [Garbage Collection](#garbage-collection) section below for details.
* Mathematical operations `+`, `-`, `*`, and `/`.
  * These work against integers, floating point numbers, or combination of the two.
* File I/O operations:
  * `fopen`, `fclose`, `fread`, and `fwrite`.
* Filesystem primitive:
  * `dir?`, `entries`, `exists?`, `file?`, `mkdir`, `mkdirs`, `rmdir`, `stat`, `unlink` and `which`.
* Comparison operations:
  * `=`, `<`, `<=`, `>=`, `>`, and `!` to invert a result.
* Special forms (only some of which are valid at the top-level, those are marked with `*`):
  * `(alias! ..)` - `*` - Alias/overwrite a function.
  * `(cond ..)`
  * `(defun ..)` - `*` - declare a function.
  * `(defconst ..)` - `*` - declare a global constant.
  * `(defvar ..)`- `*` - declare a global variable.
  * `(do ..)`
  * `(if ..)`
  * `(lambda ..)`
  * `(let ..)`
  * `(list ..)`
  * `(require ..)` - `*` - Include other source files.
  * `(set! ..)`
  * `(unless ..)`
  * `(when ..)`
  * `(while ..)`

You can see a complete list of our primitives, and their details in [PRIMITIVES.md](PRIMITIVES.md) - documenting both the built-in special-forms, and the parts of the standard library which are implemented in assembly, or `slisp` itself.

Anti-features:

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

    ./slisp -compile example/example.lisp

That will generate "example.asm", compile it to "example.o", and then invoke the linker to generate the executable "example".  If you prefer to run the commands manually you may do it this way:

    ./slisp example/example.lisp  > example.s
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



## Inception

As noted this is a _compiler_ which means that for a given lisp program we produce an executable, there is no REPL.

But I thought it might be fun to prove that my slisp is a _real lisp_, and so I implemented a lisp interpreter which can read lisp source code from files and execute it, and which also implements a REPL.

Build the compiler, and build the interpreter:

```
go build .
cd examples/
make inception
```

Now you should have the executable `inception` present, which is the lisp-interpreter.  Fire it up:

```
$ ./inception --repl
Welcome to lisp in slisp!
Enter :quit to exit.

> (defun square (x) (* x x))
(symbol square)
> square
(closure (x) (((symbol *) (symbol x) (symbol x))) <nil>)
> (square 3)
9
> (square (square (square 3)))
6561
> :quit
```

In addition to having a REPL you can also load files (and then optionally have the REPL start).  So here's running the self-contained example that is comprised of top-level functions, without a `(defun main ..)` entry-point:

```
$ ./inception inception.in
Loading .. inception.in
100
Squaring some numbers: (16 25 400 900 1600)
LAMBDA X 1*1: -> 1
LAMBDA X 2*2: -> 4
LAMBDA X 3*3: -> 9
LAMBDA X 4*4: -> 16
LAMBDA X 5*5: -> 25
This is what a function looks like: (closure (x) (((symbol +) (symbol x) (symbol n))) ((n 10)))
Adder (+10) result for  5:15
..
```

And here is loading an existing test file.  Loading this file will not immediately run the `main` function, so we add the `--repl` flag to start that up, after loading and parsing the program.  We can then make it run by calling `(main)` ourselves:

```
$ ./inception ../test/closure2.lisp  --repl
Loading .. ../test/closure2.lisp
Welcome to lisp in slisp!
Enter :quit to exit.

; loading "closure2.lisp" will define (defun main)
; now we call it via the REPL:
> (main)
25
35
5
10
22
40
```

Perhaps more impressive is that we can execute the [examples/nqueens.lisp](examples/nqueens.lisp) example, with no changes:

```
$ ./inception nqueens.lisp  --main
Loading .. nqueens.lisp

8 Queens Solver for board size 8x8

Solution 1 (1 5 8 6 3 7 2 4):

    Q . . . . . . .
    . . . . Q . . .
    . . . . . . . Q
..
..
```

> Here you'll see we added `--main` which automatically runs the `(main)` function our examples define.

So what are the differences between our _compiler_ and our _interpreter_?  Well in some ways the interpreter is more advanced as it has support for `(quote)`, it has a symbol-type, and you can get references to functions using them.  The lambdas/defuns are real standalone objects which are treated largely interchangeably and which you can also print.   The biggest difference really is in terms of speed and features:

* The interpreter has 100% of the builtin functions you would expect
  * `cons`, `entries`, `environment`, `fopen`, `map`, etc.
* However it does not load the `stdlib.slisp` file, so if you want the functions that this provides you need to prepend it to your source.
  * `cat stdlib.slisp your-program.lisp > run.me; ./inception ./run.me --main`.
* The interpreter does support all the types the compiler does though.
  * Floating point numbers, integers, strings, and character literals, etc.

The interpreter is obviously much slower than our compiled binaries, due to the overhead of interpreting everything manually.  Sometimes this slowdown is minor, other times it is signification, it really depends upon the nature of the program:

* `time ./example` -> 0.006s
  * `time ./inception example.lisp --main` -> 0.026s
* `time ./nqueens`  .> 0.053s
  * `time ./inception nqueens.lisp --main` -> 20.846s
* `time ./brainfuck` -> 0.010s
  * `time ./inception brainfuck.lisp --main` -> 6.775s

That said, and as demonstrated above, the interpreter can run many of the same programs that the compiler can.

<details>
<summary>To achieve true inception you need to run the interpreter with itself</summary>


You can of course use the interpreter to run itself, which provides true inception!

Because the interpreter doesn't contain the standard library, and it doesn't fully understand the `(require ..)` special form, you need to massage the source slightly:

     $ cat ../stdlib.slisp lisp-reader.lisp tree.lisp inception.lisp >new.txt

Once you do that you can launch the interpreter and tell it to run a second program:

     $ ./inception new.txt --repl
     Welcome to lisp in slisp!
     Enter :quit to exit.

     > (execute-file "brainfuck.lisp")
     Loading .. brainfuck.lisp
     ((symbol main) <nil>)
     > (main)
     Hello World!
     106
     > (exit)

You could also try this:

     > (require brainfuck)
     Loading .. brainfuck.lisp
     <nil>
     > (main (list "xx" "bf/hello-world.bf"))
     Hello World!
     107

Either will work and produce the `Hello World!` output we all know and love, although again it is slow.  Slower than using the compiled interpreter to run the same program (which would be "`./inception brainfuck.lisp --main`").

</details>



## Garbage Collection

When our project started it used a simple bump-allocator.  That just meant reserving a huge contiguous chunk of memory and maintaining a simple count of memory used.  Every request would just advance the "used" pointer by the size requested, and return the previous value.  This was fast and simple, but meant there was no possibility to free memory.

The introduction of the `inception` lisp-interpreter, and to a lesser extent our brainfuck and nqueens programs, really made it apparent that this wasn't tenable.  Large programs would exhaust the available heap with items that were no longer referenced.

To start with I did the obvious thing and made the allocation region larger, ignoring the problem.  But eventually that too became untenable.  So now I've implemented a stop & copy garbage collector, using Cheney's algorithm.   Every time our `(cons ..)` primitive is called we run the GC process if there have been more than 64,000 allocations since the previous garbage-collection ran.

The `(cons ..)` primitive is a lisp-fundamental, so I figure that is going to be called pretty often in user-programs, either directly or via the `(list ...)` wrapper.  But if that isn't the case you may also trigger the garbage-collection process explicitly, and see the stats via these methods:

* `(sys-gc)` run the garbage collection process immediately.
* `(sys-heap-allocs)` -  Return the number of memory allocations made since the last garbage-collection process.
  * If your program regularly calls `cons` this will never be more than 64,000.
* `(sys-heap-bytes)` - Return the size of the heap.
* `(sys-heap-data)` - Return the contents of the heap as a list of entries.
* `(sys-heap-dump)` - Dump a summary of the heap to the console.
  * This is implemented in assembly and literally writes to STDOUT.  There is no control over the formatting.
  * I wanted to write a function that would return a list of heap entries, but that might trigger a GC process.  Which would invalidate the results mid-traversal, so this is my compromise solution.
* `(sys-heap-objects)` - Return the number of objects stored upon the heap.

The stop and copy implementation is pretty simple:

* We have *two* heap areas, each of which are an identical size.
* One heap is used as the backing-store for all allocations we make.
* When a `sys-gc` request is made the current heap is inspected and all live items are copied to the other heap.
  * The new heap is then made the active one, which essentially orphans and frees the unreachable entries upon the old heap.
* The copying process has to deal with global variables, objects held within stack-frames, and those objects which might be held inside registers.
  * For register contents we cheat a little.
  * The `(cons ..)` primitive is the only one that is used to trigger "auto GC",  and we know `cons` can only be called with two arguments, so we only have to consider the two registers RDI & RSI.
  * TLDR; Our roots are "globals", "stack-locals", and potentially the contents of the two registers `rdi` and `rsi`.



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
