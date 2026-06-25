(defun showLen (x)
  (print "The length of : '")
  (print x)
  (print "' is ")
  (println (strlen x)))

(defun showCmp( a b )
  (print "comparing a:")
  (print a)
  (print " with b:")
  (print b)
  (print " result: ")
  (println (strcmp a b)))


(defun main ()
  "Test strlen/strcmp"

  ; strlen test
  (showLen "Steve")
  (showLen "")

  ; strcmp test
  (showCmp "Steve" "Steve")
  (showCmp "Steve" "Rteve")

  ; These should be identical
  (showCmp "Hello" (implode (explode "Hello"))))
