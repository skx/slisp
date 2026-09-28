(defun main()
  (let ((a  (strdup "slisp")) ; heap-allocated string
        (b  "slisp"))         ; static string

    ; The string lengths will be identical
    (println "heap-allocated str is " (strlen a) " bytes")
    (println "static str is "         (strlen b) " bytes")

    ; But the size of the "allocation" will be differ
    (println "heap-allocated str is " (alen a) " bytes")
    (println "static str is "         (alen b) " bytes")

    ;; heap-allocated things are rounded up to multiples of 16
    ;; static ones are not, but they have a fake header to allow
    ;; us to pretend we're there
  ))
