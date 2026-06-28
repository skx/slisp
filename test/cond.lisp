(defun describe_number (n)
  (cond
    ((< n 0) "negative")
    ((= n 0) "zero")
    (1 "positive")))

(defun describe (xs)
  (if xs
      (do
       (print "\tnumber ")
       (print (car xs))
       (print ": ")
       (println (describe_number (car xs)))
        (describe (cdr xs)))))

(defun main ()
  (let ((n (list 0 -1 32 -14 22 101)))
    (print "Showing descriptions of numbers in list:")
    (println n)
    (describe n)))
