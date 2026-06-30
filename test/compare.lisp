;; Test the comparison operators.
;; They return 1 (int) or NIL depending on whether they are true or not
(defun main ()
      (print "(< 3 4):")
      (print (< 3 4))
      (print " (< 4 4):")
      (print (< 4 4))
      (newline)

      (print "(> 3 4):")
      (print (> 3 4))
      (print " (> 34 4):")
      (print (> 34 4))
      (newline)

      (print "(>= 3 4):")
      (print (>= 3 4))
      (print " (>= 4 4):")
      (print (>= 4 4))
      (print " (>= 34 4):")
      (print (>= 34 4))
      (newline)

      ; floats
      (print "(>= 33.3 4):")
      (print (>= 33.3 4))
      (print " (>= 4 6.5):")
      (print (>= 4 6.5))
      (print " (>= 3.25 3.0):")
      (print (>= 3.25 3.0))
      (newline)

      (print "(<= 3 4):")
      (print (<= 3 4))
      (print " (<= 4 4):")
      (print (<= 4 4))
      (print " (<= 34 4):")
      (print (<= 34 4))
      (newline)

      ; floats
      (print "(<= 33.3 4):")
      (print (<= 33.3 4))
      (print "(<= 3.75 4):")
      (print (<= 3.75 4))
      (print " (<= 4.5 6.5):")
      (print (<= 4.5 6.5))
      (print " (<= 3.25 3.0):")
      (print (<= 3.25 3.0))
      (newline)

      (print "(= 3 3): ")
      (print (= 3 3))
      (print " (= 3 4): ")
      (print (= 3 4))
      (newline)

      (let ((a 3) (b 7))
        (print "(= a a): ")
        (print (= a a))
        (print " (= a b): ")
        (print (= a b)))
      (newline)



)
