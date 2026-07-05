;; A global variable named "foo"
(defvar foo "Before")

;; A global variable named "bar".
;;
;; Attempting to change this will generate an error at compile-time.
(defconst bar "i am unchanging")

(defun local ()
  "Test scoping by having a local variable with the same name as a global.

Spoiler: Local variable always comes first."
  (let ((foo "local"))
    (println "\tI'm inside (foo)")
    (println "\t\tlocal variable:" foo)
    (set! foo "bar")
    (println "\t\tupdated local variable:" foo)))



(defun main ()
  "Entry Point."

  ;; show global foo
  (println "global variable is:" foo)

  ;; update global foo
  (set! foo "Changed")

  ;; Show it was updated
  (println "updated global variable is now:" foo)

  ;; Call a local function
  (local)

  (println "Global variable untouched:" foo)

  (println "The global constant 'bar' is fine:" bar)
  )
