(defun gcd (m n)
  "Return the greatest common divisor between the two arguments."
  (if (= (% m n) 0) n (gcd n (% m n))))

(defun main()
  "Test the GDD function a little"

  (println (gcd 32 132))   ; 4
  (println (gcd 32 128))   ; 32
  (println (gcd 3132 3228)); 12

  )
