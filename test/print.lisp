(defun main ()
  "Test printing cons pairs."
  (println (cons 1 2))
  (println (cons 1 (cons 2 nil)))
  (println (cons 1 (cons 2 (cons 3 nil))))
  (println (cons 1 (cons 2 3)))
)
