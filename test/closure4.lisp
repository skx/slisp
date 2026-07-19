;;
;; This function returns a function that ..
;;
;; Updates a local counter with each argument passed.
;;
;; It's a slightly modified version of closure1.lisp, which
;; just counts the number of times it was called.
;;

(defun count ()
  "Count returns a function which will keep a running total."
  (let ((total 0))
    (lambda (increment)
      (if (nil? increment)                     ; nil arg?
          total                                ; then return the current value
          (set! total (+ total increment)))))) ; otherwise the updated one

(defun main()
  (let ((f (count)))

    ;; Add some numbers to the total
    (println (f  3))
    (println (f  5))
    (println (f 10))
    (println (f 12))

    ;; Calling with "nil" just returns the value
    (println "Running total is now:" (f nil))
    ))
