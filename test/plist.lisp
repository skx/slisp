(defun main ()

  ;; Start with an empty list
  (let ((p (list)))

    ;; Show starting point
    (println "Starting with empty list, and adding values.")

    ;; Add some values
    (set! p (plist-set p :name "Steve"))
    (set! p (plist-set p :age 19))

    ;; Show them
    (println "After setting name/age:")
    (println "\tName:" (plist-get p :name))
    (println "\tAge:" (plist-get p :age))

    (println "\tmissing-key:" (plist-get p :foo))
    (println "\tmissing-key, with default:" (plist-get p :foo "cake"))

    ;; Update the age and confirm it
    (set! p (plist-set p :age 29))
    (println "\tUpdated Age:" (plist-get p :age))

    ;; show the raw list
    (println "Final list: " p)))
