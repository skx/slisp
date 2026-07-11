(defun main ()
  (let ((a 1))                      ; a = 1
    (let ((f (lambda (b) (+ a b)))) ; remember a=1
      (let ((a 0))                  ; this, later, change doesn't matter
        (println (f 5))))))         ;    =>  6
