;;
;; Data structures help!
;;
;; Within our inception interpreter we used to store variables
;; as a list, and the same for user functions, and built-ins.
;;
;; That worked but it was slow, as it meant a linear scan for lookups.
;; (I also made things worse by never deleting items, just adding
;; them to the list.  Oops)
;;
;; Turns out storing things in a tree, not even a balanced/optimal
;; one is significantly faster and that really helps us execute
;; programs more quickly.
;;
;; So we have a tree-get/put and some helpers here for that purpose.
;;

;;; Node helpers

(defun node-key (node)
  (car node))

(defun node-value (node)
  (cadr node))

(defun node-left (node)
  (caddr node))

(defun node-right (node)
  (cadddr node))

(defun node-height (node)
  (if (nil? node)
      0
      (car (cddddr node))))

(defun make-node (key value left right)
  (list key
        value
        left
        right
        (+ 1
           (max (list (node-height left) (node-height right))))))


;;
;; AVL magic stuff
;;

(defun balance-factor (node)
  (- (node-height (node-left node))
     (node-height (node-right node))))


(defun rotate-right (y)
  (let ((x (node-left y)))
    (make-node
      (node-key x)
      (node-value x)
      (node-left x)
      (make-node
        (node-key y)
        (node-value y)
        (node-right x)
        (node-right y)))))


(defun rotate-left (x)
  (let ((y (node-right x)))
    (make-node
      (node-key y)
      (node-value y)
      (make-node
        (node-key x)
        (node-value x)
        (node-left x)
        (node-left y))
      (node-right y))))


(defun rebalance (node)
  (let ((b (balance-factor node)))

    ;; Left heavy
    (if (> b 1)
        (if (< (balance-factor (node-left node)) 0)
            ;; Left-Right
            (rotate-right
              (make-node
                (node-key node)
                (node-value node)
                (rotate-left (node-left node))
                (node-right node)))

            ;; Left-Left
            (rotate-right node))


        ;; Right heavy
        (if (< b -1)
            (if (> (balance-factor (node-right node)) 0)
                ;; Right-Left
                (rotate-left
                  (make-node
                    (node-key node)
                    (node-value node)
                    (node-left node)
                    (rotate-right (node-right node))))

                ;; Right-Right
                (rotate-left node))
            node))))


;;; Tree functions

;; Constructor.
(defun tree:new()
  nil)

;; Is there a key with the given name in the tree?
;;
;; Since we can set a nil value it isn't enough to do a lookup
;; of the key, as that would imply the item wasn't found.
(defun tree:bound? (tree key)
  (if (nil? tree)
      nil
      (if (= key (node-key tree))
          t
          (if (string< key (node-key tree))
              (tree:bound? (node-left tree) key)
              (tree:bound? (node-right tree) key)))))

;; Get keys in our tree
(defun tree:keys (tree)
  (tree:keys-aux tree nil))

(defun tree:keys-aux (tree acc)
  (if (nil? tree)
      acc
      (tree:keys-aux
        (node-left tree)
        (cons (node-key tree)
              (tree:keys-aux (node-right tree) acc)))))

;; Get an item from the tree
(defun tree:get (tree key)
  (if (nil? key)
      nil
      (if (nil? tree)
          nil
          (if (= key (node-key tree))
              (node-value tree)
              (if (string< key (node-key tree))
                  (tree:get (node-left tree) key)
                  (tree:get (node-right tree) key))))))


;; Put an item in the tree
(defun tree:put (tree key value)
  (if (nil? tree)
      (make-node key value nil nil)
      (if (= key (node-key tree))
          ;; Replace existing value
          (make-node
            key
            value
            (node-left tree)
            (node-right tree))
          (if (string< key (node-key tree))
              ;; Insert left
              (rebalance
                (make-node
                  (node-key tree)
                  (node-value tree)
                  (tree:put (node-left tree) key value)
                  (node-right tree)))

              ;; Insert right
              (rebalance
                (make-node
                  (node-key tree)
                  (node-value tree)
                  (node-left tree)
                  (tree:put (node-right tree) key value)))))))
