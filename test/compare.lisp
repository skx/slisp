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

          (if (<= 3.9 4)
      (println "OK1"))
    (if (<= 3 4)
      (println "OK2"))
    (if (<= 3 4.3)
      (println "OK3"))

    (if (<= 39 4)
      (println "fail1"))
    (if (<= 34 4)
      (println "fail2"))
    (if (<= 34 .3)
      (println "fail3"))



      (print "(<= 3 4):")
      (print (<= 3 4))
      (print " (<= 4 4):")
      (print (<= 4 4))
      (print " (<= 34 4):")
      (print (<= 34 4))
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
