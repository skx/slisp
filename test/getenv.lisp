(defun main()
  (println (getenv "FOO"))
  (println (getenv "NAME"))
  (println (getenv "NOT.FOUND!"))

  ;; split on space
  (println (split "Steve Kemp" #\ ))

  ;; split on "="
  (println (split "foo=bar" #\=))

  ;; split on "*" - will not be found -> nil
  (println (split "foo=bar" #\:))
  )
