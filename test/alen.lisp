(defun main(main)
  (let ((a  (strdup "slisp")) ; heap-allocated string
        (b  "slisp"))         ; static string

    ;; The string lengths will be identical
    (println "heap-allocated str is " (strlen a) " bytes")
    (println "static str is "         (strlen b) " bytes")

    ;; But the size of the "allocation" will be differ
    (let ((aLen (alen a))
          (bLen (alen b)))
      (if (<= bLen aLen)
          (println "OK, static strings are smaller than heap ones")))

    ;; heap-allocated things are rounded up to multiples of 16
    ;; static ones are not, but they have a fake header to allow
    ;; us to pretend we're there
    ;;
    ;; In our compiler static strings will be smaller, but in inception
    ;; both will be the same because both are really heap-allocated
    ;; in reality.
  ))
