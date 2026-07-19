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


;; Is the given value a number component "-", ".", and "0-9"
(defun numeric-char? (ch)
  (or
   (= ch "-")
   (= ch ".")
   (= ch "0")
   (= ch "1")
   (= ch "2")
   (= ch "3")
   (= ch "4")
   (= ch "5")
   (= ch "6")
   (= ch "7")
   (= ch "8")
   (= ch "9")))

;; Is the given string 100% numeric-plausible?
(defun number-string? (s)
  (cond
    ((= s "") nil)
    ((= (substr s 0 1) "-")
     (and
      (> (strlen s) 1)
      (number-string-aux s 1)))
    (t
     (number-string-aux s 0))))

;; helper
(defun number-string-aux (s pos)
  (if (>= pos (strlen s))
      t
      (if (numeric-char?
           (substr s pos 1))
          (number-string-aux
           s
           (+ pos 1))
          nil)))


;; skip over whitespace and comments
(defun reader-skip ()
  (let ((again t))
    (while again
      (set! again nil)

      ;; Skip whitespace
      (while
          (or
           (= (reader-peek) " ")
           (= (reader-peek) "\t")
           (= (reader-peek) "\n")
           (= (reader-peek) "\r"))
        (reader-next))

      ;; Skip a comment
      (if (= (reader-peek) ";")
          (let ()
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

  ;; skip until newline or EOF
  (while
      (and
       (!= (reader-peek) "")
       (!= (reader-peek) "\n"))
    (reader-next))
  ;; consume the newline if present
  (if (= (reader-peek) "\n")
      (reader-next)))

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
  (reader-next)   ; consume opening "

  (let ((text ""))

    (while
        (and
         (!= (reader-peek) "")
         (!= (reader-peek) "\""))

      (if (= (reader-peek) "\\")
          (do
           (reader-next)     ; consume '\'

           (let ((ch (reader-next)))
             (set! text
                   (strcat
                    text
                    (cond
                      ((= ch "n") "\n")
                      ((= ch "t") "\t")
                      ((= ch "\"") "\"")
                      ((= ch "\\") "\\")
                      (t ch))))))
          (set! text
                (strcat
                 text
                 (reader-next)))))

    (reader-next)    ; closing "
    text))


(defun reader-read-atom ()
  (let ((token ""))
    (while
        (and
         (!= (reader-peek) "")
         (!= (reader-peek) " ")
         (!= (reader-peek) "\t")
         (!= (reader-peek) "\r")
         (!= (reader-peek) "\n")
         (!= (reader-peek) "(")
         (!= (reader-peek) ")"))

      (set! token
            (strcat
             token
             (reader-next))))

    ;; Is this a number?
    (if (number-string? token)
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
