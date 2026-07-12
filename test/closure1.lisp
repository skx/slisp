(defun counter ()
  "Counter returns a function which will return an incrementing number every time it is called."
  (let ((n 0))
    (lambda ()
      (set! n (+ n 1))
      n)))

(defun main()
  (let ((f (counter)))
    (println (f))
    (println (f))
    (println (f))
    (println (f))
    ))
