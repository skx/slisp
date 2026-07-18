;;; arg-parser - Trivial arg parsing package
;;
;; This package makes the assumption that command
;; line arguments have two forms:
;;
;;  - options prefixed with "-"
;;    (for example "-l", "-ltr", or "--long".
;;
;;  - filenames with no prefix
;;
;; The constructor returns a lambda function which
;; may be used to return either "flags", or "files",
;; being the two types of argument
;;
;; To ease usage any short-form flags which have been
;; merged together are split up into distinct flags
;; so "foo -ltr" will have the flags returned as a
;; list ("-l" "-t" "-r").
;;


;;
;; Main function - which returns a lambda
;; that can retrieve either "flags" or "files".
;;
(defun arg-parser:new (copy)
  (lambda (type)
    (cond
      ((= type :flags) (arg-parser:flags-helper copy))
      ((= type :files) (arg-parser:files copy))
      (t               (do
                        (println "unknown mode for arg-parser lambda")
                        (exit 1))))))


;;
;; Internal helper routines
;;
(defun arg-parser:files (a)
  "Return non-flag arguments, i.e. files"
  (filter a
          (lambda (arg)
            (if (> (length arg) 0)
                (! (= (nth (explode arg) 0) #\-))))))

(defun arg-parser:flags-helper (args)
  "Return the flags from this argument."
  (if args
      (append
       (if (> (length (car args)) 1) (arg-parser:flags-from-arg (car args)) (list "-"))
       (arg-parser:flags-helper (cdr args)))
      nil))

(defun arg-parser:flags-from-arg (arg)
  "If the given argument is a flag then handle it appropriately."
  (let ((chars (explode arg)))
    (cond
      ;; Not a flag.
      ((! (= (nth chars 0) #\-)) nil)

      ;; Long flag (--foo)
      ((= (nth chars 1) #\-) (list arg))

      ;; Short flag(s): -abc -> ("-a" "-b" "-c")
      (t (arg-parser:expand-short-flags (cdr chars))))))

(defun arg-parser:expand-short-flags (chars)
  "Given a list of short flags 'xyz' return (-x -y -z)"
  (if (not chars)
      nil
      (cons (strcat "-" (string (car chars)))
            (arg-parser:expand-short-flags (cdr chars)))))
