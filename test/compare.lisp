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
  (print "(= 3 3):" (= 3 3) " (= 3 4):" (= 3 4))                  ; ints
  (newline)
  (print "(= 3.0 3.0):" (= 3.0 3.0) " (= 3.4 3.5):" (= 3.4 3.5))  ; floats
  (newline)
  (print "(= * 42):" (= #\* 42) " (= a a):" (= #\a #\a))          ; chars
  (newline)
  (print "(= * 42):" (= #\* 42.0) " (= 3 3.0):" (= 3 3.0))        ; mixed
  (print " (= 3 56):" (= 3 56) " (= 3.0 3):" (= 3 3))
  (newline)

  ;; Two strings, but not the same interned object
  (let ((a "Steve") (b (implode (list #\S #\t #\e #\v #\e ))))
    (print "(= Steve Steve):" (= a b))
    (print " (= Steve Smith):" (= "Steve" "Smith" ))
    (print " (= Steve 42):" (= "Steve" 42 ))
    (print " (= Steve 42.1):" (= "Steve" 42.1 ))
    (print " (= Steve S):" (= "Steve" #\S ))
    (print " (= nil nil):" (= nil nil))
    (print " (= t t):" (= t t))
    (newline))


  (let ((a 3) (b 7))
    (print "(= a a): " (= a a) " (= a b): " (= a b))
    (newline))
)
