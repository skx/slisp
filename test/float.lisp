(defun float1 ()
  3.1)

(defun main()
  "Print floating point numbers!"

  ; float1 is a function that returns a float
  (print (float1))
  (print " x 2 = ")
  (println (+ (float1) (float1)))

  ; So that was float+float
  (println "mixed addition:")
  (println (+ 1 (float1)))
  (println (+ (float1) 1))

  (print (float1))
  (print " - ")
  (print (float1))
  (print " = ")
  (println (- (float1) (float1)))

  ; So that was float-float
  (println "mixed subtraction:")
  (println (- (float1) 100 ))
  (println (- 1 (float1)))

  ;; multiplication
  (println "multiplication:")
  (println (* (float1) (float1)))
  (println (* 1 (float1)))
  (println (* (float1) 1))

  ;; divisipn
  (println "division:")
  (println (/ (float1) (float1)))
  (println (/ 1 (float1)))
  (println (/ (float1) 1))

  )
