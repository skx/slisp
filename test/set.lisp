(defun main()
  "Test that set! will update the value of a local binding, as produced by 'let'."

  (let ((x 1))
     ;; x = 1
     (printint x)
     (newline)

     ;; mutate
     (set! x 42)

     ;; confirm it worked
     (printint x)
     (newline)))
