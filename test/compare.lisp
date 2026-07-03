;; Test the comparison operators.
;; They return 1 (int) or NIL depending on whether they are true or not
(defun main ()
  ;; <
  (print "(< 3 4):" (< 3 4) " (< 4 4):" (< 4 4))
  (newline)

  ;; >
  (print "(> 3 4):" (> 3 4) " (> 34 4):" (> 34 4))
  (newline)

  ;; >=
  (print "(>= 3 4):" (>= 3 4) " (>= 4 4):" (>= 4 4) " (>= 34 4):" (>= 34 4))
  (newline)
  (print "(>= 33.3 4):" (>= 33.3 4) " (>= 4 6.5):" (>= 4 6.5) " (>= 3.25 3.0):" (>= 3.25 3.0))
  (newline)

  ;; <=
  (print "(<= 3 4):" (<= 3 4) " (<= 4 4):" (<= 4 4) " (<= 34 4):" (<= 34 4))
  (newline)
  (print "(<= 33.3 4):" (<= 33.3 4) "(<= 3.75 4):" (<= 3.75 4) " (<= 4.5 6.5):" (<= 4.5 6.5) " (<= 3.25 3.0):" (<= 3.25 3.0))
  (newline)

  ;; =
  (print "(= 3 3): " (= 3 3) " (= 3 4): " (= 3 4))
  (newline)

  (let ((a 3) (b 7))
    (print "(= a a): " (= a a) " (= a b): " (= a b))
    (newline))
)
