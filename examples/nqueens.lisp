;; nqueens solver for arbitrary sized boards
;;
;; Deeply recursive using backtracking this dies
;; with a segfault for me on sizes >= 10.
;;
;; Usage
;;
;;   ./nqueens [size]
;;


(defvar SIZE 8)

(defvar solution-count 0)

(defun safe-helper (placed col distance)
  (if (nil? placed)
      t
      (if (or (= (car placed) col)
              (= (abs (- (car placed) col)) distance))
          nil
          (safe-helper (cdr placed)
                       col
                       (+ distance 1)))))

(defun safe? (placed col)
  (safe-helper placed col 1))


;; printing routines
(defun print-cell (queen-col col)
  (if (= queen-col col)
      (print "Q ")
      (print ". ")))

(defun print-row-helper (queen-col col)
  (if (<= col SIZE)
      (do
        (print-cell queen-col col)
        (print-row-helper queen-col (+ col 1)))))

(defun print-row (queen-col)
  (do
    (print "\t")
    (print-row-helper queen-col 1)
    (newline)))

(defun print-board (board)
  (if board
      (do
        (print-row (car board))
        (print-board (cdr board)))))

(defun print-solution (board)
  (set! solution-count (+ solution-count 1))
  (println "Solution " solution-count " " board ":\n")
  (print-board board)
  (newline)
  (newline))

;;
;; ------------------------------------------------------------
;; Solver
;;

(defun solve-row (row placed)
  (if (> row SIZE)
      ;; complete solution
      (print-solution (reverse placed))

      ;; otherwise try every column
      (try-column row placed 1)))

(defun try-column (row placed col)
  (if (<= col SIZE)
      (do
        (if (safe? placed col)
            (solve-row (+ row 1)
                       (cons col placed)))
        (try-column row
                    placed
                    (+ col 1)))))

;;
;; ------------------------------------------------------------
;;

(defun main (args)
  (if (= (length args) 2)
      (set! SIZE (atoi (car (cdr args)))))

  (println "\n8 Queens Solver for board size " SIZE "x" SIZE "\n")

  (set! solution-count 0)
  (solve-row 1 nil)

  (println "Found " solution-count " solutions.")
)
