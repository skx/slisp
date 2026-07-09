;; Test floating-point division.
;;
;; Integer division produces SigFPE when dividing by zero,
;; but floating-points do not - by default - instead they
;; return results that make sense, and we now format them
;; correctly.
;;
(defun main ()
  (println (/ 3 0.0))    ; +Inf
  (println (/ -3 0.0))   ; -Inf
  (println (/ 0.0 0.0))) ; NaN
