(defun dumpchars_inner (lst idx)
  "For each list item print the value of the list item, and it's index"
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
  "Call the internal helper to dump the characters and indexes of each character in the string.

(The explode function returns a list of characters.)"
  (dumpchars_inner (explode str) 0))



(defun main()

  (let ((str "Hello, world!"))
    ;; Print the characters of the string as a list
    (println (explode str))
    ;; Print them character by character.
    (dumpchars str)
  ))
