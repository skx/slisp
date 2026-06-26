;; brainfuck.lisp - Brainfuck interpreter in slisp
;;
;; This is slow.
;; This is terrible.
;; This works.
;;
;; TODO: Wrap cell inc/dec - via % 256.
;; TODO: Input via ","; note our stdlib is missing getc..

;;; Helpers

;; set the Nth item of a list to the given value
(defun setNth (list n val)
  (if (> n 0)
      (cons (car list)
            (setNth (cdr list) (- n 1) val))
      (cons val (cdr list))))

;; Get the Nth item of a list
(defun nth (lst n)
  (if (= n 0)
    (car lst)
    (nth (cdr lst) (- n 1))))


;;; Brainfuck loop finding

(defun findOpen (program pos depth)
  (let ((ch (nth program pos)))
    (if (= ch #\])
        (findOpen program (- pos 1) (+ depth 1))
        (if (= ch #\[)
            (if (= depth 1)
                pos
                (findOpen program (- pos 1) (- depth 1)))
            (findOpen program (- pos 1) depth)))))


(defun findClose (program pos depth)
  (let ((ch (nth program pos)))
    (if (= ch #\[)
        (findClose program (+ pos 1) (+ depth 1))
        (if (= ch #\])
            (if (= depth 1)
                pos
                (findClose program (+ pos 1) (- depth 1)))
            (findClose program (+ pos 1) depth)))))


;;; Brainfuck Handlers

;; + handler
(defun execPlus (program i cells ptr)
  (run program
       (+ i 1)
       (setNth cells ptr (+ (nth cells ptr) 1))
       ptr))

;; - handler
(defun execMinus (program i cells ptr)
  (run program
       (+ i 1)
       (setNth cells ptr (- (nth cells ptr) 1))
       ptr))

;; > handler
(defun execGt (program i cells ptr)
  (run program (+ i 1) cells (+ ptr 1)))

;; < handler
(defun execLt (program i cells ptr)
  (run program (+ i 1) cells (- ptr 1)))

;; . handler
(defun execDot (program i cells ptr)
    (putc (chr (nth cells ptr)))
    (run program (+ i 1) cells ptr))

;; [ handler
(defun execOpen (program i cells ptr)
  (if (= (nth cells ptr) 0)
      (run program
           (+ (findClose program (+ i 1) 1) 1)
           cells
           ptr)
      (run program
           (+ i 1)
           cells
           ptr)))

;; ] handler
(defun execClose (program i cells ptr)
  (if (= (nth cells ptr) 0)
      (run program
           (+ i 1)
           cells
           ptr)
      (run program
           (+ (findOpen program (- i 1) 1) 1)
           cells
           ptr)))


;;; Interpreter

;; Run a brainfuck program
(defun run (program i cells ptr)

  ; if we're inside the program
  (if (< i (length program))

      ;; get the instruction
      (let ((ins (nth program i)))

        ;; dispatch it
        (cond
          ((= ins #\+)  (execPlus  program i cells ptr))
          ((= ins #\-)  (execMinus program i cells ptr))
          ((= ins #\>)  (execGt    program i cells ptr))
          ((= ins #\<)  (execLt    program i cells ptr))
          ((= ins #\.)  (execDot   program i cells ptr))
          ((= ins #\[)  (execOpen  program i cells ptr))
          ((= ins #\])  (execClose program i cells ptr))

          ;; ignore unknown character/instruction
          (1 (run program (+ i 1) cells ptr))))))


;; driver
(defun brainfuck (program)
  "Run the given program with 30 cells"
  (run
   ;; program
   (explode program)
   ;; offset being executed
   0
   ;; cells
   (list 0 0 0 0 0 0 0 0 0 0   ; 10
         0 0 0 0 0 0 0 0 0 0   ; 20
         0 0 0 0 0 0 0 0 0 0 ) ; 30
   ;; cell ptr value
   0))


;; Entry-point
(defun main ()

  (brainfuck "++++++++[>++++[>++>+++>+++>+<<<<-]>+>+>->>+[<]<-]>>.>---.+++++++..+++.>>.<-.<.+++.------.--------.>>+.>++.")
)
