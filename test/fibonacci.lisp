(defun fibonacci (n)
  (if (<= n 1)
    n
    (+ (fibonacci (- n 1)) (fibonacci (- n 2)))))


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
    (println "Showing results of fibonacci value for each entry in that list:")

    ;; Call the fibonacci function for each element in our list.
    (map (lambda (n) (println "\tfibonacci " n ": " (fibonacci n))) n)))
