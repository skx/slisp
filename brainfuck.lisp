;; brainfuck.lisp - Brainfuck interpreter in slisp
;;
;; This is slow.
;; This is terrible.
;; This works.
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
       (setNth cells ptr (% (+ (nth cells ptr) 1) 256))
       ptr))

;; - handler
(defun execMinus (program i cells ptr)
  (run program
       (+ i 1)
       (setNth cells ptr (% (- (nth cells ptr) 1) 256))
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

;; , handler
(defun execComma (program i cells ptr)
  (let ((x (getc)))
    (if (nil? x)
         (set! x 0))
    (run program (+ i 1) (setNth cells ptr x) ptr)))

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
          ((= ins #\,)  (execComma program i cells ptr))
          ((= ins #\[)  (execOpen  program i cells ptr))
          ((= ins #\])  (execClose program i cells ptr))

          ;; ignore unknown character/instruction
          (1 (run program (+ i 1) cells ptr))))))


;; Create ranges of numbers in a list
(defun makeCells (count)
    (if (> count 0)
        (cons 0 (makeCells (- count 1)))
      nil))

;; driver
(defun brainfuck (program)
  "Run the given program with 1000 cells."
  (run (explode program) 0 (makeCells 1000) 0))


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
