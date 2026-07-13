;; life.lisp - Game of life.
;;
;; Board size: 80 x 25
;;

(defconst WIDTH 80)
(defconst HEIGHT 25)

;;
;; The board we're working on and displaying
;;
(defvar board nil)

;;
;; Generation count
;;
(defvar generation 0)

;;
;; The next board
;;
(defvar next-board nil)


;; Print an escape character
;;
;; This is a bit clunky, but it allows our code to be executed by the
;; inception interpreter.
(defun escape ()
  (chr 27))

;; Build a row containing WIDTH zeroes.
(defun make-row ()
  (repeated WIDTH 0))


;; Build a board containing HEIGHT rows.
(defun make-board ()
  (map
    (lambda (n)
      (make-row))
    (seq (- HEIGHT 1))))


;; Clear each cell of the board
(defun clear-board (b)
  (let ((y 0))
    (while (< y HEIGHT)
      (let ((row (nth b y))
            (x 0))
        (while (< x WIDTH)
          (nth! row x 0)
          (set! x (+ x 1))))
      (set! y (+ y 1))))
  b)


;; Draw the state of the board.
(defun draw-board (b)
  (print (escape) "[?25l")  ; hide cursor
  (print (escape) "[H")     ; move home & clear
  (print (escape) "[2J")

  (let ((y 0))
    (while (< y HEIGHT)
      (let ((row (nth b y))
            (x 0))
        (while (< x WIDTH)
          (if (= (nth row x) 1)
              (print "#")
              (print " "))
          (set! x (+ x 1)))
        (newline)
        (set! y (+ y 1)))))
  (println "generation " generation)
  (print (escape) "[?25h") ; restore cursor
)


;;
;; Standard glider:
;;
;; . O .
;; . . O
;; O O O
;;
(defun place-glider (b x y)
  (set-cell b (+ x 1) (+ y 0) 1)
  (set-cell b (+ x 2) (+ y 1) 1)
  (set-cell b (+ x 0) (+ y 2) 1)
  (set-cell b (+ x 1) (+ y 2) 1)
  (set-cell b (+ x 2) (+ y 2) 1))


;;
;; Random entries for the board
;;
(defun randomize (b)
  (let ((y 0))
    (while (< y HEIGHT)
      (let ((row (nth b y))
            (x 0))
        (while (< x WIDTH)
          (if (= (random 5) 0)
              (nth! row x 1)
              (nth! row x 0))
          (set! x (+ x 1))))
      (set! y (+ y 1)))))

;;
;; Count the eight neighbours around a cell.
;;
(defun count-neighbours (b x y)
  (let ((n 0))

    (set! n (+ n (get-cell b (- x 1) (- y 1))))
    (set! n (+ n (get-cell b (+ x 0) (- y 1))))
    (set! n (+ n (get-cell b (+ x 1) (- y 1))))

    (set! n (+ n (get-cell b (- x 1) (+ y 0))))
    (set! n (+ n (get-cell b (+ x 1) (+ y 0))))

    (set! n (+ n (get-cell b (- x 1) (+ y 1))))
    (set! n (+ n (get-cell b (+ x 0) (+ y 1))))
    (set! n (+ n (get-cell b (+ x 1) (+ y 1))))

    n))


;;
;; Evolve.
;;
(defun next-state (alive neighbours)
  (if (= alive 1)
      (if (< neighbours 2)
          0
          (if (> neighbours 3)
              0
              1))
      (if (= neighbours 3)
          1
          0)))

;;
;; Get the given row
;;
(defun row (b y)
  (nth b y))

;;
;; Get a cell, by coordinates.
;;
(defun get-cell (b x y)
  (if (< x 0)
      0
      (if (< y 0)
          0
          (if (>= x WIDTH)
              0
              (if (>= y HEIGHT)
                  0
                  (nth (row b y) x))))))

;;
;; Set the contents of the given cell
;;
(defun set-cell (b x y value)
  (nth! (row b y) x value))


;;
;; Step forwards, evolving each cell.
;;
(defun step-board (src dst)
  (let ((y 0))
    (while (< y HEIGHT)
      (let ((src-row (row src y))
            (dst-row (row dst y))
            (x 0))
        (while (< x WIDTH)
          (nth! dst-row
                x
                (next-state
                    (nth src-row x)
                    (count-neighbours src x y)))
          (set! x (+ x 1)))
      )
      (set! y (+ y 1))))
  dst)


;;
;; Swap the global boards.
;;
(defun swap-boards ()
  (let ((tmp board))
    (set! board next-board)
    (set! next-board tmp)))


;;
;; Seed withe either a bunch of gliders, or
;; random cells.
;;
(defun seed (glide)
  (clear-board board)
  (if glide
      (do
       (place-glider board  4 2)
       (place-glider board 25 2)
       (place-glider board 20 10)
       (place-glider board 20 20)
       (place-glider board 50 14)
       (place-glider board 48 15))
      (randomize board)))



;;
;; Entry point.
;;
(defun main (args)
  (set! board (make-board))
  (set! next-board (make-board))

  ;; no arguments? then random display
  ;; otherwise gliders
  (if (nil? args)
      (seed 1) ; glider
      (if (> (length args) 1)
          (seed 1) ; glider
          (seed nil))) ; random

  (while 1
    (set! generation (+ generation 1))
    (draw-board board)
    (step-board board next-board)
    (swap-boards)
    (system "sleep 0.1")))
