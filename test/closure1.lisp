(defun counter ()
  (let ((n 0))
    (lambda ()
      (do
        (set! n (+ n 1))
        n))))

(defun main()
  (let ((f (counter)))
    (println (f))
    (println (f))
    (println (f))
    (println (f))
    ))
