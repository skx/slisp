;; Print the given item intelligently.
;; We handle integers, strings, nil and cons pairs.
(defun print (x)
  (if (int? x)
      (printint x))
  (if (nil? x)
      (printstr "<nil>"))
  (if (str? x)
      (printstr x))
  (if (cons? x)
      (do
       (putc 40)       ; (
       (print (car x))
       (putc 32)       ; space
       (putc 46)       ; .
       (putc 32)       ; space
       (print (cdr x))
       (putc 41)       ; )
       )))

;; exactly like print, but with a newline on the end.
(defun println (x)
  (print x)
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
        (print "factorial ")
        (print (car xs))
        (print ": ")
        (println (fact (car xs)))
        (factorials (cdr xs)))
      ))

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
  (let ((x (double 13)) (y 12))
    (do

     (if nil
         (do
          (println "BUG: nil should not be true.  Terminating")
          (exit 1)))

     (if 1
         (println "Hello, I am a working `if` statement!"))

      (println "Hello World!")
      (println "I am compiled lisp")

      (println "Some random maths")
      (println (square y))
      (println (double x))

      ;; little countdown to test maths
      (println "Counting down from 10-0")
      (println (+ (* 4 2) 2))
      ;; 9
      (println (- 10 1))
      ;; 8
      (println (+ 6 2))
      ;; 7
      (println (/ 14 2))
      ;; 6
      (println (* 3 2))
      ;; 5
      (println (- 30 (* 5 5)))
      ;; 4
      (println (/ 8 2))
      ;; 3
      (println (+ (* 2 1) 1))
      ;; 2
      (println (- (* 10 10) 98))
      ;; 1
      (println (- 3 2))
      ;; 0
      (println (- 98 98))

      ;; now some factorials.
      (println "Showing results of factorial - 1-10:")
      (factorials (list 1 2 3 4 5 6 7 8 9 10))

      (print "The length of the list of numbers we printed is:" )
      (println (length (list 1 2 3 4 5 6 7 8 9 10)))


      ;; print a pair of characters
      ;(putc 42)
      ;(putc 10)

      (println nil)

      (print "Summing numbers 1..10: ")
      (println (sum (list 1 2 3 4 5 6 7 8 9 10)))

      ; create a cons cell, and print it :)
      (println (cons (cons (cons 12 31) 392) nil))

      (println (cons 1 (cons 2 (cons 3 nil))))
      ;; return value is the last thing compiled.
      0
      ;; You can be more explicit with (exit 0)
      )))
