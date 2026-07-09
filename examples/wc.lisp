;; wc.lisp - Word Count clone

(defun whitespace?(ch)
  "Is the given character whitespace?"
  (or (= ch #\Space)
      (= ch #\Tab)
      (= ch #\Newline)))

(defun words (chars)
  "Use reduce to sum up words. Which are things separated by whitespace"
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
  "Show lines/chars/words from the data - with filename too"
  (let ((size  (length data))
        (chars (explode data))
        (lines (filter chars (lambda (x) (= x #\Newline)))))
    (println (length lines) " " (car (words chars)) " " size " " filename)))

(defun process (file)
  "If we can open the file, read it and start processing"
  (let ((handle  (fopen file "r")) ; open
        (data    (fread handle))   ; read
        (discard (fclose handle))) ; close
    (if data
        (show-data data file))))


(defun main (args)
  "Process each specified filename, if any"
  (map (lambda (x) (process x )) (cdr args)))
