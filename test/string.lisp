(defun pr (xs)
  "Print each element of a list"
  (if xs
      (do
       (print (car xs))
       (pr (cdr xs)))))

(defun showLen (x)
  (pr (list "The length of : '" x "' is " (strlen x)))
  (newline))

(defun showCmp( a b )
  (pr (list
       "comparing a:" a
       " with b:" b
       " (strcmp a b): " (strcmp a b)
       " (= a b): " (= a b)))
  (newline)
  )


(defun main ()
  "Test strlen/strcmp"

  ; strlen test
  (showLen "Steve")
  (showLen "")

  ; strcmp test
  (showCmp "Steve" "Steve")
  (showCmp "Steve" "Rteve")

  ; These should be identical
  (showCmp "Hello" (implode (explode "Hello")))

  ; string conversion
  (print (string "Hello!\n"))
  (print (string #\h))
  (print (string #\e))
  (print (string #\l))
  (print (string #\l))
  (print (string #\o))
  (print (string #\\n))
  ; (println (string 32))
  )
