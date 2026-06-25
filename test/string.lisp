(defun showLen (x)
  (print "The length of : '")
  (print x)
  (print "' is ")
  (println (strlen x)))

(defun main ()
  "Test strlen/strcmp"

  (showLen "Steve")
  (showLen "")
  )
