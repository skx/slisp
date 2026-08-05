(defun main (args)
  "Test quote, quasiquote, and unquote-splicing."

  ; a bare symbol is quoted into the equivalent string.
  (println 'hello)

  ; a quoted list of literals.
  (println '(1 2 3))

  ; a quoted, empty, list is nil.
  (println '())

  ; quasiquote: literal elements are untouched, unquoted elements are
  ; evaluated for real, and unquote-splicing inlines a list.
  (let ((n 2))
    (println `(1 ,(+ n 1) ,@(list 4 5) 6)))

  ; nested lists work fine too.
  (println '((1 2) (3 4)))
)
