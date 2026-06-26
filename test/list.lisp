(defun main ()
  (let ((lst (list 3 4 5 (list 5 2 1) (list 3 12 99))))
    (print "Original list: ")
    (print lst)
    (newline)

    (print "Flatted list: ")
    (print (flatten lst))
    (newline)

    (print "Minimum entry in list: ")
    (print (min (flatten lst)))
    (newline)

    (print "Maximum entry in list: ")
    (print (max (flatten lst)))
    (newline)

  ))
