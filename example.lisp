;;; example.lisp -- Some useful and realistic examples of what we can compile.
;;
;; This file is half demonstration, and half a test program for the compiler.
;;


;; print every item in a list.
;;
;; (print ..) is included in stdlib.slisp, and knows how to print
;; integers, nil, strings, cons cells, etc.
;;
(defun print_list (xs)
  (if xs
      (do
       (print (car xs))
       (print_list (cdr xs)))))

;; print every item in a list, then add a newline.
(defun print_list_ln (xs)
  (print_list xs)
  (newline))


;; double the given argument
(defun double (x)
  (* x 2))

;; square the given argument
(defun square (x)
  (* x x))

;; factorial.  woo!
(defun fact (n)
  (if (<= n 1) 1 (* n (fact (- n 1)))))

;; print the factorial of every number in the given list
(defun factorials (xs)
  (if xs
      (do
       (print_list_ln (list "\tfactorial " (car xs) ": " (fact (car xs))))
        (factorials (cdr xs)))
      ))

;;
;; main is the entry-point to our compiled code.
;;
(defun main (args)
  "main is the (mandatory) entry-point to our code.

This must always be defined, and is where execution starts from."

  ;;
  ;; Declare some variables:
  ;;
  ;;   x -> (* 2 13)
  ;;   y -> 12
  ;;
  (let (
        (x (double 13))
        (y 12))

    (println "Hello World! I am a compiled lisp, and I received command line arguments:")
    (println args)

     (if nil
         (do
          (println "BUG: nil should not be true.  Terminating")
          (exit 1))
         (println "Hello, I am a working 'if' statement!"))

      (println "Some random maths now:")
      (println (square y))
      (println (double x))

      ;; now some factorials, and list operations.
      ;;
      ;; Define the list of numbers (1-10)
      (let ((n (list 1 2 3 4 5 6 7 8 9 10)))

         (print "We have a list of numbers: ")
         (println n)

         (print "Reversed: ")
         (println (reverse n))

         (println "Showing results of factorial for each entry in that list:")
         (factorials n)

         (print "Summing those numbers: ")
         (println (sum n))

         (print "The length of the list of numbers we handled: " )
         (println (length n))

         (print "Squaring every item of the list, using map:" )
         (println (map (lambda (x) (square x)) n))

         (print "List time is over now.\n\n"))


      (println "Expect 10 from this (named) lambda:")
      (let ((x (lambda (a b) (+ a b))))
        (println (x 3 7)))

      (println "Expect 10 from this (immediate) lambda:")
      (println ( (lambda (a) (+ 3 a)) 7))

      ; create a cons cell, and print it :)
      (println "Creating some cons cells and printing them")
      (println (cons (cons (cons 12 31) 392) nil))
      (println (cons 1 (cons 2 (cons 3 nil))))

      ;; return value is the last thing compiled.
      0
      ;; You can be more explicit with (exit 0)
      ))
