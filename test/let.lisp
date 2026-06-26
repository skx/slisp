;; Test nested-scopes behave properly
(defun main ()
  (let ((program (explode "++")))
    (println program)

    (let ((x 123))
      (println program))

    (println program)))
