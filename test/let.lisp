;; Test nested-scopes behave properly
(defun main ()

  ; confirm that new scopes don't overwrite
  ; nested scopes due to offset-reuse.
  (let ((program (explode "++")))
    (println program)

    (let ((x 123))
      (println program)
      (println x))

    (println program))

  ; now confirm bindings can refer to earlier ones
  (let ((x 3)
        (y (* x x)))
    (println y))
  )
