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

      ;; exit code - use "(exit 3)" if you prefer
      0)
```

See [example.lisp](example.lisp) for a genuine/bigger example, including a more complex `print` definition that understands `nil` and cons pairs, as well as applying functions to lists.



## Features

* Support for functions, bindings, and basic mathematical operations.
* A rough and ready bump-allocator for easy heap-allocated cons-cells.
* Support for integers, nil, strings and cons pairs.
  * We have some run-time type detection via functions
    * `(cons? x)` - True if the item is a cons pair.
    * `(int? x)` - True if the item is an int.
    * `(lambda? x)` - True if the item is a lambda.
    * `(nil? x)` - True if the item is nil.
    * `(str? y)` - True if the item is a string.
* Primitives
  * `(exit N)` - Exit with the given status-code.
  * `(printint N)` - prints the given number to STDOUT.
  * `(printstr STR)` - prints the given string to STDOUT.
  * `(newline)` - prints a newline.
  * `(putc 42)` - write the given ASCII character to STDOUT.
* Special forms
  * `(do ..)`
  * `(if ..)`
  * `(lambda ..)`
  * `(let ..)`
  * `(list ..)`
    * This turns `(list 1 2 3 4)` into `(cons 1 (cons 2 (cons 3 (cons 4 nil))))` at parse-time.
    * Our `print` function handles displaying this correctly.  Woo.

Anti-features:

* No closures.
* No "set!" for global variables.
  * You want a named variable?  Use `let`.



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
