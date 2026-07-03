;;
;; adapted from https://bernsteinbear.com/blog/lisp/16_stdlib/
;;

(defun merge (xs ys)
  "Merge the contents of the two lists, assuming they're both sorted"
  (if (null? xs)
    ys
    (if (null? ys)
      xs
      (if (< (car xs) (car ys))
        (cons (car xs) (merge (cdr xs) ys))
        (cons (car ys) (merge xs (cdr ys)))))))

(defun mergesort (xs)
  "Perform a merge-sort, to sort the contents of the specified list"
  (if (null? xs)
    xs
    (if (null? (cdr xs))
      xs
      (let ((size (length xs))
             (half (/ size 2))
             (first (take half xs))
             (second (drop half xs)))
        (merge (mergesort first) (mergesort second))))))

(defun main ()
  (let ((p (list 3 10 -2 -200 9289 38 321 01 38 112 10 23 5 2 6 7 7 8 9 1)))
    (print "Before sort:")
    (println p)
    (print "After sort:")
    (println (mergesort p))
    ))
