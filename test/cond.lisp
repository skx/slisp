(defun describe_number (n)
  (cond
    ((< n 0) "negative")
    ((= n 0) "zero")
    (1 "positive")))

(defun describe (xs)
  (if xs
      (do
       (println "\tnumber " (car xs) ": " (describe_number (car xs)))
       (describe (cdr xs)))))

(defun main ()
  (let ((n (list 0 -1 32 -14 22 101)))
    (println "Showing descriptions of numbers in list:" n)
    (describe n)))
