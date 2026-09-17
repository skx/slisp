;; factorial.
(defun fact (n)
  (if (<= n 1) 1 (* n (fact (- n 1)))))


;;
;; main is the entry-point to our compiled code.
;;
(defun main (args)

  ;; Create a list of numbers.
  ;;
  ;; This could have been written like this:
  ;;
  ;;     (let ((n (seq 10))) ...
  ;;
  (let ((n (list 0 1 2 3 4 5 6 7 8 9 10)))

    ;; Show what we're doing
    (println "Showing results of factorial for each entry in that list:")

    ;; Call the factorial function for each item in the list.
    (map (lambda (n) (println "\tfactorial " n  ": " (fact n))) n)))
