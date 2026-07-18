;; Ensure that the normal maths are now our variadic versions.
(alias! "+"    "maths:+")
(alias! "-"    "maths:-")
(alias! "*"    "maths:*")
(alias! "/"    "maths:/")

(defun maths:+ (&xs)
  "sum all the numbers in the list"
  (reduce xs (lambda (a b) (sys_plus a b)) 0))

(defun maths:- (first &rest)
  (if (nil? rest)
      (sys_minus 0 first)
      (reduce rest (lambda (a b) (sys_minus a b)) first)))

(defun maths:* (first &rest)
  (reduce rest (lambda (a b) (sys_multiply a b)) first))

(defun maths:/ (first &rest)
  (if (nil? rest)
      (sys_divide 1 (+ first 0.0))
      (reduce rest (lambda (a b) (sys_divide a b)) first)))
