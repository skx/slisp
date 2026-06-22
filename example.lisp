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

(defun println (x)
  (print x)
  (newline))

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
         (println "Hello, I am if!"))

      (println "Hello World!")
      (println "I am compiled lisp.  kinda.")
      (println "Some random maths")
       ;; print 12 * 12 -> 144
      (println (square y))

      ;; print 13 * 2 * 2 -> 52.
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
      (println "Showing results of factorial - 1-20")
      (println (fact 1))
      (println (fact 2))
      (println (fact 3))
      (println (fact 4))
      (println (fact 5))
      (println (fact 6))
      (println (fact 7))
      (println (fact 8))
      (println (fact 9))
      (println (fact 10))
      (println (fact 11))
      (println (fact 12))
      (println (fact 13))
      (println (fact 14))
      (println (fact 15))
      (println (fact 16))
      (println (fact 17))
      (println (fact 18))
      (println (fact 19))
      (println (fact 20))

      ;; print a pair of characters
      ;(putc 42)
      ;(putc 10)

      (println nil)

      ; create a cons cell, and print it :)
      (println (cons (cons (cons 12 31) 392) nil))

      (println (cons 1 (cons 2 (cons 3 nil))))
      ;; return value is the last thing compiled.
      0
      ;; You can be more explicit with (exit 0)
      )))
