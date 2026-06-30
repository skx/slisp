(defun main ()
  
  (println "start")

  ; and; true
  (if (and (list t t))
      (println "and is OK"))
  ; and; false
  (if (and (list t t t t nil t))
     (println "not ok")
     (println "Still okay!"))

  ; or: true
  (if (or (list t nil nil nil ))
      (println "or is OK"))
  ; or: false
  (if (or (list nil nil nil ))
      (println "not OK")
      (println "OR is still okay :)"))


  (println "end")
  )
