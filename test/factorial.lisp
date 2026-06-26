;; factorial.
(defun fact (n)
  (if (<= n 1) 1 (* n (fact (- n 1)))))

;; print the factorial of every number in the given list
(defun factorials (xs)
  (if xs
      (do
       (print "\tfactorial ")
       (print (car xs))
       (print ": ")
       (println (fact (car xs)))
       (factorials (cdr xs)))))

;;
;; main is the entry-point to our compiled code.
;;
(defun main ()

      (let ((n (list 0 1 2 3 4 5 6 7 8 9 10)))
         (println "Showing results of factorial for each entry in that list:")
         (factorials n)))
