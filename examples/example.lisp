;;; example.lisp -- Some useful and realistic examples of what we can compile.
;;
;; This file is half demonstration, and half a test program for the compiler.
;;


;; double the given argument
(defun double (x)
  (* x 2))

;; square the given argument
(defun square (x)
  (* x x))

;; factorial.  woo!
(defun fact (n)
  (if (<= n 1) 1 (* n (fact (- n 1)))))

;; print the factorial of every number in the given list, via map.
(defun factorials (xs)
  (map (lambda (n) (println "\tfactorial " n ": " (fact n))) xs))

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
      (println "\t" (square y))
      (println "\t" (double x))

      ;; now some factorials, and list operations.
      ;;
      ;; Define the list of numbers (1-10)
      (let ((n (list 1 2 3 4 5 6 7 8 9 10)))

         (println "We have a list of numbers: " n)

         (println "Reversed: " (reverse n))

         (println "Showing results of factorial for each entry in that list:")
         (factorials n)

         (println "Summing those numbers: " (sum n))

         (println "The length of the list of numbers we handled: " (length n))

         (print "Squaring every item of the list, using map:" )
         (println (map (lambda (x) (square x)) n))

         (print "List time is over now.\n\n"))


      (print "Expect 10 from this (named) lambda: ")
      (let ((x (lambda (a b) (+ a b))))
        (println (x 3 7)))

      (print "Expect 10 from this (immediate) lambda: ")
      (println ( (lambda (a) (/ 100 a)) 10))

      ; create a cons cell, and print it :)
      (println "Creating some cons cells and printing them")
      (println "\t" (cons (cons (cons 12 31) 392) nil))
      (println "\t" (cons 1 (cons 2 (cons 3 nil))))

      ;; return value is the last thing compiled.
      ;; You can be more explicit with (exit 0)
      (exit 0)
      ))
