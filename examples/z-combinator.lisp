;;
;; The Z-combinator is the less famous cousin of the Y-combinator.
;;
;; This is the only one we can implement because we're non-lazy
;; in our argument evaluation.
;;
;; But this works in both our compiler and our interpreter so I'll
;; take the victory where I can!
;;


;; The real lambdas were the friends we made along the way.
(defun Z (f)
  ((lambda (x)
     (f (lambda (v)
          ((x x) v))))
   (lambda (x)
     (f (lambda (v)
          ((x x) v))))))

(defvar factorial
  (Z
    (lambda (fact)
      (lambda (n)
        (if (= n 0)
            1
            (* n (fact (- n 1))))))))


(defvar fib
  (Z
    (lambda (fib)
      (lambda (n)
        (if (< n 2)
            n
            (+ (fib (- n 1))
               (fib (- n 2))))))))

(defvar tree
  (Z
    (lambda (self)
      (lambda (n)
        (if (= n 0)
            1
            (+ (self (- n 1))
               (self (- n 1))))))))


(defun main (args)
  (println (factorial 1))
  (println (factorial 2))
  (println (factorial 3))
  (println (factorial 4))
  (println (factorial 5))
  (println (factorial 6))
  (println (factorial 7))
  (println (factorial 8))
  (println (factorial 9))
  (println (factorial 10))

  (println (fib 1))
  (println (fib 2))
  (println (fib 3))
  (println (fib 4))
  (println (fib 5))
  (println (fib 6))
  (println (fib 7))
  (println (fib 8))
  (println (fib 9))
  (println (fib 10))

  (println (tree 0))
  (println (tree 1))
  (println (tree 2))
  (println (tree 3))
  (println (tree 4))
  (println (tree 5))
  (println (tree 6))
  (println (tree 7))
  (println (tree 8))
  (println (tree 9))
  (println (tree 10))
  )
