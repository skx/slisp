(defun fibonacci (n)
  (if (<= n 1)
    n
    (+ (fibonacci (- n 1)) (fibonacci (- n 2)))))

(defun fibonaccis (xs)
  (if xs
      (do
       (print "\tfibonacci ")
       (print (car xs))
       (print ": ")
       (println (fibonacci (car xs)))
       (fibonaccis (cdr xs)))))

;;
;; main is the entry-point to our compiled code.
;;
(defun main ()

      (let ((n (list 0 1 2 3 4 5 6 7 8 9 10)))
         (println "Showing results of fibonacci value for each entry in that list:")
         (fibonaccis n)))
