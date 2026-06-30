(defun main()
  "Print floating point numbers!"

  ; float1 is a function that returns a float
  (print (float1))
  (print " x 2 = ")
  (println (+ (float1) (float1)))

  ; So that was float+float
  (println "mixed maths:")
  (println (+ 1 (float1)))
  (println (+ (float1) 1))
  )
