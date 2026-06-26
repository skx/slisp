(defun main ()
  (let ((lst (list 3 4 5 (list 5 2 1) (list 3 12 "Steve"))))
    (print "Original list: ")
    (print lst)
    (newline)

    (print "Flatted list: ")
    (print (flatten lst))
    (newline)))
