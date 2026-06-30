;; brainfuck.lisp - Brainfuck interpreter in slisp
;;
;; This is slow.
;; This is terrible.
;; This works.
;;
;; Usage:
;;
;;   ./brainfuck [path/to/program.bf]
;;
;; If no program is supplied the default one is executed.
;;
;;
;; IMPORTANT NOTE!
;; ###############
;;
;; Note that this program is **deeply** recursive, and
;; because each function call reserves (currently) 256 bytes
;; of stack space it will blow up the stack and end up
;; terminating with a segfault for all but the smallest
;; programs.
;;
;; To work around that run:
;;
;;   ulimit -s unlimited
;;
;; Before running this program, which will make the stack
;; allocation unlimited, and allow that deep recursion to
;; succeed.
;;

;;; Helpers

;; Return a copy of the given list, with the Nth item set to the given value.
(defun setNth (list n val)
  (if (> n 0)
      (cons (car list)
            (setNth (cdr list) (- n 1) val))
      (cons val (cdr list))))


;;; Brainfuck loop finding

(defun findOpen (len program pos depth)
  (let ((ch (nth program pos)))
    (if (= ch #\])
        (findOpen len program (- pos 1) (+ depth 1))
        (if (= ch #\[)
            (if (= depth 1)
                pos
                (findOpen len program (- pos 1) (- depth 1)))
            (findOpen len program (- pos 1) depth)))))


(defun findClose (len program pos depth)
  (let ((ch (nth program pos)))
    (if (= ch #\[)
        (findClose len program (+ pos 1) (+ depth 1))
        (if (= ch #\])
            (if (= depth 1)
                pos
                (findClose len program (+ pos 1) (- depth 1)))
            (findClose len program (+ pos 1) depth)))))



;;; Interpreter

(defun run (program)
  (let ((i 0)
        (len (length program))
        (ptr 0)
        (cells (makeCells 1000)))

    (while (< i len)
      (let ((ins (nth program i)))
        (cond

          ;; +
          ((= ins #\+) (do
                        (set! cells
                              (setNth cells ptr
                                      (% (+ (nth cells ptr) 1) 256)))
                        (set! i (+ i 1))))

          ;; -
          ((= ins #\-) (do
                        (set! cells
                              (setNth cells ptr
                                      (% (- (nth cells ptr) 1) 256)))
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
               (set! i (+ (findClose len program (+ i 1) 1) 1))
               (set! i (+ i 1))))

          ;; ]
          ((= ins #\])
           (if (= (nth cells ptr) 0)
               (set! i (+ i 1))
               (set! i (+ (findOpen len program (- i 1) 1) 1))))

          ;; ,
          ((= ins #\.) (do
                        (putc (chr (nth cells ptr)))
                        (set! i (+ i 1))))

          ;; ,
          ((= ins #\,) (do
                        (set! cells
                              (setNth cells ptr
                                      (% (getc) 256)))
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
