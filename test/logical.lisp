(defun main ()

  (println "start")

  ; and; true
  (if (and t t)
      (println "and is OK"))
  ; and; false
  (if (and t t t t nil t)
     (println "not ok")
     (println "Still okay!"))

  ; or: true
  (if (or t nil nil nil)
      (println "or is OK"))
  ; or: false
  (if (or nil nil nil)
      (println "not OK")
      (println "OR is still okay :)"))


  (println "end")
  )
