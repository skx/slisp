(defun makeAdder (n)
  (lambda (x)
    (+ x n)))

(defun main ()
  (let ((ten (makeAdder 10))
        (five (makeAdder 5)))

    ;; Add ten to 15, 25, and -5
    (println (ten 15))
    (println (ten 25))
    (println (ten (- 0 5)))

    ;; Add five to 5, 17 & 35
    (println (five 5))
    (println (five 17))
    (println (five 35))
    ))
