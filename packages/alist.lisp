;;; alist functions.
;;
;; An association list is a way of implementing a "hash-like" structure
;; using only a list of lists.  For example to store details about a person
;; you might have this alist:
;;
;;  ( (name "Steve") (age 42) (email "steve@example.com") )
;;

(package alist

         ;; Export these functions as prefixed versions
         (alias! "alist-get"    "alist:get")
         (alias! "alist-keys"   "alist:keys")
         (alias! "alist-new"    "alist:new")
         (alias! "alist-remove" "alist:remove")
         (alias! "alist-set"    "alist:set")
         (alias! "alist-values" "alist:values")

         (defun new()
           "Create a new associated list / hash"
           (list))

         (defun get (alist key &default)
           "Given an alist get the value of the specified key.

Return nil if it wasn't present, or the given default value."
           (cond
             ((null? alist)             (car default))
             ((= (car (car alist)) key) (car (cdr (car alist))))
             (t                         (get (cdr alist) key (car default)))))

         (defun keys (alist)
           "Return a list of all stored keys."
           (cond
             ((null? alist) nil)
             (t             (cons (car (car alist)) (keys (cdr alist))))))

         (defun remove (alist key)
           "Given an alist remove the entry for the given key, if it was present.

Returns the updated list."
           (cond
             ((null? alist)             nil)
             ((= (car (car alist)) key) (remove (cdr alist) key))
             (t                         (cons (car alist) (remove (cdr alist) key)))))

         (defun set (alist key value)
           "Add the given key/value pair to the specified alist.

This removes any pre-existing value with that key first."
           (cons (list key value) (remove alist key)))

         (defun values (alist)
           "Return a list of all stored values."
           (cond
             ((null? alist) nil)
             (t             (cons (car (cdr (car alist))) (values (cdr alist))))))


         ) ; end of package
