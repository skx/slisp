(defun foo (a b &c)
  "Test of variable arguments"
  (println "A:" a " B:" b " C:" c))

(defun main ()
  (foo 1 2)
  (foo 1 2 3)
  (foo 1 2 3 4 5 6 7 8 9 10 11 12 13)
  (foo 1 2 "Steve" "List!")
  (foo 1 2 (list 3 4) "Test")
  )
