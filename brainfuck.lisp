;; brainfuck.lisp - Brainfuck interpreter in slisp
;;
;; This is slow.
;; This is terrible.
;; This works.
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


;;; Brainfuck Handlers

;; + handler
(defun execPlus (len program i cells ptr)
  (run len
       program
       (+ i 1)
       (setNth cells ptr (% (+ (nth cells ptr) 1) 256))
       ptr))

;; - handler
(defun execMinus (len program i cells ptr)
  (run len
       program
       (+ i 1)
       (setNth cells ptr (% (- (nth cells ptr) 1) 256))
       ptr))

;; > handler
(defun execGt (len program i cells ptr)
  (run len program (+ i 1) cells (+ ptr 1)))

;; < handler
(defun execLt (len program i cells ptr)
  (run len program (+ i 1) cells (- ptr 1)))

;; . handler
(defun execDot (len program i cells ptr)
    (putc (chr (nth cells ptr)))
    (run len program (+ i 1) cells ptr))

;; , handler
(defun execComma (len program i cells ptr)
  (let ((x (getc)))
    (if (nil? x)
         (set! x 0))
    (run len program (+ i 1) (setNth cells ptr x) ptr)))

;; [ handler
(defun execOpen (len program i cells ptr)
  (if (= (nth cells ptr) 0)
      (run len program
           (+ (findClose len program (+ i 1) 1) 1)
           cells
           ptr)
      (run len program
           (+ i 1)
           cells
           ptr)))

;; ] handler
(defun execClose (len program i cells ptr)
  (if (= (nth cells ptr) 0)
      (run len
           program
           (+ i 1)
           cells
           ptr)
      (run len
           program
           (+ (findOpen len program (- i 1) 1) 1)
           cells
           ptr)))


;;; Interpreter

;; Run a brainfuck program
(defun run (len program i cells ptr)

  ; if we're inside the program
  (if (< i len)

      ;; get the instruction
      (let ((ins (nth program i)))

        ;; dispatch it
        (cond
          ((= ins #\+)  (execPlus  len program i cells ptr))
          ((= ins #\-)  (execMinus len program i cells ptr))
          ((= ins #\>)  (execGt    len program i cells ptr))
          ((= ins #\<)  (execLt    len program i cells ptr))
          ((= ins #\.)  (execDot   len program i cells ptr))
          ((= ins #\,)  (execComma len program i cells ptr))
          ((= ins #\[)  (execOpen  len program i cells ptr))
          ((= ins #\])  (execClose len program i cells ptr))

          ;; ignore unknown character/instruction
          (1 (run len program (+ i 1) cells ptr))))))


;; Create ranges of numbers in a list
(defun makeCells (count)
    (if (> count 0)
        (cons 0 (makeCells (- count 1)))
      nil))

;; driver
(defun brainfuck (program)
  "Run the given program with 1000 cells."
  (run (length program) (explode program) 0 (makeCells 1000) 0))


;; Entry-point
(defun main (args)

  ; If we got an argument
  (if (= (length args) 2)
      (do
       ; if the argument was "cat"
       (if (= (car (cdr args)) "cat")
          ; run the cat-program
          (do
           (brainfuck ",[.,]")
           (exit 0)))
       ; if the argument was "factor"
       (if (= (car (cdr args)) "factor")
          ; run the factor-program
          (do
           (brainfuck "* factor an arbitrarily large positive integer** Copyright (C) 1999 by Brian Raiter* under the GNU General Public License>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>-** read in the number*<<<<<<<<<+[-[>>>>>>>>>>][-]<<<<<<<<<<[[->>>>>>>>>>+<<<<<<<<<<]<<<<<<<<<<]  >>>>>>>>>>,----------]>>>>>>>>>>[------------------------------------->>>>>>>>>->]<[+>[>>>>>>>>>+>]<-<<<<<<<<<<]-** display the number and initialize the loop variable to two*[>++++++++++++++++++++++++++++++++++++++++++++++++.  ------------------------------------------------<<<<<<<<<<<]++++++++++++++++++++++++++++++++++++++++++++++++++++++++++.--------------------------.[-]>>>>>>>>>>>>++<<<<+** the main loop*[ [-]>>  *  * make copies of the number and the loop variable  *  [>>>>[-]>[-]>[-]>[-]    >[-]>[-]    <<<<<<<[->>>+>+<<<<]>>>>>>>>]  <<<<<<<<<<[>>>>>>[-<<<<+>>>>]<<<<<<<<<<<<<<<<]>>>>>>>>>>  [>[->>>+>>+<<<<<]>>>>>>>>>]  <<<<<<<<<<[>>>>>>[-<<<<<+>>>>>]<<<<<<<<<<<<<<<<]>>>>>>>>>>  *  * divide the number by the loop variable  *  [>>>[-]>>>[-]>[-]>>>]                                  initialize  <<<<<<<<<<[<<<<<<<<<<]  >>>>>>>>>[-]>>>>>>>+<<<<<<<<[+]+  [ ->>                               double divisor until above dividend    [>>>>>>[->++<]>>>>]<<<<<<<<<<    [>>>>>>>>[-]>[-]       <<<<[->>>++<<<]<<<<<<<<<<<<<<<]>>>>>>>>>>    [>>>>>>>>[->+<[->+<[->+<[->+<[->+<[->+<[->+<[->+<[->+<            [->--------->>>>>>>>>+<<<<<<<<<<[->+<]]]]]]]]]]]>>]    <<<<<<<<<<[>>>>>>>>>[-<+<<<+>>>>]<<<<<<<<<<<<<<<<<<<]>>>>>>>>>>    [>>>>>>>[-<+>[-<+>[-<+>[-<+>[-<+>[-<+>[-<+>[-<+>[-<+>            [-<--------->>>>>>>>>>>+<<<<<<<<<<[-<+>]]]]]]]]]]]>>>]    <<<<<<<<<<    [>>>>[->>>+>>+<<<<<]<<<<<<<<<<<<<<]    >>>>>>>>>>[>>>>>>>[-<<<+>>>]>>>]<<<<<<<<<<    [>>>>>>>>[->-<]>      [<<<<<<<<<[<[-]>>>>>>>>>>[-<<<<<<<<<<+>>>>>>>>>>]<<<<<<<<<<<<<<<<<<<]        >>>>>>>>>>>>>>>>>>>]      <<<<<<<<<<<<<<<<<<<]    >>>>>>>>>[+[+[+[+[+[+[+[+[+[+[[-]<+>]]]]]]]]]]]<  ]  >>>>>>>>  [                                   subtract divisor from dividend    <<<<<<    [>>>>>>>>[-]>[-]<<<<<[->>>+>+<<<<]>>>>>>]<<<<<<<<<<    [>>>>>>>>[-<<<<+>>>>]<<<[->>>+>+<<<<]<<<<<<<<<<<<<<<]>>>>>>>>>>    [>>>>>>>>>[-<<<<+>>>>]>]<<<<<<<<<<    [>>>>>>>>[-<->]<<<<<<<<<<<<<<<<<<]>>>>>>>>>>    [>>>>>>>[->+<[->+<[->+<[->+<[->+<[->+<[->+<[->+<[->+<[->+<            [++++++++++[+>-<]>>>>>>>>>>-<<<<<<<<<<]]]]]]]]]]]>>>]    >>>>>>>+    [                                 if difference is nonnegative then      [-]<<<<<<<<<<<<<<<<<            replace dividend and increment quotient      [>>>>[-]>>>>[-<<<<+>>>>]<<[->>+<<]<<<<<<<<<<<<<<<<]>>>>>>>>>>      [>>>>>>>>[->+<<<+>>]>>]<<<<<<<<<<      [>>>[->>>>>>+<<<<<<]<<<<<<<<<<<<<]>>>>>>>>>>      [>>>>>>>>>[-<<<<<<+>>>>>>[-<<<<<<+>>>>>>                [-<<<<<<+>>>>>>[-<<<<<<+>>>>>>                [-<<<<<<+>>>>>>[-<<<<<<+>>>>>>                [-<<<<<<+>>>>>>[-<<<<<<+>>>>>>                [-<<<<<<+>>>>>>[-<<<<<<--------->>>>>>>>>>>>>>>>+<<<<<<<<<<                [-<<<<<<+>>>>>>]]]]]]]]]]]>]      >>>>>>>    ]                                 halve divisor and loop until zero    <<<<<<<<<<<<<<<<<[<<<<<<<<<<]>>>>>>>>>>    [>>>>>>>>[-]<<[->+<]<[->>>+<<<]>>>>>]<<<<<<<<<<    [+>>>>>>>[-<<<<<<<+>>>>>>>[-<<<<<<<->>>>>>+>             [-<<<<<<<+>>>>>>>[-<<<<<<<->>>>>>+>             [-<<<<<<<+>>>>>>>[-<<<<<<<->>>>>>+>             [-<<<<<<<+>>>>>>>[-<<<<<<<->>>>>>+>             [-<<<<<<<+>>>>>>>]]]]]]]]]<<<<<<<             [->>>>>>>+<<<<<<<]-<<<<<<<<<<]    >>>>>>>    [-<<<<<<<<<<<+>>>>>>>>>>>]      >>>[>>>>>>>[-<<<<<<<<<<<+++++>>>>>>>>>>>]>>>]<<<<<<<<<<    [+>>>>>>>>[-<<<<<<<<+>>>>>>>>[-<<<<<<<<->>>>>+>>>              [-<<<<<<<<+>>>>>>>>[-<<<<<<<<->>>>>+>>>              [-<<<<<<<<+>>>>>>>>[-<<<<<<<<->>>>>+>>>              [-<<<<<<<<+>>>>>>>>[-<<<<<<<<->>>>>+>>>              [-<<<<<<<<+>>>>>>>>]]]]]]]]]<<<<<<<<              [->>>>>>>>+<<<<<<<<]-<<<<<<<<<<]    >>>>>>>>[-<<<<<<<<<<<<<+>>>>>>>>>>>>>]>>    [>>>>>>>>[-<<<<<<<<<<<<<+++++>>>>>>>>>>>>>]>>]<<<<<<<<<<    [<<<<<<<<<<]>>>>>>>>>>    >>>>>>  ]  <<<<<<  *  * make copies of the loop variable and the quotient  *  [>>>[->>>>+>+<<<<<]>>>>>>>]  <<<<<<<<<<  [>>>>>>>[-<<<<+>>>>]<<<<<[->>>>>+>>+<<<<<<<]<<<<<<<<<<<<]  >>>>>>>>>>[>>>>>>>[-<<<<<+>>>>>]>>>]<<<<<<<<<<  *  * break out of the loop if the quotient is larger than the loop variable  *  [>>>>>>>>>[-<->]<    [<<<<<<<<      [<<[-]>>>>>>>>>>[-<<<<<<<<<<+>>>>>>>>>>]<<<<<<<<<<<<<<<<<<]    >>>>>>>>>>>>>>>>>>]<<<<<<<<<<<<<<<<<<]  >>>>>>>>[>-<[+[+[+[+[+[+[+[+[+[[-]>+<]]]]]]]]]]]>+  [ [-]    *    * partially increment the loop variable    *    <[-]+>>>>+>>>>>>>>[>>>>>>>>>>]<<<<<<<<<<    *    * examine the remainder for nonzero digits    *    [<<<<<<[<<<<[<<<<<<<<<<]>>>>+<<<<<<<<<<]<<<<]    >>>>>>>>>>>>>>>>>>>>[>>>>>>>>>>]<<<<<<<<<<[<<<<<<<<<<]    >>>>-    [ [+]      *      * decrement the loop variable and replace the number with the quotient      *      >>>>>>>>-<<[>[-]>>[-<<+>>]>>>>>>>]<<<<<<<<<<      *      * display the loop variable      *      [+>>[>>>>>>>>+>>]<<-<<<<<<<<<<]-      [>>++++++++++++++++++++++++++++++++++++++++++++++++.         ------------------------------------------------<<<<<<<<<<<<]      ++++++++++++++++++++++++++++++++.[-]>>>>    ]    *    * normalize the loop variable    *    >>>>>>    [>>[->>>>>+<<<<<[->>>>>+<<<<<       [->>>>>+<<<<<[->>>>>+<<<<<       [->>>>>+<<<<<[->>>>>+<<<<<       [->>>>>+<<<<<[->>>>>+<<<<<       [->>>>>+<<<<<[->>>>>--------->>>>>+<<<<<<<<<<       [->>>>>+<<<<<]]]]]]]]]]]>>>>>>>>]    <<<<<<<<<<[>>>>>>>[-<<<<<+>>>>>]<<<<<<<<<<<<<<<<<]    >>>>>>>>>  ]<]>>** display the number and end*[>>>>>>>>>>]<<<<<<<<<<[+>[>>>>>>>>>+>]<-<<<<<<<<<<]-[>++++++++++++++++++++++++++++++++++++++++++++++++.<<<<<<<<<<<]++++++++++.")
           (exit 0)))))

  ; not two arguments, or not cat
  ; hello world
  (brainfuck "++++++++[>++++[>++>+++>+++>+<<<<-]>+>+>->>+[<]<-]>>.>---.+++++++..+++.>>.<-.<.+++.------.--------.>>+.>++.")
  )
