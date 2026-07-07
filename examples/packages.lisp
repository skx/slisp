;; Declare a package named "foo"
(package foo
         ;; Declare a global variable within this package.
         (defvar version 1)

         ;; Declare a package-function, named "noise"
         ;;
         ;; This will be callable by "foo:noise" from outside the package.
         (defun noise()
           (println "\tNoise:meow"))

         ;; Declare a package-function, named "steve"
         ;;
         ;; This will be callable by "foo:steve" from outside the package.
         ;;
         ;; Here note we call "noise" unqualified, we don't need to add
         ;; the package prefix if we don't want to.
         (defun steve ()
           (println "foo::steve")
           (println "\tMy version is " version)
           (noise)
         )
)

;; Declare a package named "bar"
(package bar
         ;; Declare a global variable within this package.
         (defvar version 2)

         ;; Declare a package-function, named "noise"
         ;;
         ;; This will be callable by "bar:noise" from outside the package.
         (defun noise()
           (println "\tNoise:woof"))

         ;; Declare a package-function, named "steve"
         ;;
         ;; This will be callable by "bar:steve" from outside the package.
         ;;
         ;; Here note we call "noise" unqualified, we don't need to add
         ;; the package prefix if we don't want to.
         ;;
         (defun steve ()
           (println "bar::steve")
           (println "\tMy version is " version)
           (bar:noise)
         )
)

(defun main()
  (foo:steve)
  (foo:noise)
  (bar:steve)
  (bar:noise)
)
