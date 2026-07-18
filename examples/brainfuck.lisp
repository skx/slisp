;; brainfuck.lisp - Brainfuck interpreter in slisp
;;
;; This is slow.  This is terrible.  This works.
;;
;; Usage:
;;
;;   ./brainfuck [path/to/program.bf]
;;
;; If no program is supplied the default one is executed
;; (which is the "hello world" program).
;;


;;; utility

(defun string-to-list (s)
  "Split the given string into a list of characters.

We use this because while our compiler has 'implode' it will return a list of
CHARACTERS and our interpreter doesn't understand character types."
  (if (= (strlen s) 0)
      nil
      (cons
       (substr s 0 1)
       (string-to-list (substr s 1 (- (strlen s) 1))))))

(defun program-length (xs)
  "We have a custom function here because our interpreter doesn't have a
(length str|lst) implementation in its standard library."
  (if (nil? xs)
      0
      (+ 1 (program-length (cdr xs)))))


;;; Brainfuck loop finding

(defun buildJumps (program)
  (let ((table (makeCells (program-length program))))
    (buildJumpsRec program 0 nil table)
    table))

(defun buildJumpsRec (program pos stack table)
  (if (= pos (program-length program))
      table
      (let ((ch (nth program pos)))
        (cond
          ((= ch "[")
           (buildJumpsRec
            program
            (+ pos 1)
            (cons pos stack)
            table))
          ((= ch "]")
           (let ((open (car stack)))

             ;; store both directions
             (nth! table open pos)
             (nth! table pos open)
             (buildJumpsRec
              program
              (+ pos 1)
              (cdr stack)
              table)))
          (t
           (buildJumpsRec
            program
            (+ pos 1)
            stack
            table))))))

;;; Interpreter

(defun run (program)

  (let ((i 0)                          ; offset into program
        (len (program-length program)) ; length of program
        (ptr 0)                        ; PTR value
        (cells (makeCells 1000))       ; cells.
        (jumps (buildJumps program)))  ; jumps


    ; while we've not run off the end of the program
    (while (< i len)

      ; get the next instruction
      (let ((ins (nth program i)))


        ; handle it
        (cond

          ;; +
          ((= ins "+")
                        (let ((v (nth cells ptr)))
                          (nth! cells ptr (% (+ v 1) 256)))
                        (set! i (+ i 1)))

          ;; -
          ((= ins "-")
                        (let ((v (nth cells ptr)))
                          (nth! cells ptr (% (- v 1) 256)))
                        (set! i (+ i 1)))

          ;; >
          ((= ins ">")
                        (set! ptr (+ ptr 1))
                        (set! i (+ i 1)))

          ;; <
          ((= ins "<")
                        (set! ptr (- ptr 1))
                        (set! i (+ i 1)))

          ;; [
          ((= ins "[")
           (if (= (nth cells ptr) 0)
               (set! i (+ (nth jumps i) 1))
               (set! i (+ i 1))))

          ;; ]
          ((= ins "]")
           (if (= (nth cells ptr) 0)
               (set! i (+ i 1))
               (set! i (+ (nth jumps i) 1))))

          ;; ,
          ((= ins ".")
                        (print (chr (nth cells ptr)))
                        (set! i (+ i 1)))

          ;; ,
          ((= ins ",")
                        (nth! cells ptr (getc))
                        (set! i (+ i 1)))

          ;; skip over unknown instructions
          (t (set! i (+ i 1))))))))


;; Create ranges of numbers in a list
(defun makeCells (count)
  (if (> count 0)
      (cons 0 (makeCells (- count 1)))
      nil))


;; Entry-point
(defun main (args)

  (let ((program "++++++++[>++++[>++>+++>+++>+<<<<-]>+>+>->>+[<]<-]>>.>---.+++++++..+++.>>.<-.<.+++.------.--------.>>+.>++."))

    ; If we got an argument.
    (if (= (length args) 2)
        (do
         (let ((handle  (fopen (car (cdr args)) "r")) ; open
               (data    (fread handle))               ; read
               (discard (fclose handle)))             ; close
           (if data
               (set! program data)
               (do
                (print "failed to read file ")
                (println (car (cdr args)))
                 (exit 1))))))


    (run (string-to-list program))))
