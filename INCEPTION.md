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
> (main (list "closure2"))
25
35
5
10
22
40
```

> **NOTE**: `(main)` takes an argument which is a list of CLI arguments the binary should have received, including the name of the binary as the first argument.

We have a series of test-cases located with `test/`, by default we compile each and execute them to ensure their output matches known-good results.  All the tests are usually run with `make test` however we can also run all the test-cases using inception:

```
cd test/
make test-inception
```

In addition to that we can run each of our example files too, for example [examples/nqueens.lisp](examples/nqueens.lisp):

```
$ cd examples/ ; make inception
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

So what are the differences between our _compiler_ and our _interpreter_?  Well in some ways the interpreter is more advanced as it has a real symbol-type, and you can get references to functions using them.  The lambdas/defuns are real standalone objects which are treated largely interchangeably and which you can also print.

The `alias!` function works for user-defined functions, but sadly doesn't allow you to override or change built-in functions, as they are in a different namespace.  This works:

    (defun steve (a b) (println "ADD") (sys_plus a b))
    (alias! + steve)
    (+ 3 4)

But this doesn't work, if it did we'd have a recursion problem too of course:

    (defun x (n) (println "CALLED!") (string n))
    (alias! string x)
    (string "steve")

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
<br>

You can of course use the interpreter to run itself, which provides true inception!  You can then go on to run a third program, using the nested interpreter:

     $ ./inception inception.lisp --repl
     Welcome to lisp in slisp!
     Enter :quit to exit.

     > (main (list "a.out" "--repl"))            ; Start the second copy of inception.
     Loaded stdlib.lisp in 15690ms.

     Welcome to lisp in slisp; Inception!
     Enter :quit to exit.

     self-hosted> (require brainfuck)            ; Using that load brainfuck.lisp
     Loading .. brainfuck.lisp
     <nil>
     self-hosted> (main)                         ; And launch it
     main: too few arguments supplied
     Hello World!
     106
     self-hosted> :quit                          ; Exit the nested interpreter
     <nil>
     > :quit                                     ; Exit the compiled interpreter
     $

You could also try this:

     > (require brainfuck)
     Loading .. brainfuck.lisp
     <nil>
     > (main (list "xx" "bf/hello-world.bf"))
     Hello World!
     107

Either will work and produce the `Hello World!` output we all know and love, although it is slow.  Slower than using the compiled interpreter to run the same program (which would be "`./inception brainfuck.lisp --main`").

> **NOTE** You might need to run `ulimit -s unlimited` to avoid segfaults due to stack exhaustion with the nested inception usage.
</details>
