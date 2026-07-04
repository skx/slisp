(defun insert-by (cmp x xs)
  (if xs
      (if (cmp x (car xs))
          (cons x xs)
          (cons (car xs) (insert-by cmp x (cdr xs))))
      (list x)))

(defun sort-by (cmp xs)
  (if xs
      (insert-by cmp
                 (car xs)
                 (sort-by cmp (cdr xs)))
      nil))


(defun main ()
  ;; get all files in the parent directory
  (let ((results (entries ".."))
        (mdown   nil))

    ;; filter the list
    (set! mdown (filter results
                        ;; for each filename
                        (lambda (file)
                          ;; split the path by "."
                          (let ((path (split file #\.)))
                            ;; if the split worked (there was a dot in the name)
                            (if (cons? path)
                                ;; if the extension is ".md"
                                (if (= "md" (nth path 1))
                                    ;; return the path
                                    path))))))

    ;; So now "mdown" should have all the filenames which end in ".md"
    ;; Sort them so that we can get a predictable order
    (set! mdown (sort-by (lambda (a b) (> 0 (strcmp a b))) mdown))

    (println "Found the following .md files in the parent directory:")
    (map (lambda (x) (println "\t" x)) mdown)
))
