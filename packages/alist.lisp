;;; alist functions.
;;
;; An association list is a way of implementing a "hash-like" structure
;; using only a list of lists.  For example to store details about a person
;; you might have this alist:
;;
;;  ( (name "Steve") (age 42) (email "steve@example.com") )
;;


(defun alist:new()
  "Create a new associated list / hash"
  (list))

(defun alist:get (alist key &default)
  "Given an alist get the value of the specified key.

Return nil if it wasn't present, or the given default value."
  (cond
    ((null? alist)             (car default))
    ((= (car (car alist)) key) (car (cdr (car alist))))
    (t                         (alist:get (cdr alist) key (car default)))))

(defun alist:keys (alist)
  "Return a list of all stored keys."
  (cond
    ((null? alist) nil)
    (t             (cons (car (car alist)) (alist:keys (cdr alist))))))

(defun alist:remove (alist key)
  "Given an alist remove the entry for the given key, if it was present.

Returns the updated list."
  (cond
    ((null? alist)             nil)
    ((= (car (car alist)) key) (alist:remove (cdr alist) key))
    (t                         (cons (car alist) (alist:remove (cdr alist) key)))))

(defun alist:set (alist key value)
  "Add the given key/value pair to the specified alist.

This removes any pre-existing value with that key first."
  (cons (list key value) (alist:remove alist key)))

(defun alist:values (alist)
  "Return a list of all stored values."
  (cond
    ((null? alist) nil)
    (t             (cons (car (cdr (car alist))) (alist:values (cdr alist))))))
