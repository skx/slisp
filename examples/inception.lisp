;; A minimal lisp interpreter, meant to be compiled with slisp.
;;
;; Features:
;;
;;   Integer & floating-point numbers.
;;   Strings.
;;   Functions.
;;   Lambdas (closures).
;;
;; We store built-in primitives, variables, and functions in different namespaces.
;;
;; Special forms:
;;   IF
;;   LAMBDA
;;   LET
;;   QUOTE
;;   DEFUN
;;   DEFVAR
;;
;; Nothing too surprising I guess.
;;


;; Since we don't "tag" things we wrap them in lists
(defun builtin (name)
  (list "builtin" name))

(defun closure (params body env)
  (list "closure" params body env))

(defun symbol (name)
  (list "symbol" name))


;; now we need helpers to see if we have a given thing
(defun builtin? (x)
  (and (cons? x)
       (= (car x) "builtin")))

(defun closure? (x)
  (and (cons? x)
       (= (car x) "closure")))

(defun symbol? (x)
  (and
   (cons? x)
   (= (car x) "symbol")))


;; And some type-specific helpers.
(defun symbol-name (x)
  (cadr x))

(defun builtin-name (x)
  (cadr x))

;; get the names of parameters
(defun symbols-names (lst)
  (if (nil? lst)
      nil
      (cons
       (symbol-name (car lst))
       (symbols-names (cdr lst)))))




;; global storage for user-functions
(defvar *functions*  nil)

;; global storage for global variables
(defvar *globals* nil)

(defun global-get (name)
  (env-get *globals* name))

(defun global-set (name value)
  (set! *globals* (env-set *globals* name value)))


;; Lookup a builtin function.
(defun lookup-builtin (name)
  (cond
    ((= name "+")       (builtin "+"))
    ((= name "%")       (builtin "%"))
    ((= name "-")       (builtin "-"))
    ((= name "*")       (builtin "*"))
    ((= name "/")       (builtin "/"))
    ((= name "<")       (builtin "<"))
    ((= name "<=")      (builtin "<="))
    ((= name ">")       (builtin ">"))
    ((= name ">=")      (builtin ">="))
    ((= name "=")       (builtin "="))
    ((= name "abs")     (builtin "abs"))
    ((= name "cons")    (builtin "cons"))
    ((= name "car")     (builtin "car"))
    ((= name "cdr")     (builtin "cdr"))
    ((= name "list")    (builtin "list"))
    ((= name "nat")     (builtin "nat"))
    ((= name "not")     (builtin "not"))
    ((= name "print")   (builtin "print"))
    ((= name "println") (builtin "println"))
    ((= name "seq")     (builtin "seq"))
    ((= name "nil?")    (builtin "nil?"))
    (t nil)))

;; add a function
(defun add-function (name params body)
  (set! *functions*
        (cons
         (list
          name
          (closure params body nil))
         *functions*))
  name)

;; get a function, by name
(defun lookup-function (name)
  (lookup-function-aux
   name
   *functions*))

;; helper
(defun lookup-function-aux (name functions)
  (if (nil? functions)
      nil
      (if (= (car (car functions)) name)
          (cadr (car functions))
          (lookup-function-aux
           name
           (cdr functions)))))

;; lookup a binding from the environment
(defun env-get (env name)
  (if (nil? env)
      nil
      (if (= (car (car env)) name)
          (cadr (car env))
          (env-get (cdr env) name))))


;; set a variable in the environment
(defun env-set (env name value)
  (cons
   (list name value)
   env))


;; eval: where the magic happens.
(defun eval (expr env)
  (cond
    ;; integers evaluate to themselves
    ((int? expr) expr)

    ;; floats are self-evaluating too.
    ((float? expr) expr)

    ;; strings are self-evaluating too.
    ((str? expr) expr)

    ;; symbols will be looked up
    ((symbol? expr) (eval-symbol expr env))

    ;; lists are expressions
    ((cons? expr) (eval-list expr env))

    ;; something unknown
    (t nil)))

;; helper which is used by apply, do, and let.
(defun eval-body (forms env)
    (let ((result nil))
        (while forms
            (set! result (eval (car forms) env))
            (set! forms (cdr forms)))
        result))


;; function-call, lambda, builtin, etc.
(defun eval-call (expr env)
  (let ((fn   (eval (car expr) env))
        (args (map (lambda (x) (eval x env)) (cdr expr))))
    (apply fn args)))

;; special form: cond
(defun eval-cond (expr env)
    (eval-cond-clauses (cdr expr) env))

(defun eval-cond-clauses (clauses env)
    (if (nil? clauses)
        nil
        (let ((clause (car clauses)))
            (if (or
                    (= (car clause) t)
                    (eval (car clause) env))
                (eval-body (cdr clause) env)
                (eval-cond-clauses
                    (cdr clauses)
                    env)))))

;; special form: defun
(defun eval-defun (expr)
  (add-function
   (symbol-name (cadr expr))    ; name
   (symbols-names (caddr expr)) ; params
   (cdddr expr))                ; body
  ;; return the name of the defun
  (cadr expr))

;; special form: defvar - return the value
(defun eval-defvar (expr env)
  (let ((name  (symbol-name (cadr expr)))
        (value (eval (caddr expr) env)))
    (set! *globals* (env-set *globals* name value)) value))

;; special form: do - run each statement in the list
(defun eval-do (expr env)
  (eval-body (cdr expr) env))

;; special form; if
(defun eval-if (expr env)
  (if (eval (cadr expr) env)
      (eval (caddr expr) env)
      (eval (cadddr expr) env)))

;; evaluate a lambda
(defun eval-lambda (expr env)
    (closure
        (symbols-names (cadr expr)) ; params
        (cddr expr)                 ; body
        env))

;; special form: let
(defun eval-let (expr env)
  (let ((bindings (cadr expr))
        (body     (cddr expr))
        (new-env  env))
    (while bindings
      (set! new-env
            (env-set
             new-env
             (symbol-name (car (car bindings)))
             (eval (cadr (car bindings))
                   env)))
      (set! bindings (cdr bindings)))
    (eval-body body new-env)))


;; evaluate a list - sleazy
(defun eval-list (expr env)
  (let ((op (symbol-name (car expr))))
    (cond
      ((= op "cond")
       (eval-cond expr env))
      ((= op "defun")
       (eval-defun expr))
      ((= op "defvar")
       (eval-defvar expr env))
      ((= op "do")
       (eval-do expr env))
      ((= op "if")
       (eval-if expr env))
      ((= op "lambda")
       (eval-lambda expr env))
      ((= op "let")
       (eval-let expr env))
      ((= op "quote")
       (eval-quote expr env))
      (t
       (eval-call expr env)))))

;; return the contents of a symbol
(defun eval-symbol (sym env)
  (let ((name (symbol-name sym)))

    ;; boolean constants
    (cond
      ((= name "t")
       t)

      ((= name "nil")
       nil)

      (t
    ;; environment
    (let ((value (env-get env name)))
      (if value
          value

          ;; global variable
          (let ((value (env-get *globals* name)))
            (if value
                value

                ;; builtin
                (let ((fn (lookup-builtin name)))
                  (if fn
                      fn
                      ;; user-function
                      (lookup-function name)))))))))))

;; special form: quote
(defun eval-quote (expr env)
  (cadr expr))


;; apply for built-in and user-functoins
(defun apply (fn args)
  (cond
    ;; native builtins
    ((builtin? fn) (apply-builtin fn args))

    ;; user functions and lambdas
    ((closure? fn) (apply-closure fn args))

    ;; nothing known
    (t nil)))

;; built-in functions all live here, which is a bit horrid.
(defun apply-builtin (fn args)
  (let ((name (builtin-name fn)))
    (cond
      ((= name "+")
       (+ (car args) (cadr args)))

      ((= name "%")
       (% (car args) (cadr args)))

      ((= name "-")
       (- (car args) (cadr args)))

      ((= name "*")
       (* (car args) (cadr args)))

      ((= name "/")
       (/ (car args) (cadr args)))

      ((= name "<")
       (< (car args) (cadr args)))

      ((= name "<=")
       (<= (car args) (cadr args)))

      ((= name ">")
       (> (car args) (cadr args)))

      ((= name ">=")
       (>= (car args) (cadr args)))

      ((= name "=")
       (= (car args) (cadr args)))

      ((= name "abs")
       (abs (car args)))

      ((= name "cons")
       (cons (car args) (cadr args)))

      ((= name "car")
       (car (car args)))

      ((= name "cdr")
       (cdr (car args)))

      ((= name "list")
       args)

      ((= name "nat")
       (nat (car args)))

      ((= name "not")
       (not (car args)))

      ((= name "print")
       (do
        (while args
          (print (car args))
          (set! args (cdr args)))
        nil))

      ((= name "println")
       (do
        (while args
          (print (car args))
          (set! args (cdr args)))
        (print "\n")
        nil))

      ((= name "seq")
       (seq (car args)))

      ((= name "nil?")
       (nil? (car args)))

      (t
       nil))))


(defun apply-closure (closure args)
  (let ((params (cadr closure))
        (body   (caddr closure))
        (env    (cadddr closure)))

    ;; extend the captured environment
    (while params
      (set! env
            (env-set
             env
             (car params)
             (car args)))

      (set! params (cdr params))
      (set! args   (cdr args)))

    ;; Now execute all the statements in the body
    (eval-body body env)))


;;
;; Reader
;;
(defvar *reader-text* "")
(defvar *reader-pos* 0)

(defun reader-init (text)
  (set! *reader-text* text)
  (set! *reader-pos* 0))

(defun reader-eof ()
  (>= *reader-pos*
      (strlen *reader-text*)))

(defun reader-peek ()
  (if (reader-eof)
      ""
      (substr
       *reader-text*
       *reader-pos*
       1)))

(defun reader-next ()
  (let ((ch (reader-peek)))
    (set! *reader-pos* (+ *reader-pos* 1))
    ch))


;; Is the given value a number component "-", ".", and "0-9"
(defun numeric-char? (ch)
  (or
   (= ch "-")
   (= ch ".")
   (= ch "0")
   (= ch "1")
   (= ch "2")
   (= ch "3")
   (= ch "4")
   (= ch "5")
   (= ch "6")
   (= ch "7")
   (= ch "8")
   (= ch "9")))

;; Is the given string 100% numeric-plausible?
(defun number-string? (s)
  (cond
    ((= s "") nil)
    ((= (substr s 0 1) "-")
     (and
      (> (strlen s) 1)
      (number-string-aux s 1)))
    (t
     (number-string-aux s 0))))

;; helper
(defun number-string-aux (s pos)
  (if (>= pos (strlen s))
      t
      (if (numeric-char?
           (substr s pos 1))
          (number-string-aux
           s
           (+ pos 1))
          nil)))


;; skip over whitespace and comments
(defun reader-skip ()
  (let ((again t))
    (while again
      (set! again nil)

      ;; Skip whitespace
      (while
          (or
           (= (reader-peek) " ")
           (= (reader-peek) "\t")
           (= (reader-peek) "\n")
           (= (reader-peek) "\r"))
        (reader-next))

      ;; Skip a comment
      (if (= (reader-peek) ";")
          (let ()
            (reader-skip-comment)
            (set! again t))))))

(defun reader-read ()
  (reader-skip)
  (cond
    ((= (reader-peek) "(")
     (reader-read-list))
    ((= (reader-peek) ")")
     (do
      (println "unexpected ')' character.  Aborting")
      (exit 1)))
    ((= (reader-peek) "\"")
     (reader-read-string))
    ((= (reader-peek) "'")
     (reader-next)
     (list
      "quote"
      (reader-read)))
    (t
     (reader-read-atom))))

(defun reader-skip-comment ()
  ;; consume the ';'
  (reader-next)

  ;; skip until newline or EOF
  (while
      (and
       (!= (reader-peek) "")
       (!= (reader-peek) "\n"))
    (reader-next))
  ;; consume the newline if present
  (if (= (reader-peek) "\n")
      (reader-next)))

(defun reader-read-list ()
  ;; consume '('
  (reader-next)
  (let ((items nil))
    (reader-skip)
    (while (!= (reader-peek) ")")
      (set! items
            (cons
             (reader-read)
             items))
      (reader-skip))
    ;; consume ')'
    (reader-next)
    (reverse items)))

(defun reader-read-string ()
  ;; consume opening "
  (reader-next)
  (let ((text ""))
    (while
        (and
         (!= (reader-peek) "")
         (!= (reader-peek) "\""))
      (set! text
            (strcat
             text
             (reader-next))))

    ;; consume closing "
    (if (= (reader-peek) "\"")
        (reader-next))
    text))

(defun reader-read-atom ()
  (let ((token ""))
    (while
        (and
         (!= (reader-peek) "")
         (!= (reader-peek) " ")
         (!= (reader-peek) "\t")
         (!= (reader-peek) "\r")
         (!= (reader-peek) "\n")
         (!= (reader-peek) "(")
         (!= (reader-peek) ")"))

      (set! token
            (strcat
             token
             (reader-next))))

    ;; Is this a number?
    (if (number-string? token)
        (atof token)
        (symbol token))))


;; parse a complete program and return all the expressions within it.
(defun reader-parse-program ()
  (let ((forms nil))
    (reader-skip)
    (while (not (reader-eof))
      (set! forms
            (cons
             (reader-read)
             forms))
      (reader-skip))
    (reverse forms)
    ))

;; given a list of expressions, evaluate them all
(defun eval-program (forms)
  (let ((result nil))
    (while forms
      (set! result (eval (car forms) nil))
      (set! forms  (cdr forms)))
    result))

;; Evaluate a program comprised of expressions
(defun run-program (text)
  (reader-init text)
  (eval-program (reader-parse-program)))


;; REPL code
(defun repl ()
  (println "Welcome to lisp in slisp!")
  (println "Enter :quit to exit.")
  (println "")

  (let ((run t))
    (while run
      (print "> ")
      (let ((line (read-line)))
        (if (= line ":quit")
            (set! run nil)
            (let ((result
                   (repl-execute-line line)))
              (if result
                  (println result))))))))

(defun repl-execute-line (text)
  (reader-init text)
  (eval-program (reader-parse-program)))


;;
;; Entry point.
;;
(defun main (args)
  "Entry-Point, either start the REPL or process each named file."

  ;; no args?  show error and terminate
  (if (= (length args) 1 )
      (do
       (println "Usage " (car args) " --repl | path/to/run")
       (exit 1)))

  ;; We process each named file (skipping --repl)
  (map (lambda (name)
         (if (!= name "--repl")
             (do
              (println "Loading .. " name)
              (let ((handle (fopen name "r"))  ; open
                    (data   (fread handle))    ; read
                    (res    (fclose handle)))  ; close
                (if data
                    (run-program data))))))
         (cdr args))


  ;; Is this REPL mode?  Then run it
  (if (member? args "--repl")
      (do
       (repl)
       (exit 0)))
)
