# Brief `slisp` Introduction

`slisp` is a typical toy lisp with support for integers, strings, characters, lambdas, and functions.



## Primitive Types

Primitive types work as you would expect:

* Comments are begun with ";" and continue until the end of the line.
  * There are no block comments.
* Numbers are written as integers:
  * `(print 3)`
* Floating point numbers are not supported, so this is an error:
  * `(print 3.4)`
* Strings are just encoded literally, and escaped characters are honored:
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



## IF

`if` is a standard of lisp, and we support it:

    (if 1
      (print "This is executed")
     (print "This is not"))

Multiple expressions in the "else" branch:

    (if nil
      (print "This is not executed")
     (print "This is executed")
     (print "This is executed too")
     (print "This is also executed")
     (print "This is executed as well ..")
     )

The return value of the expression is the return value of the last executed expression.



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



## Lambdas

Lambdas are functions which can be passed around, and we implement closures to allow counters and adders to be created, etc.

Here's an example of applying a function to each entry in a list:

     ; Create a scope with a list "n" containing numbers 1-20.
     (let ((n (nat 20)))

        ;; Print the results of squaring every item in that list.
        (println (map (lambda (x) (* x x)) n))
     )



## Lists

Lists are internally created as cons-pairs, and you can create such a thing like so:

    ; Manually create the list "1 2 3"
    (cons 1 (cons 2 (cons 3 nil)))

But using the `list` function allows that to be done more neatly:

    ; The same thing
    (list 1 2 3)

(For the common case of creating lists of numbers see the `nat`, `seq` and `range` functions in our [PRIMITIVES.md](PRIMITIVES.md) list.



## See Also

* [README.md](README.md)
  * The main project introduction.
* [PRIMITIVES.md](PRIMITIVES.md)
  * The list of built-in functions, whether implemented in Golang or `slisp` itself.
