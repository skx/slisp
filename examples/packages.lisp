;; Declare a package named "foo"
(package foo
         ;; Declare a global variable within this package.
         (defvar version 1)

         ;; Local function accesses the local version variable
         ;; and can change it with that name.
         (defun update()
           (println "Version before change was " version)
           (set! version "alpha")
           (println "version after change is " version))

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

         ;; Local function accesses the local version variable
         ;; and can change it with that name.
         (defun update()
           (println "Version before change was " version)
           (set! version "alpha")
           (println "version after change is " version))

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
  (foo:update)
  (foo:steve)
  (foo:noise)

  ; change the package variable
  (set! foo:version "1.one.alpha")
  (foo:steve)

  (newline)

  (bar:update)
  (bar:steve)
  (bar:noise)

  ; change the package variable
  (set! bar:version "2.two.beta")
  (bar:steve)

)
