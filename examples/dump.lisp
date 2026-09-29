;; dump.lisp - Simple utility to dump file contents.
;;
;; Show the contents as hex bytes, and also ASCII
;; characters when they are printable.
;;
;; Example usage:
;;
;;  ./dump ./dump
;;

;; Load our argument-parsing library
(require arg-parser)

(defun dump (str)
  "Dump an allocated region of memory to STDOUT as hex and ASCCI characters."

  ;; remove header from the allocated length
  (let ((len (- (alen str) 24))
        (i   0)    ;; index
        (hex "")   ;; hex version of output
        (ch  "")   ;; character version of output
        (c   nil)) ;; current character

    ;; For each byte
    (while  (< i len)
      (set! c (strat str i))

      ;; hex will have the hex bytes of the characters.
      (set! hex (strcat (strcat hex (dec2hex c)) " "))

      ;; ch will have the character bytes.
      ;; Need to mask out non-printable characters
      (if (and (>= c 32) (<= c 127))
          (set! ch (strcat ch (string (chr c))))
        (set! ch (strcat ch ".")))

      ;; end of group?  Then print and reset.
      (when (= (% i 16) 15)
        (print hex)
        (print " ")
        (print ch)
        (print "\n")
        (set! ch "")
        (set! hex ""))

      ;; Next loop
      (set! i (+ 1 i)))

    ;; If the dump wasn't a multiple of 8 bytes in size
    ;; print the remainder.  But pad out the hex to make
    ;; the ASCII line up properly.
    (while (<= (strlen hex) (* 5 16))  ;; "0x00 " = 5 bytes
      (set! hex (strcat hex " ")))

    (print hex)
    (print "")
    (print ch)
    (print "\n")
    (newline)))

(defun dump_file (name)
  (let ((handle  (fopen name "r")) ; open
        (data    (fread handle))   ; read
        (discard (fclose handle))) ; close
    (if data
        (dump data)
      (do
       (println "Error reading " name)
       (exit 1)))))


(defun help ()
  (println "Usage: dump file1 file2 .. fileN")
  (exit 0))


(defun main (args)

  ;; parse arguments
  (let ((parser (arg-parser:new (cdr args)))
        (files  (parser :files)))

    ;; process flags.
    (map (lambda (arg)
           (cond
             ((or (= arg "--help") (= arg "-?"))  (help))
             ((or (= arg "-h") (= arg "-?"))      (help))
             (t                                   (do (println "Unknown argument: " arg "\n") (help)))))
         (parser :flags))

    ;; If we have files process them, otherwise show help
    (if (> (length files) 0)
        (map (lambda (file) (dump_file file)) files)
      (help))))
