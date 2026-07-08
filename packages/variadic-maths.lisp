(package maths

         ;; Export these functions as prefixed versions
         (alias! "+"    "maths:+")
         (alias! "-"    "maths:-")
         (alias! "*"    "maths:*")
         (alias! "/"    "maths:/")

         (defun + (&xs)
           "sum all the numbers in the list"
           (reduce xs (lambda (a b) (plus a b)) 0))

         (defun - (first &rest)
           (if (nil? rest)
               (minus 0 first)
               (reduce rest (lambda (a b) (minus a b)) first)))

         (defun * (first &rest)
           (reduce rest (lambda (a b) (multiply a b)) first))

         (defun / (first &rest)
           (if (nil? rest)
               (divide 1 (+ first 0.0))
               (reduce rest (lambda (a b) (divide a b)) first)))

) ;; end of package maths
