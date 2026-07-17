;; nqueens solver for arbitrary sized boards
;;
;; This solves the 8x8 board by default, but you
;; may specify a larger sized board via the CLI
;; arguments.
;;
;; In the past this would exhaust our heap memory
;; on sizes of >= 12, however now we've added an
;; explicit GC call after printing each solution and
;; so it can handle a lot larger sizes.
;;
;; That said I got bored waiting for the solutions
;; to be generated for 20x20 - it was enough to show
;; that it did find a bunch of valid solutions so
;; I terminated my search after printing out a few
;; hundred of them!
;;
;; And still generating solutions for 22x22 will exhaust
;; the heap before generating even the first solution.
;; 21 seems to be about the most I can manage at the moment
;; but that seems like a reasonable size.
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
