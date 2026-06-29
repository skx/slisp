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
  )

  (let ((ones (repeated 10 1))
        (me   (repeated 10 "me")))
    (print "Repeated ones: ")
    (println ones)

    (print "Repeated me: ")
    (println me)
  )

  (repeat 5 (lambda (n) (print "I was called by repeat: ") (println n)))

  (println (join (list "This"
                       " is "
                       "a"
                       " test "
                       "of joining"
                       " strings!")))

  ; Add a separator is nice.
  (println (join-by (list "1" "2" "3" "4" "5" "6") ","))

)
