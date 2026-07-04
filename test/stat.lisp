(defun main ()

  ;; basic tests
  (println "/etc exists as directory:" (dir? "/etc"))
  (println "/etc/fstab exists as directory:" (dir? "/etc/fstab"))
  (println "/etc/fstab exists as file:" (file? "/etc/fstab"))

  (let ((s (stat "stat.lisp")))

    ;; If stat succeeds it returns a three-element list
    (let ((type (nth s 0))
          (size (nth s 1))
          (mode (nth s 2)))

      ;; Show what we got
      (println "file type for stat.lisp: " type)
      (println "file size for stat.lisp: " size)
      (println "file mode for stat.lisp: " mode)))

  ;; Returns nil on failure
  (println "stat of a missing file:")
  (println (stat "/the/file/does'nt/exist")))
