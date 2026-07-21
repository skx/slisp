(defun tree-get (tree key)
  (if (nil? tree)
      nil
      (if (= key (car tree))
          (cadr tree)
          (if (string< key (car tree))
              (tree-get (caddr tree) key)
              (tree-get (cadddr tree) key)))))

(defun tree-put (tree key value)
  (if (nil? tree)
      (list key value nil nil)

      (if (= key (car tree))

          ;; replace existing value
          (list key
                value
                (caddr tree)
                (cadddr tree))

          (if (string< key (car tree))

              ;; insert left
              (list (car tree)
                    (cadr tree)
                    (tree-put (caddr tree) key value)
                    (cadddr tree))

              ;; insert right
              (list (car tree)
                    (cadr tree)
                    (caddr tree)
                    (tree-put (cadddr tree) key value))))))

;; Test the functions
(defun main ()
  (let ((tr nil))
    (set! tr (tree-put tr "apple" "apple"))
    (set! tr (tree-put tr "cake" "cake"))
    (set! tr (tree-put tr "shoe" "shoe"))
    (set! tr (tree-put tr "pig" "pig"))
    (set! tr (tree-put tr "dog" "cat"))
    (println tr)))
