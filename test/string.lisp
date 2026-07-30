(defun showLen (x)
  (println "The length of : '" x "' is " (strlen x)))

(defun showCmp( a b )
  (println
       "comparing a:" a
       " with b:" b
       " (strcmp a b): " (strcmp a b)
       " (= a b): " (= a b)))

(defun main (args)
  "Test strlen/strcmp"

  ; strlen test
  (showLen "Steve")
  (showLen "")

  ; strcmp test
  (showCmp "Steve" "Steve")
  (showCmp "Steve" "Rteve")

  ; These should be identical even though the addresses will be different
  (showCmp "Hello" (implode (explode "Hello")))
  (showCmp "Hello" (strdup "Hello"))
  (showCmp "Hello" (strcat (strdup "Hell") (strdup "o")))

  ; string conversion
  (print (string "Hello!\n"))
  (print (string #\h))
  (print (string #\e))
  (print (string #\l))
  (print (string #\l))
  (print (string #\o))
  (print (string #\\n))
  (println (string 32))
  (println (string -16386))
  (println (string (- 0 (/ 1.0 3.0))))
  (println (string 2.5))
  (println (string -1.75))
  )
