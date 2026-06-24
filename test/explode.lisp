(defun dumpchars_inner (lst idx)
  (if (nil? lst)
      nil
    (do
     (printstr "character ")
     (printint idx)
      (printstr ": ")
      (putc (car lst))
      (newline)
      (dumpchars_inner (cdr lst) (+ idx 1)))))


(defun dumpchars (str)
  (dumpchars_inner (explode str) 0))



(defun main()

  (let ((str "Hello, world!"))
    ;; Print the characters of the string as a list
    (println (explode str))
    ;; Print them character by character.
    (dumpchars str)
  ))
