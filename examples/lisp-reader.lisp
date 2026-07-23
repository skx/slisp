;; The code is actually located within ../packages/
(require lisp-reader)



;;;
;;; Test Code
;;;
(defun reader-dump-program(path)
  (let ((handle (fopen path "r"))  ; open
        (data   (fread handle))    ; read
        (res    (fclose handle)))  ; close
    (when data
      (reader:init data)
      (println path ":")
      (println (reader:parse-program)))))


;;
;; Dump all arguments, or just the named one.
;;
;; args.lisp is small enough to be parsed quickly.
;;
(defun main(args)
  (if (nil? args)
      (reader-dump-program "args.lisp")
      (if (= 1 (length args))
          (reader-dump-program "args.lisp")
          (map (lambda (x) (reader-dump-program x)) (cdr args)))))
