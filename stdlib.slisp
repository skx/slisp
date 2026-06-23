;;; stdlib.lisp - Lisp standard library, prepended to user programs.


;; print the given item intelligently.
(defun print (x)
  (if (int? x)
      (printint x))
  (if (nil? x)
      (printstr "<nil>"))
  (if (str? x)
      (printstr x))
  (if (lambda? x)
      (printstr "<lambda>"))
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


;; Create a new list by calling the given function for every list element
(defun map (fn lst)
  (if (nil? lst)
      nil
      (cons
        (fn (car lst))
        (map fn (cdr lst)))))

;; count the length of the given list
(defun length (xs)
  (if xs
      (+ 1 (length (cdr xs)))
      0))

;; sum all the numbers in the list
(defun sum (xs)
  (if xs
      (+ (car xs)
         (sum (cdr xs)))
      0))
