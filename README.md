# slisp

This is either named for "Simple Lisp" or "Steve's lisp", but either way this repository contains a trivial lisp compiler.

Given a lisp-program output the assembly version of that program, which may be compiled/linked/executed.



## Example

```lisp
     ;; if the value is a number?  print it as an integer.
     ;; otherwise print it as a string
     (defun print (x)
       (if (int? x)
          (printint x)
         (do
           (printstr x)
           (newline))))

     ;; factorial.  woo.
     (defun fact (n)
       (if (<= n 1) 1 (* n (fact (- n 1)))))

    (defun main ()
      ;; now some factorials.
      (print "Showing results of factorial - 1-20")
      (print (fact 1))
      (print (fact 2))
      (print (fact 3))
      (print (fact 4))
      (print (fact 5))
      (print (fact 6))
      (print (fact 7))
      (print (fact 8))
      (print (fact 9))
      (print (fact 10))

      ;; exit code
      0)
```

See [example.lisp](example.lisp) for a genuine/bigger example.



## Features

* Support for functions, bindings, and most mathematical operations.
* Support for integers and strings.
  * We have `(int? x)` and `(str? y)` to let you do run-time type-detection.
* Primitives
  * `(printint 3)` - prints the given number to STDOUT.
  * `(printstr "Steve")` - prints the given string to STDOUT.
  * `(newline)` - prints a newline.
  * `(putc 42)` - write the given ASCII character to STDOUT.

Anti-features:

* No lists, no lambdas, no closures, and no cons-cells.



## Usage

Build the compiler:

    go build .

Use it to compile and link a program:

    ./slisp example.lisp  > example.s
    nasm -f elf64 example.s
    ld -o example example.o

Finally execute your program:

    ./example

`make test` will ensure everything happens correctly for our [example.lisp](example.lisp) file.
