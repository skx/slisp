;;
;; A property list (plist for short) is a list of paired elements
;; stored in a regular list.
;;
;; The example of storing details of a person in the alist-section
;; above would become this;
;;
;;  ( name "Steve" age 42 email "steve@example.com" )
;;


(defun plist:new()
  (list))

(defun plist:get (plist key &default)
  "Given a plist get the value of the specified key.

Return nil if it wasn't present, or the given default value."
  (cond
    ((null? plist)       (car default))
    ((= (car plist) key) (car (cdr plist)))
    (t                   (plist:get (cdr (cdr plist)) key (car default)))))

(defun plist:keys (plist)
  "Return a list of all stored keys."
  (cond
    ((null? plist) nil)
    (t             (cons (car plist) (plist:keys (cdr (cdr plist)))))))

(defun plist:remove (plist key)
  "Given a plist remove any entry for the given key, if present.

            Returns the updated list."
  (cond
    ((null? plist)       nil)
    ((= (car plist) key) (plist:remove (cdr (cdr plist)) key))
    (t                   (cons (car plist) (cons (car (cdr plist)) (plist:remove (cdr (cdr plist)) key))))))

(defun plist:set (plist key value)
  "Add the given key/value pair to the specified plist.

This removes any pre-existing value with that key first."
  (append (plist:remove plist key) (list key value)))

(defun plist:values (plist)
  "Return a list of all stored values."
  (cond
    ((null? plist)     nil)
    (t                 (cons (car (cdr plist)) (plist:values (cdr (cdr plist)))))))
