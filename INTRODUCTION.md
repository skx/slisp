# Brief `slisp` Introduction

`slisp` is a typical toy lisp with support for floating-point numbers,
integers, strings, characters, lambdas, and functions.



## Primitive Types

Primitive types work as you would expect:

* Comments are begun with ";" and continue until the end of the line.
  * There are no block comments.
* We support integers and floating point numbers for mathematical operations.
* Integers may be written in any base the golang `strconv.ParseInt` function supports:
  * `(print 3)`
  * `(print 0xff)`
  * `(print 0b10101010)`
* Floating point numbers are only supported literally, in base10:
  * `(print 3.4)`
* We don't have a boolean type, but `nil` (or the empty list) is false, and `t` is true.
* Strings are encoded literally, and escaped characters are honored:
  * `(print "Hello, world\n")`
* Characters are written with a `#\` prefix:
  * `(print #\*)`
* Lists are written using parenthesis to group them:
  * `(print (list 1 2 3))`



## Bindings

To start a new scope, with local variables, use `let`:

    (let ((foo "bar")
          (baz  "bart"))
      (print "foo is ")
      (println foo)
      (print "baz is ")
      (println baz))

To update the contents of a bound variable use `set!` which we saw above:

    (set! foo "bar")

So:

    (let ((foo 23))
      (print foo)           ; prints 23.
      (newline)
      (set! foo (* 3 foo))
      (print foo)           ; prints 69.
      (newline))

Note that bindings within the `let` statements can refer to previous bindings, and so this is also valid:

    (let ((x 3)
          (y (* x x)))
       (print y))



## DO

The function `do` allows any number of expressions to be evaluated, and is useful if you want to run multiple expressions in inside one of the branches of an `if` expression, for example.

Any time you want to run multiple expressions but only one is permitted use `do`:

      (do
        (print "I'm the first expression")
        (print "I'm the second expression")
        (print "Multiple expressions can happen here.."))

Our `defun`, `lambda`, and `let` expressions allow an unlimited number of expressions to be executed within their bodies.  Our `if` expression only allows a single expression to be executed, but using `do` you can run more.



## IF

`if` is a standard of lisp, and we support it:

    (if 1
      (print "This is executed")
     (print "This is not"))

The return value of the expression is the return value of the last executed expression.

If you want to run multiple expressions in either the "true" or "false" branch use `do` as seen above.

    (if (not nil)
      (do
        (print "I'm true")
        (print "Multiple expressions can happen here..")
        ))



## Functions

To define a function use `defun`:

    (defun fact (n)
      (if (<= n 1) 1 (* n (fact (- n 1)))))

Optionally you may write some help/usage information in your definition:

    (defun fact (n)
      "Return the factorial of the given number."
      (if (<= n 1) 1 (* n (fact (- n 1)))))

Here's another simple function:

    ;; square the given argument
    (defun square (x)
       (* x x))


There **must** be a function named `main`, as that is the entry-point to the lisp program.  This function can be defined either like so:

    (defun main() ... )

Or like this, if you wish to receive the command-line arguments, supplied as a list:

    (defun main(args) ... )



## Lambdas

Lambdas are functions which can be passed around, and we implement closures to allow counters and adders to be created, etc.

Here's an example of applying a function to each entry in a list:

     ; Create a scope with a list "n" containing numbers 1-20.
     (let ((n (nat 20)))

        ;; Print the results of squaring every item in that list.
        (println (map (lambda (x) (* x x)) n))
     )



## Lists

Lists are internally created as cons-pairs, and you can create such a thing manually like so:

    ; Manually create the list "1 2 3"
    (cons 1 (cons 2 (cons 3 nil)))

But using the `list` function allows that to be done more neatly:

    ; The same thing
    (list 1 2 3)

(For the common case of creating lists of numbers see the `nat`, `seq` and `range` functions in our [PRIMITIVES.md](PRIMITIVES.md) list.)



## Looping

We support the `while` expression to run loops:

    (let ((i 0))
      (while (< i 10)
        (println i)
        (set! i (+ i 1))))

You can see this demonstrated in the [brainfuck.lisp](brainfuck.lisp) program.



## See Also

* [README.md](README.md)
  * The main project introduction.
* [PRIMITIVES.md](PRIMITIVES.md)
  * The list of built-in functions, whether implemented in Golang or `slisp` itself.
