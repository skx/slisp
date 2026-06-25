;; little countdown to test maths
(defun main ()
  "Simple mathematical operations, which will do a count-down from 10-0."

  (printint (+ (* 4 2) 2))
  (newline)

  ;; 9
  (printint (- 10 1))
  (newline)

  ;; 8
  (printint (+ 6 2))
  (newline)

  ;; 7
  (printint (/ 14 2))
  (newline)

  ;; 6
  (printint (* 3 2))
  (newline)

  ;; 5
  (printint (- 30 (* 5 5)))
  (newline)

  ;; 4
  (printint (/ 8 2))
  (newline)

  ;; 3
  (printint (+ (* 2 1) 1))
  (newline)

  ;; 2
  (printint (- (* 10 10) 98))
  (newline)

  ;; 1
  (printint (- 0xFF 0b11111110))
  (newline)

  ;; 0
  (printint (- 98 98))
  (newline)

  (let ((n (list 1 2 3 4 5 6 7 8 9 10)))
    (print "List ")
    (print n)
    (newline)

    (print "Even:")
    (print (filter n (lambda (x) (even? x))))
    (newline)

    (print "Odd:")
    (print (filter n (lambda (x) (odd? x))))
    (newline)))
