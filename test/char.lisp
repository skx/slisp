(defun main()

  ;; "*\n*\n"
  (print #\*)
  (print #\\n)
  (print #\*)
  (print #\\n)

  (let ((a #\a))
    (print "Character: ")
    (println a)
    (print "Number:")
    (print (+ a 1))
    (newline)
    (print "Character again: ")
    (putc (+ a 1))
    (newline))


  (putc (chr 42))
  (putc (chr 32))
  (putc (chr 42))
  (putc #\\n)

  (print "The ASCII for 'x' is:")
  (print (ord #\x))
  (newline)

  (print "The integer 41 is :")
  (print (chr 41))
  (newline)

  )
