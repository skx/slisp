;;
;; A property list (plist for short) is a list of paired elements
;; stored in a regular list.
;;
;; The example of storing details of a person in the alist-section
;; above would become this;
;;
;;  ( name "Steve" age 42 email "steve@example.com" )
;;



(package plist

         ;; Export these functions as prefixed versions
         (alias! "plist-get"    "plist:get")
         (alias! "plist-keys"   "plist:keys")
         (alias! "plist-new"    "plist:new")
         (alias! "plist-remove" "plist:remove")
         (alias! "plist-set"    "plist:set")
         (alias! "plist-values" "plist:values")

         (defun new()
           (list))

         (defun get (plist key &default)
           "Given a plist get the value of the specified key.

Return nil if it wasn't present, or the given default value."
           (cond
             ((null? plist)       (car default))
             ((= (car plist) key) (car (cdr plist)))
             (t                   (get (cdr (cdr plist)) key (car default)))))

         (defun keys (plist)
           "Return a list of all stored keys."
           (cond
             ((null? plist) nil)
             (t             (cons (car plist) (keys (cdr (cdr plist)))))))

         (defun remove (plist key)
           "Given a plist remove any entry for the given key, if present.

            Returns the updated list."
           (cond
             ((null? plist)       nil)
             ((= (car plist) key) (remove (cdr (cdr plist)) key))
             (t                   (cons (car plist) (cons (car (cdr plist)) (remove (cdr (cdr plist)) key))))))

         (defun set (plist key value)
           "Add the given key/value pair to the specified plist.

This removes any pre-existing value with that key first."
           (append (remove plist key) (list key value)))

         (defun values (plist)
           "Return a list of all stored values."
           (cond
             ((null? plist)     nil)
             (t                 (cons (car (cdr plist)) (values (cdr (cdr plist)))))))

) ; end of package
