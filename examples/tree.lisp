;; The code is actually located within ../packages/
(require tree)

;;; Test code

(defun main ()
  (let ((tr nil))
    (set! tr (tree-put tr "apple" "apple"))
    (set! tr (tree-put tr "cake" "cake"))
    (set! tr (tree-put tr "shoe" "shoe"))
    (set! tr (tree-put tr "pig" "pig"))
    (set! tr (tree-put tr "dog" "cat"))

    (if (tree-bound? tr "apple")
        (println "OK.. apple is bound")
        (println "FAIL: apple"))

    (if (tree-bound? tr "fake")
        (println "FAIL: fake")
        (println "OK.. fake is not bound"))

    (println tr)))
