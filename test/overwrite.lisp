;; Test that latest function wins, for duplicate names

(defun foo ()
  (println "I should be invisible"))

(defun foo ()
  (println "Hello, World!"))


(defun main ()
  (foo))
