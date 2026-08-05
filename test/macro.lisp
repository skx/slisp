(defmacro my-if (c then else)
  "A macro version of 'if'."
  `(cond (,c ,then) (t ,else)))

(defmacro my-unless (c &body)
  "A variadic macro."
  `(if ,c nil (do ,@body)))

(defmacro my-or (&args)
  "Expands into a runtime list of the (evaluated) arguments."
  args)

(defun main (args)
  "Test defmacro."

  (println "yes:" (my-if 1 "yes" "no"))
  (println  "no:" (my-if nil "yes" "no"))

  (my-unless nil
    (println "ran-1")
    (println "ran-2"))

  (my-unless 1
    (println "should-not-run"))

  (println (my-or 1 2 3))
)
