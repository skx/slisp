# Brief `slisp` Introduction

`slisp` is a typical toy lisp with support for floating-point numbers,
integers, strings, characters, lambdas, and functions.



## Primitive Types

* Comments are begun with ";" and continue until the end of the line.
  * There are no block comments.
* We support integers and floating point numbers for mathematical operations.
* Integers may be written in any base the golang `strconv.ParseInt` function supports:
  * `(print 3)`
  * `(print 0xff)`
  * `(print 0b10101010)`
* Floating point numbers are only supported literally, in base10:
  * `(print 3.4)`
* We don't have a boolean type, but `nil` (or the empty list) is false.
  * Anything else is true, and we have a `t` symbol for when you want to show that explicitly.
* Strings are encoded literally, and escaped characters are honored:
  * `(print "Hello, world\n")` has a trailing newline, as you would expect.
* Characters are written with a `#\` prefix:
  * `(print #\*)`
* Lists are written using parenthesis to group them:
  * `(print (list 1 2 3))`

We don't have "symbols" exposed to the language, but if you prefix a variable with "`:`" it will become visually distinct, and this is useful when working with alists, or plists.  Internally that is actually translated to a stringified version of the variable name (So `(print :name)` becomes `(print "name")` - that might seem weird but it works for alist/plist usage, etc.)



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

A function may be defined with the sole/last argument having an `&`-prefix, which means this is a function which will accept a variable number of arguments.  When such functions are called any extra parameters are converted into a list and available in that way.  For example:

    ;; Accept any number of arguments
    (defun foo (&args)
       (println args))

    (foo)                  ; Prints: <nil>
    (foo "Hello" "World")  ; Prints: ("Hello" "World")
    (foo 1 2 3)            ; Prints: (1 2 3)

You can see this demonstrated in the `print` function, in our [stdlib.lisp](stdlib.lisp) file.

**NOTE**: There must be a function named `main`, as that is the entry-point to the compiled program.  This function can be defined either like so:

    (defun main() ... )

Or like this, if you wish to receive the command-line arguments, supplied as a list:

    (defun main(args) ... )



## Global Variables

A global variable may be defined via `defvar`, much like our other bindings there are only two arguments:

    ; Create a global variable
    (defvar version 0.5)

A global variable may be declared as constant, which will cause errors when attempts are made to modify it, for this use `defconst`:

    ; Create a global variable which may not be modified
    (defconst pi 3.14159)

We only allow "`defconst`", "`defun`" and "`defvar`" to appear at the top-level of scripts.



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


### Association Lists (alist)

We only have lists as our main data-type, but using a list we can create something that is hash-like in behavior.  An association list is one way of implementing that, it is a list of individual key-value lists, which can be used to store data.

> This is hash-like because keys are unique, if you set the value of a key `:name` twice the second update removes the previous value.

Imagine you wanted to store details about a person you might use something like this to represent their details:

    ( (name "Bob") (age 42) (location Europe) )

Here's how you might use the functions:

    (let ((a (alist-new)))
       (set! a (alist-set :name   "Steve"))
       (set! a (alist-set :enmail "steve@example.com"))
       (set! a (alist-set :hair   "Red"))

       ;; Do stuff

       (println "Person name " (alist-get a :name)))


### Property Lists (plist)

A property-list, or plist, is a similar way of using a list to store details in a hash-like manner.

> Because if you set the key ":foo" you remove any previous value.

Compared to an alist the list is flat, so an example might look like this:

     ( name "Bob" age 42 location "Europe" )

Here's how you might use the functions:

    (let ((p (plist-new)))
       (set! p (plist-set :name   "Steve"))
       (set! p (plist-set :enmail "steve@example.com"))
       (set! p (plist-set :hair   "Red"))

       ;; Do stuff

       (println "Person name " (plist-get p :name)))



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
