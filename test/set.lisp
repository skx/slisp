;; create a binding for "x" [=1]
;; update it, via set!
(defun main()
  (let ((x 1))
     ;; x = 1
     (printint x)
     (newline)

     ;; mutate
     (set! x 42)

     ;; confirm it worked
     (printint x)
     (newline)))
