;; A global variable named "foo"
(defvar foo "Before")

;; A global variable named "bar".
;;
;; Attempting to change this will generate an error at compile-time.
(defconst bar "i am unchanging")

;; Declare some colours, as RGB values.
(defvar blue    (list   0   0 255))
(defvar midblue (list 20 174 244))
(defvar pink    (list 255 192 203))
(defvar green   (list   0 255   0))
(defvar yellow  (list 255 255 224))
(defvar cyan    (list   0 244 244))
(defvar mustard (list 255 206  27))

;; A scene we might pretend to draw.
;;
;; Mostly here to confirm we load things correctly in our
;; init-code
;;
;; Ordering is important as we have no Z-index.
(defvar scene
  (list
   (list "box" 350 150 400 300 pink)
   (list "circle" 20 20 15 mustard)
   (list "box" 0 0 20 400 blue)
   (list "circle" 20 20 75 midblue)
   (list "circle" 20 20 175  cyan)
   (list "circle" 320 210 75 green )
   (list "background" yellow)))


(defun local ()
  "Test scoping by having a local variable with the same name as a global.

Spoiler: Local variable always comes first."
  (let ((foo "local"))
    (println "\tI'm inside (foo)")
    (println "\t\tlocal variable:" foo)
    (set! foo "bar")
    (println "\t\tupdated local variable:" foo)))



(defun main (args)
  "Entry Point."

  (sys-gc)

  ;; show global foo
  (println "global variable is:" foo)

  ;; update global foo
  (set! foo "Changed")

  ;; Show it was updated
  (println "updated global variable is now:" foo)

  ;; Call a local function
  (local)

  (sys-gc)

  (println "Global variable untouched:" foo)

  (println "The global constant 'bar' is fine:" bar)
  )
