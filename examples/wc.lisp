;; wc.lisp - Word Count clone
;;
;; NOTE: "./wc *.lisp" does NOT produce a "total" line, which is an obvious omission.


;; Load our argument-parsing package.
(require arg-parser)

;; variables set by CLI parser.
(defvar show-lines nil)
(defvar show-words nil)
(defvar show-bytes nil)


(defun whitespace?(ch)
  "Is the given character whitespace?"
  (or (= ch #\Newline)
      (= ch #\Return)
      (= ch #\Space)
      (= ch #\Tab)))

(defun words (chars)
  "Use reduce to sum up words, which is basically a matter of matching boundaries."
  (reduce chars
        (lambda (state ch)
          (let ((count (car state))
                (in-word (cdr state)))
            (if (whitespace? ch)
                (cons count nil)
                (if in-word
                    state
                    (cons (+ count 1) t)))))
        (cons 0 nil)))

(defun show-data (data filename)
  "Show lines/chars/words from the data alongside the filename"
  (let ((size  (length data))
        (chars (explode data))
        (lines (filter chars (lambda (x) (= x #\Newline)))))
    (if show-lines
        (print (length lines) " "))

    (if show-words
        (print (car (words chars)) " "))

    (if show-bytes
        (print size " "))

    (println filename)))


(defun process (file)
  "If we can open the file, read it and start processing."
  (let ((handle  (fopen file "r")) ; open
        (data    (fread handle))   ; read
        (discard (fclose handle))) ; close
    (if data
        (show-data data file))))

(defun main (args)
  "Produce char/line/word stats for each named file."

  (let ((parser (arg-parser:new (cdr args)))
        (files  (parser :files)))

    (map (lambda (arg)
           (cond
             ((or (= arg "-l") (= arg "--lines")) (set! show-lines t))
             ((or (= arg "-c") (= arg "--bytes")) (set! show-bytes t))
             ((or (= arg "-m") (= arg "--chars")) (set! show-bytes t))
             ((or (= arg "-w") (= arg "--words")) (set! show-words t))
             ((= arg "--help") (println "help"))
             (t                (do
                                (println "unknown arg '" arg "'")
                                (exit 1)))))
         (parser :flags))

    ;; No options?  Behave like wc and show "everything"
    (if (and (not show-lines)
             (not show-words)
             (not show-bytes))
        (do
          (set! show-lines t)
          (set! show-words t)
          (set! show-bytes t)))

    ;; now process files
    (map (lambda (x) (process x)) files)))
