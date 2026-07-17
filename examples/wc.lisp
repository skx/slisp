;; wc.lisp - Word Count clone
;;

;; Load our argument-parsing package.
(require arg-parser)

;; formatting width for the numbers
(defconst field-width 5)

;; variables set by CLI parser.
(defvar show-lines nil)
(defvar show-words nil)
(defvar show-bytes nil)

;; global totals, for when he handle multiple files.
(defvar total-lines 0)
(defvar total-words 0)
(defvar total-bytes 0)

;; helpers
(defun first (xs)
  (nth xs 0))

(defun second (xs)
  (nth xs 1))

(defun third (xs)
  (nth xs 2))

(defun leftpad (str n)
  (if (< (length str) n)
      (leftpad (strcat " " str) n)
      str))


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

(defun file-stats (data)
  "Get the statistics for a given piece of data."
  (let ((size  (length data))
        (chars (explode data))
        (lines (length (filter chars (lambda (x) (= x #\Newline)))))
        (wds (car (words chars))))
    (list lines wds size)))

(defun show-data (stats filename)
  "Show lines/chars/words from the data alongside the filename"
  (let ((lines (first stats))
        (words (second stats))
        (bytes (third stats)))

    (if show-lines
        (print (leftpad (string lines) field-width) " "))

    (if show-words
        (print (leftpad (string words) field-width) " "))

    (if show-bytes
        (print (leftpad (string bytes) field-width) " "))

    (println filename)))


(defun process (file)
  "If we can open the file, read it and start processing."

  (let ((handle  (fopen file "r"))
        (data    (fread handle))
        (discard (fclose handle)))

    (if data
        ;; get the stats
        (let ((stats (file-stats data)))

          ;; update the global totals
          (set! total-lines (+ total-lines (first stats)))
          (set! total-words (+ total-words (second stats)))
          (set! total-bytes (+ total-bytes (third stats)))

          ;; print this file
          (show-data stats file)

          t)
        nil)))

(defun help ()
  (println "Usage:")
  (println "\twc [flags] file1 file2 .. fileN\n")
  (println "Flags:")
  (println "\t -l, --lines - Show line counts.")
  (println "\t -c, --bytes - Show byte counts.")
  (println "\t -m, --chars - Show char counts.")
  (println "\t -w, --words - Show word counts.")

  (exit 1))

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
             ((or (= arg "--help") (= arg "-?"))  (help))
             ((or (= arg "-h") (= arg "-?"))      (help))
             (t                                   (do (println "Unknown argument: " arg "\n") (help)))))
         (parser :flags))

    ;; No options?  Behave like wc and show "everything"
    (if (and (not show-lines)
             (not show-words)
             (not show-bytes))
        (do
          (set! show-lines t)
          (set! show-words t)
          (set! show-bytes t)))


    ; For each file process them in order
    (map (lambda (file) (process file)) files)

    ; more than one file?  Then show the totals
    (if (> (length (parser :files)) 1)
        (do

         (if show-lines
             (print (leftpad (string total-lines) field-width) " "))

         (if show-words
             (print (leftpad (string total-words) field-width) " "))

          (if show-bytes
              (print (leftpad (string total-bytes) field-width) " "))

          (println "total")))))
