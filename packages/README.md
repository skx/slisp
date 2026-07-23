# packages

These packages are embedded within our compiler and they may be loaded like so:

    (require foo)

This inserts the contents of "foo.lisp" which will then be parsed by the usual process.  (There is a _little_ more complexity when used dynamically via `inception`, but we can ignore that.)

By convention all "public" functions should be prefixed with the package-name, and a ":".

So you'll see the tree package has functions such as `tree:bound?`, `tree:get`, and `tree:put`.


### alist

Association list code, included as part of `stdlib.lisp`, so there is no need to additionally require it.

```lisp
(require alist)
```


### arg-parser

A simple utility package which allows parsing command-line arguments into "flags" and "other" (named "files" on the basis that is probably what they are).

Usage is demonstrated in the [examples/wc.lisp](examples/wc.lisp) utility program, but in-brief create an instance of the object:

    (require arg-parser)
    (defun main(args)
      (let ((parser (arg-parser:new (cdr args))))
        (print "Flags " (parser :flags))
        (print "Files " (parser :files))))


### lisp-reader

A lisp-reader which converts s-expressions to parsed structure, used by our inception interpreter.  (Located beneath examples/).


### plist

Property list code, included as part of `stdlib.lisp`, so there is no need to additionally require it.

```lisp
(require plist)
```


### tree

A package containing simple AVL-tree routines, used by our inception interpreter.  (Located beneath examples/).

```lisp
(require tree)
```


### variadic-maths

Code that converts mathematical primitives into variadic functions.

```lisp
(require variadic-maths)
```
