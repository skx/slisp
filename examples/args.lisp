(require arg-parser)

(defun a (a b c d e f)
  "Just confirm we support six arguments"
  (sum (list a b c d e f))
  )


(defun main(args)
  (a 1 2 3 4 5 6)
  (let ((parser (arg-parser:new (cdr args))))
    (println "Flags " (parser :flags))
    (println "Files " (parser :files))))
