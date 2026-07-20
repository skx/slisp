;;
;; The lisp-reader package reads source code and returns lisp objects.
;;
;; It is designed to be included in other programs, and is included
;; in that way by inception.lisp.
;;
;; There is a standalone `main` function at the foot of this file for
;; running a self-contained test.  When other programs include it
;; their own main functions will override the one here.
;;

;; The length of the program.
(defvar *reader-length* 0)

;; The position we're at goes here.
(defvar *reader-pos*    0)

;; The source we parse goes here.
(defvar *reader-text*   "")

;; Helper.
(defun symbol (name)
  (list "symbol" name))


;; constructor
(defun reader-init (text)
  "Setup state for processing the given text input."
  (set! *reader-length* (strlen text))
  (set! *reader-pos* 0)
  (set! *reader-text* text))

(defun reader-eof ()
  "Have we reached the end of our input?"
  (>= *reader-pos* *reader-length*))

(defun reader-peek ()
  "Look at the next character, without consuming it"
  (if (reader-eof)
      ""
      (substr
       *reader-text*
       *reader-pos*
       1)))

(defun reader-next ()
  "Return the next character"
  (let ((ch (reader-peek)))
    (set! *reader-pos* (+ *reader-pos* 1))
    ch))

;; is the character whitespace?
(defun whitespace? (ch)
  (or (= ch " ")
      (= ch "\t")
      (= ch "\n")
      (= ch "\r")))

;; Skip to the next token
(defun reader-skip ()
  (let ((again t)
        (ch ""))

    (while again
      (set! again nil)

      (set! ch (reader-peek))

      ;; Skip whitespace
      (while (whitespace? ch)
        (reader-next)
        (set! ch (reader-peek)))

      ;; Skip a comment
      (if (= ch ";")
          (do
            (reader-skip-comment)
            (set! again t))))))

(defun reader-read ()
  (reader-skip)
  (cond
    ((= (reader-peek) "(")
     (reader-read-list))
    ((= (reader-peek) ")")
     (do
      (println "unexpected ')' character.  Aborting")
      (exit 1)))
    ((= (reader-peek) "\"")
     (reader-read-string))
    ((= (reader-peek) "#")
       (list "character" (reader-read-dispatch)))
    ((= (reader-peek) "'")
     (reader-next)
     (list
      (symbol "quote")
      (reader-read)))
    (t
     (reader-read-atom))))

(defun reader-skip-comment ()
  ;; consume the ';'
  (reader-next)

  (let ((ch (reader-peek)))

    ;; skip until newline or EOF
    (while
        (and
         (!= ch "")
         (!= ch "\n"))

      (reader-next)
      (set! ch (reader-peek)))

    ;; consume the newline if present
    (if (= ch "\n")
        (reader-next))))

(defun reader-read-dispatch ()
  ;; consume '#'
  (reader-next)
  (cond
    ((= (reader-peek) "\\")
     (reader-read-char-literal))
    (t
     (println "Unknown reader dispatch")
     (exit 1))))


(defun alpha? (ch)
  "Is the given STRING character an alphabetical one?"
  (let ((c (car (explode ch))))
    (or (and (>= (ord c) (ord #\a)) (<= (ord c) (ord #\z)))
        (and (>= (ord c) (ord #\A)) (<= (ord c) (ord #\Z))))))


(defun reader-read-char-literal ()
  ;; consume '\'
  (reader-next)

  ;; read first character
  (let ((token (reader-next))
        (ch "")
        (reading t))

    (if (= token "")
        (do
          (println "Unexpected EOF in character literal")
          (exit 1)))

    ;; If the first character is alphabetic, continue
    ;; reading alphabetic characters.
    (if (alpha? token)
        (do
          (set! ch (reader-peek))

          (while reading
            (if (= ch "")
                (set! reading nil)
                (if (alpha? ch)
                    (do
                      (reader-next)
                      (set! token (strcat token ch))
                      (set! ch (reader-peek)))
                    (set! reading nil))))))

    (cond
      ;; named characters
      ((= token "Space")   " ")
      ((= token "Newline") "\n")
      ((= token "Tab")     "\t")

      ;; single-character literals
      ((= (strlen token) 1)
       token)

      (t
       (do
         (println "Unknown character literal:")
         (println token)
         (exit 1))))))

(defun reader-read-list ()
  ;; consume '('
  (reader-next)
  (let ((items nil))
    (reader-skip)
    (while (!= (reader-peek) ")")
      (set! items
            (cons
             (reader-read)
             items))
      (reader-skip))
    ;; consume ')'
    (reader-next)
    (reverse items)))

(defun reader-read-string ()
  ;; consume opening "
  (reader-next)

  (let ((text "")
        (ch (reader-peek)))

    (while
        (and
         (!= ch "")
         (!= ch "\""))

      (if (= ch "\\")
          (do
            ;; consume '\'
            (reader-next)

            ;; read escaped character
            (set! ch (reader-next))

            (set! text
                  (strcat
                   text
                   (cond
                     ((= ch "n") "\n")
                     ((= ch "t") "\t")
                     ((= ch "r") "\r")
                     ((= ch "\"") "\"")
                     ((= ch "\\") "\\")
                     (t ch))))

            ;; peek next character
            (set! ch (reader-peek)))

          (do
            ;; consume ordinary character
            (reader-next)

            (set! text
                  (strcat text ch))

            ;; peek next character
            (set! ch (reader-peek)))))

    ;; consume closing quote
    (if (= ch "\"")
        (reader-next))

    text))

(defun digit? (ch)
  (or (= ch "0")
      (= ch "1")
      (= ch "2")
      (= ch "3")
      (= ch "4")
      (= ch "5")
      (= ch "6")
      (= ch "7")
      (= ch "8")
      (= ch "9")))

(defun reader-read-atom ()
  (let ((token "")
        (numeric t)
        (seen-digit nil)
        (seen-dot nil)
        (pos 0)
        (ch (reader-peek)))

    (while
        (and
         (!= ch "")
         (!= ch " ")
         (!= ch "\t")
         (!= ch "\r")
         (!= ch "\n")
         (!= ch "(")
         (!= ch ")"))

      ;; consume the character we've already peeked
      (reader-next)

      (set! token (strcat token ch))

      ;; Update numeric state
      (if numeric
          (cond
            ((digit? ch)
             (set! seen-digit t))

            ((= ch ".")
             (if seen-dot
                 (set! numeric nil)
                 (set! seen-dot t)))

            ((= ch "-")
             (if (!= pos 0)
                 (set! numeric nil)))

            (t
             (set! numeric nil))))

      (set! pos (+ pos 1))

      ;; Peek once for the next iteration
      (set! ch (reader-peek)))

    (if (and numeric seen-digit)
        (atof token)
        (symbol token))))

;; parse a complete program and return all the expressions within it.
(defun reader-parse-program ()
  (let ((forms nil))
    (reader-skip)
    (while (not (reader-eof))
      (set! forms
            (cons
             (reader-read)
             forms))
      (reader-skip))
    (reverse forms)
    ))


;;;
;;; Test Code
;;;
(defun reader-dump-program(path)
  (let ((handle (fopen path "r"))  ; open
        (data   (fread handle))    ; read
        (res    (fclose handle)))  ; close
    (when data
      (reader-init data)
      (println path ":")
      (println (reader-parse-program)))))


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
