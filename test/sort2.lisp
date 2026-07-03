;; Same basic idea as the code in sort1.lisp, but instead of
;; hard-coding the comparison operation we pass a lambda.
;;
;; That allows sorting strings via "strcmp", or sorting by
;; reverse by inverting the comparison operation.
;;
(defun insert-by (cmp x xs)
  (if xs
      (if (cmp x (car xs))
          (cons x xs)
          (cons (car xs) (insert-by cmp x (cdr xs))))
      (list x)))

(defun sort-by (cmp xs)
  (if xs
      (insert-by cmp
                 (car xs)
                 (sort-by cmp (cdr xs)))
      nil))


(defun main ()

  ;; Sort numbers
  (let ((input (list  2 489 21 2 39 10 1894 782 21 1 3.2 93.1 1 -32 -3)))
    (print "Before: ")
    (println input)
    (set! input (sort-by (lambda (a b) (< a b)) input))
    (print "Sorted: ")
    (println input))

  ;; Sort words
  (let ((input (split-all "The Quick Brown Fox Jumped Over The Cake" #\ )))
    (print "Before: ")
    (println input)
    (set! input (sort-by (lambda (a b) (> 0 (strcmp a b))) input))
    (print "Sorted: ")
    (println input)))
