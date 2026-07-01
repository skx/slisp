(defun insert (x xs)
  "Insert the given number into the correct place in the list"
  (if xs
      (if (< x (car xs))
          (cons x xs)
          (cons (car xs)
                (insert x (cdr xs))))
      (list x)))

(defun sort (xs)
  "Sort a list, by inserting each item."
  (if xs
      (insert (car xs)
              (sort (cdr xs)))
      nil))

(defun main ()
  (let ((input (list  2 489 21 2 39 10 1894 782 21 1 3.2 93.1 1 -32 -3)))
    (print "Before: ")
    (println input)
    (set! input (sort input))
    (print "Sorted: ")
    (println input)))
