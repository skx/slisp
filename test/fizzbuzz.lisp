;; get the fizzbuzz output for the given value.
(defun fizzbuzz (n)
  (cond
    ((= (% n 15) 0) "FizzBuzz")
    ((= (% n 3)  0) "Fizz")
    ((= (% n 5)  0) "Buzz")
    (1                  n)))


;; handle the number
(defun handle_number (xs)
  (if xs
      (do
       (println (fizzbuzz (car xs)))
       (handle_number (cdr xs)))))

(defun main ()
  (handle_number (nat 100)))
