# slisp

This is either named for "Simple Lisp" or "Steve's lisp", but either way this repository contains a trivial lisp compiler.

Given a lisp-program output the assembly version of that program, which may be compiled/linked/executed.



## Motivation

I've spent a few weeks writing a compiler for a home-made language, [s-lang](https://github.com/skx/s-lang).  Initially that language only used integers, but later I added floats/strings/pointers with appropriate type-markers in the lower bits of the values.

I found the overhead of dealing with typing and syntax a bit complex, and kinda backed myself into a corner with it - I wrote a reasonably complete standard-library with File I/O, getenv, and other things.

The language was complete enough to write a brainfuck interpreter, but unfortunately this was slow (taking two minutes to render the mandelbrot program).  So I ended up writing a JIT assembler to compile brainfuck programs to native code - and that was fast enough to render the mandelbrot example in 3 seconds.

However adding more types, and dynamic things felt like it would be too complex as it would involve ripping out so much of what I'd done.  The compiler, the standard library, and the interface between the two.

So this repository was born:

* Implement a compiler.
* With proper typing from the ground-up.
* Use lisp because the syntax is trivial to parse.
* And I've written interpreters for it in the past so there are dragons, but somewhat friendly ones.



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
