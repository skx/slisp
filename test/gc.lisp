(defvar foo (cons 3 3))
(defvar by-five (lambda (x) (* x x)))
(defvar cake (/ 9.0 3.0))
(defvar name "Steve")
(defvar count 34)

(defun make-adder (n)
  (lambda (x) (+ x n)))

(defvar add5 (make-adder 5))

(defun main (args)

  ; global variable that is a string
  (println foo)

  ; global variable that is a lambda
  (let ((x (list 1 2 3 4 5)))
    (println (map by-five x)))

  ; global variable that is a float
  (println cake)

  ; global variable that is a string
  (println name)

  ; global variable that is an integer
  (println count)

  (println (add5 95))

  (sys-gc)
  (sys-gc)
  (sys-gc)

  ; global variable that is a string
  (println foo)

  ; global variable that is a lambda
  (let ((x (list 1 2 3 4 5)))
    (println (map by-five x)))

  ; global variable that is a float
  (println cake)

  ; global variable that is a string
  (println name)

  ; global variable that is an integer
  (println count)

  (println (add5 95))
)
