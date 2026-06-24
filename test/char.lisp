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
    (newline)))
