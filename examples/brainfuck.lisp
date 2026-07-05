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


;;; Brainfuck loop finding

(defun buildJumps (program)
  (buildJumpsRec program 0 nil nil))


(defun buildJumpsRec (program pos stack result)
  (if (= pos (length program))
      result

      (let ((ch (nth program pos)))
        (if (= ch #\[)
            ;; push position
            (buildJumpsRec
                program
                (+ pos 1)
                (cons pos stack)
                result)

            (if (= ch #\])
                (let ((open (car stack)))
                    (buildJumpsRec
                        program
                        (+ pos 1)
                        (cdr stack)
                        (cons (list open pos) result)))

                (buildJumpsRec
                    program
                    (+ pos 1)
                    stack
                    result))))))

(defun findJump (table pos)
    (if table
        (let ((pair (car table)))
            (if (= (car pair) pos)
                (car (cdr pair))
                (if (= (car (cdr pair)) pos)
                    (car pair)
                    (findJump (cdr table) pos))))
        nil))

;;; Interpreter

(defun run (program)
  (let ((i 0)                         ; offset into program
        (len (length program))        ; length of program
        (ptr 0)                       ; PTR value
        (cells (makeCells 1000))      ; cells.
        (jumps (buildJumps program))) ; jumps

    ; while we've not run off the end of the program
    (while (< i len)

      ; get the instruction
      (let ((ins (nth program i)))


        ; handle it
        (cond

          ;; +
          ((= ins #\+) (do
                        (setnth cells ptr (% (+ (nth cells ptr) 1) 256))
                        (set! i (+ i 1))))

          ;; -
          ((= ins #\-) (do
                        (setnth cells ptr (% (- (nth cells ptr) 1) 256))
                        (set! i (+ i 1))))

          ;; >
          ((= ins #\>) (do
                        (set! ptr (+ ptr 1))
                        (set! i (+ i 1))))

          ;; <
          ((= ins #\<) (do
                        (set! ptr (- ptr 1))
                        (set! i (+ i 1))))


          ;; [
          ((= ins #\[)
           (if (= (nth cells ptr) 0)
               (set! i (+ (findJump jumps i) 1))
               (set! i (+ i 1))))

          ;; ]
          ((= ins #\])
           (if (= (nth cells ptr) 0)
               (set! i (+ i 1))
               (set! i (+ (findJump jumps i) 1))))

          ;; ,
          ((= ins #\.) (do
                        (putc (chr (nth cells ptr)))
                        (set! i (+ i 1))))

          ;; ,
          ((= ins #\,) (do
                        (setnth cells ptr (% (getc) 256))
                        (set! i (+ i 1))))

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

  (run (explode program))))
