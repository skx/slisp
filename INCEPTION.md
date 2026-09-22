## Inception

Our repository contains `slisp` which is a Lisp compiler, but I thought it might be fun to prove that it is a _real lisp_, and so I implemented a lisp interpreter which can be compiled.

As the lisp interpreter can load and execute its own source code I named it `inception`.

You can build both the compiler, and the interpreter by executing `make` at the root of the repository:

```
make
```

Now you can launch the interpreter like so:

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



## Launching Files

In addition to having a REPL you can also load files (and then optionally have the REPL start).  So here's running the self-contained example that is comprised of top-level functions, without a `(defun main ..)` entry-point:

```
$ ./inception examples/inception.in
Loaded stdlib.lisp in 1761ms.
Loading .. examples/inception.in
..
Address:Somewhere in London
```

And here is loading an existing test file - note that loading a file will **not** automatically execute the `main` function, because it wouldn't know what to pass as the arguments.  So instead we launch with the `--repl` flag, and manually invoke the `main` function:

```
$ ./inception test/closure2.lisp --repl
Loaded stdlib.lisp in 1763ms.
Loading .. test/closure2.lisp

Welcome to lisp in slisp; Inception!
Enter :quit to exit.

Help
====
Help for most core functions is available - e.g. (help print)
Run '(help-all [str])' to see all functions [matching str] and their help-text.
Available functions may be listed with (functions), just those matching a string
via (functions "str").

> (main (list "closure2"))
25
35
5
10
22
40
<nil>
```

> **NOTE**: `(main)` takes an argument which is a list of CLI arguments the binary should have received, including the name of the binary as the first argument.  Here we just pass the name of the binary.

We have a series of test-cases located with `test/`, by default we compile each and execute them to ensure their output matches known-good results.  All the tests are usually run with `make test` however we can also run all the test-cases using inception:

```
cd test/
make test-inception
```

In addition to that we can run each of our example files too, for example [examples/nqueens.lisp](examples/nqueens.lisp):

```
$ ./inception examples/nqueens.lisp  --main
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



## Interpreter Differences

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

* `time examples/example` -> 0.006s
  * `time ./inception examples/example.lisp --main` -> 1.854s
* `time examples/nqueens`  .> 0.043s
  * `time ./inception examples/nqueens.lisp --main` -> 11.545s
* `time examples/brainfuck` -> 0.010s
  * `time ./inception examples/brainfuck.lisp --main` -> 2.323s

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

     self-hosted> (require examples/brainfuck)   ; Using that load brainfuck.lisp
     Loading .. examples/brainfuck.lisp
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

     > (require examples/brainfuck)
     Loading .. examples/brainfuck.lisp
     <nil>
     > (main (list "xx" "examples/bf/hello-world.bf"))
     Hello World!
     107

Either will work and produce the `Hello World!` output we all know and love, although it is slow.  Slower than using the compiled interpreter to run the same program (which would be "`./inception examples/brainfuck.lisp --main`").

> **NOTE** You _might_ need to run `ulimit -s unlimited` to avoid segfaults due to stack exhaustion with the nested inception usage.   Now that I've added tail call optimizations that seems unnecessary, but if stack exhaustion is reached a message will alert you of that fact.
</details>
