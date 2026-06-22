;; Print the given item intelligently.
(defun print (x)
  (if (int? x)
      (printint x))
  (if (nil? x)
      (printstr "<nil>"))
  (if (str? x)
      (printstr x))
  (if (cons? x)
      (do
       (putc 40)      ; "("
       (printcons x)  ; List items, separated by spaces
        (putc 41)     ; ")"
      )))

(defun printcons (x)
  (print (car x))
  (if (nil? (cdr x))
      nil
      (if (cons? (cdr x))
          (do
            (putc 32)
            (printcons (cdr x)))
          (do
            (printstr " . ")
            (print (cdr x))))))

;; exactly like print, but with a newline on the end.
(defun println (x)
  (print x)
  (newline))

(defun main ()
  (println (cons 1 2))
  (println (cons 1 (cons 2 nil)))
  (println (cons 1 (cons 2 (cons 3 nil))))
  (println (cons 1 (cons 2 3)))
)
