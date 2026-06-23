;;; example.lisp -- Some useful and realistic examples of what we can compile.
;;
;; This file is half demonstration, and half a test program for the compiler.
;;
;;

;; Print the given item intelligently.
(defun print (x)
  (if (int? x)
      (printint x))
  (if (nil? x)
      (printstr "<nil>"))
  (if (str? x)
      (printstr x))
  (if (cons? x)
      (do
       (putc 40)      ; "("
       (printcons x)  ; List items, separated by spaces
        (putc 41)     ; ")"
      )))

(defun printcons (x)
  (print (car x))
  (if (nil? (cdr x))
      nil
      (if (cons? (cdr x))
          (do
            (putc 32)
            (printcons (cdr x)))
          (do
            (printstr " . ")
            (print (cdr x))))))

;; exactly like print, but with a newline on the end.
(defun println (x)
  (print x)
  (newline))

;; print every item in a list.
(defun print_list (xs)
  (if xs
      (do
       (print (car xs))
       (print_list (cdr xs)))))

;; print every item in a list, then add a newline.
(defun print_list_ln (xs)
  (print_list xs)
  (newline))


;; Add one to the given argument.
(defun add1 (x)
  (+ x 1))

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
       (print_list_ln (list "factorial " (car xs) ": " (fact (car xs))))
        (factorials (cdr xs)))
      ))

;; Create a new list by calling the given function for every list element
(defun map (fn lst)
  (if (nil? lst)
      nil
      (cons
        (fn (car lst))
        (map fn (cdr lst)))))

;; count the length of the given list
(defun length (xs)
  (if xs
      (+ 1 (length (cdr xs)))
      0))

;; sum all the numbers in the list
(defun sum (xs)
  (if xs
      (+ (car xs)
         (sum (cdr xs)))
      0))

;; Setup a binding for "X" to be a function-result.
;; Setup Y for a literal.
;; Print them both
(defun main ()
  (let (
        (x (double 13))
        (y 12)
        (n (list 1 2 3 4 5 6 7 8 9 10)))

     (println "Hello World! I am a compiled lisp")

     (if nil
         (do
          (println "BUG: nil should not be true.  Terminating")
          (exit 1))
         (println "Hello, I am a working `if` statement!"))

      (println "Some random maths now:")
      (println (square y))
      (println (double x))

      ;; now some factorials, and list operations.
      ;;
      ;; Define the list of numbers (1-10)
      (let ((n (list 1 2 3 4 5 6 7 8 9 10)))
         (println "Showing results of factorial - 1-10:")
         (factorials n)
         (print "Summing those numbers: ")
         (println (sum n))
         (print "The length of the list of numbers we handled: " )
          (println (length n))
          (print "Squaring every item of the list, using map:" )
          (println (map (lambda (x) (square x)) n)))

      ; create a cons cell, and print it :)
      (println (cons (cons (cons 12 31) 392) nil))
      (println (cons 1 (cons 2 (cons 3 nil))))

      (println "Expect 10 from this (named) lambda:")
      (let ((x (lambda (a b) (+ a b))))
        (println (x 3 7)))

      (println "Expect 10 from this (immediate) lambda:")
      (println ( (lambda (a) (+ 3 a)) 7))

      ;; return value is the last thing compiled.
      0
      ;; You can be more explicit with (exit 0)


      ))
