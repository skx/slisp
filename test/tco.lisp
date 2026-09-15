(defun loop (n acc)
  (if (<= n 0)
      acc
      (loop (- n 1) (+ acc 1))))

(defun main (args)
  (println "result:" (loop 50000000 0))
  0)
