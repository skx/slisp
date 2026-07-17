(defun main ()
  ;; explicitly call GC
  (let ((x (nat 1000)))
    (repeat 1000 (lambda (n) (sys-gc)))
    (if (= (length x) 1000)
        (println "After 1000 GC cycles our list is as expected.")
        (println "PANIC!  ERROR!  Purple Alert!"))
    (println x))

  ;; call "environment" which will make a bunch of allocations and
  ;; which will thus indirectly invoke "cons" and trigger GC
  (let ((x (nat 5000))
        (y "TEST"))
    (repeat 5000 (lambda (n) (environment)))
    (if (= 0 (strcmp y "TEST"))
        (println "After 1000 GC cycles our string is as expected.")
        (println "PANIC!  ERROR!  Purple Alert!  String changed: '" y "'"))
    ))
