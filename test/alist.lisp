(require alist)

(defun main ()

  ;; alist tests
  (let ((a (list (list :name "John") (list :age 42) (list :city "Paris"))))

    ;; Show what we started with
    (println "We now have the following alist: " a)
    (println "\tname:" (alist:get a :name))
    (println "\tage:" (alist:get a :age))
    (println "\tcity:" (alist:get a :city))
    (println "\tmissing-key:" (alist:get a :missing_key))
    (println "\tmissing-key with default:" (alist:get a :missing_key "cake"))

    ;; Make a change
    (println "\tremoving city ..")
    (set! a (alist:remove a :city))
    (println "\t\tcity is now:" (alist:get a :city))

    ;; Make more change
    (println "\tupdating name .. twice ..")
    (set! a (alist:set (alist:set a :name "Steve") :name "bob"))
    (println "\t\tupdated name:" (alist:get a :name))

    ;; And show the whole list.
    (println "\tFinal list:" a)
    (println "\tKeys: " (alist:keys a))
    (println "\tVals: " (alist:values a))))
