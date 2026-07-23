;;
;; Data structures help!
;;
;; We used to store variables as a list, and the same for
;; user functions, and built-ins.  However that meant a linear
;; scan for lookups.
;;
;; Turns out storing things in a tree, not even a balanced/optimal
;; one is significantly faster and that really helps us execute
;; programs more quickly.
;;
;; So we have a tree-get/put and some helpers here for that purpose.
;;

;;; Node helpers

;; Get the node key from a tree item.
(defun node-key   (node)
  (car node))

;; Get the node value from a tree itme.
(defun node-value (node)
  (cadr node))

;; Get the left-item from a tree itme.
(defun node-left  (node)
  (caddr node))

;; Get the right-item from a tree itme.
(defun node-right (node)
  (cadddr node))


;;; Tree functions


;; Is there a key with the given name in the tree?
;;
;; Since we can set a nil value it isn't enough to do a lookup
;; of the key, as that would imply the item wasn't found.
(defun tree-bound? (tree key)
  (if (nil? tree)
      nil
      (if (= key (node-key tree))
          t
          (if (string< key (node-key tree))
              (tree-bound? (node-left tree) key)
              (tree-bound? (node-right tree) key)))))

;; Get an item from the tree
(defun tree-get (tree key)
  (if (nil? key)
      nil
      (if (nil? tree)
          nil
          (if (= key (car tree))
              (cadr tree)
              (if (string< key (car tree))
                  (tree-get (caddr tree) key)
                  (tree-get (cadddr tree) key))))))

;; put an item in the tree
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
