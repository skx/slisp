;; if the value is a number?  print it as an integer.
;; otherwise print it as a string
(defun print (x)
  (if (int? x)
      (printint x)
      (do
       (printstr x)
       (newline))))


;; Add one to the argument
(defun add1 (x)
  (+ x 1))

;; double the argument
(defun double (x)
  (* x 2))

;; square the argument
(defun square (x)
  (* x x))

;; factorial.  woo.
(defun fact (n)
  (if (<= n 1) 1 (* n (fact (- n 1)))))

;; Setup a binding for "X" to be a function-result.
;; Setup Y for a literal.
;; Print them both
(defun main ()
  (let ((x (double 13)) (y 12))
    (do
     (if 1
         (print "Hello, I am if!"))

      (print "Hello World!")
      (print "I am compiled lisp.  kinda.")
      (print "Some random maths")
       ;; print 12 * 12 -> 144
      (print (square y))

      ;; print 13 * 2 * 2 -> 52.
      (print (double x))

      ;; little countdown to test maths
      (print "Counting down from 10-0")
      (print (+ (* 4 2) 2))
      ;; 9
      (print (- 10 1))
      ;; 8
      (print (+ 6 2))
      ;; 7
      (print (/ 14 2))
      ;; 6
      (print (* 3 2))
      ;; 5
      (print (- 30 (* 5 5)))
      ;; 4
      (print (/ 8 2))
      ;; 3
      (print (+ (* 2 1) 1))
      ;; 2
      (print (- (* 10 10) 98))
      ;; 1
      (print (- 3 2))
      ;; 0
      (print (- 98 98))

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
      (print (fact 11))
      (print (fact 12))
      (print (fact 13))
      (print (fact 14))
      (print (fact 15))
      (print (fact 16))
      (print (fact 17))
      (print (fact 18))
      (print (fact 19))
      (print (fact 20))

      ;; print a pair of characters
      (putc 42)
      (putc 10)

      ;; return value is the last thing compiled.
      0
      ;; You can be more explicit with (exit 0)
      )))
