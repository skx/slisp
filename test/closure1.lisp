;;
;; This function returns a function that ..
;;
;; Updates a local counter each time it is called.
;;
;; It's a slightly modified version of closure4.lisp, which
;; allows a running-total to be maintained.
;;
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
