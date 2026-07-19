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
;;
;;   AND
;;   COND
;;   DEFUN
;;   DEFCONST
;;   DEFVAR
;;   DO
;;   IF
;;   LAMBDA
;;   LET
;;   OR
;;   QUOTE
;;   SET!
;;   UNLESS
;;   WHEN
;;   WHILE
;;
;; We have a couple of builtin-functions we handle specially, but most are deferred to the
;; host/compiler's version of the code.
;;



;; Our standard-library provides "(read-line)" but that
;; terminates on newline.  We want to read a complete
;; multi-line sexp for REPL usage.
(defun read-line-sexp ()
  "Read a complete SEXP from STDIN."
  (let ((text (read-line))
        (depth 0)
        (line ""))

    (set! depth (paren-depth text))

    (while (> depth 0)
      (set! line (read-line))
      (set! text (strcat text line))
      (set! depth (+ depth (paren-depth line))))

    text))

;; Count depth for our read-line-sexp helper
(defun paren-depth (text)
  "Get the depth of the given text, by counting '(' and ')'."
  (let ((depth 0)
        (chars (explode text))
        (len   (length text))
        (i     0))
    (while (< i len)
      (let ((ch (nth chars i)))
        (cond
          ((= ch #\() (set! depth (+ depth 1)))
          ((= ch #\)) (set! depth (- depth 1))))
        (set! i (+ i 1))))
    depth))

(defvar *DEBUG* nil)


;; Since we don't "tag" types as we do in our compiler instead we wrap them in lists, and identify
;; them via a string-compare of the first element.
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
  (if (env-bound? *globals* name)
      (set! *globals*
            (env-update *globals* name value))
      (set! *globals*
            (env-set *globals* name value))))


;; Lookup a builtin function.
(defun lookup-builtin (name)
  (cond
    ((= name "+")        (builtin "+"))
    ((= name "%")        (builtin "%"))
    ((= name "-")        (builtin "-"))
    ((= name "*")        (builtin "*"))
    ((= name "/")        (builtin "/"))
    ((= name "<")        (builtin "<"))
    ((= name "<=")       (builtin "<="))
    ((= name ">")        (builtin ">"))
    ((= name ">=")       (builtin ">="))
    ((= name "=")        (builtin "="))
    ((= name "abs")      (builtin "abs"))
    ((= name "cons")     (builtin "cons"))
    ((= name "car")      (builtin "car"))
    ((= name "cdr")      (builtin "cdr"))
    ((= name "chr")      (builtin "chr"))
    ((= name "exit")     (builtin "exit"))
    ((= name "getc")     (builtin "getc"))
    ((= name "getenv")   (builtin "getenv"))
    ((= name "length")   (builtin "length"))
    ((= name "list")     (builtin "list"))
    ((= name "map")      (builtin "map"))
    ((= name "nat")      (builtin "nat"))
    ((= name "newline")  (builtin "newline"))
    ((= name "not")      (builtin "not"))
    ((= name "nth")      (builtin "nth"))
    ((= name "nth!")     (builtin "nth!"))
    ((= name "ord")      (builtin "ord"))
    ((= name "print")    (builtin "print"))
    ((= name "println")  (builtin "println"))
    ((= name "random")   (builtin "random"))
    ((= name "repeated") (builtin "repeated"))
    ((= name "reverse")  (builtin "reverse"))
    ((= name "seq")      (builtin "seq"))
    ((= name "strcmp")   (builtin "strcmp"))
    ((= name "strlen")   (builtin "strlen"))
    ((= name "substr")   (builtin "substr"))
    ((= name "sys-heap-allocs")  (builtin "sys-heap-allocs"))
    ((= name "sys-heap-bytes")   (builtin "sys-heap-bytes"))
    ((= name "sys-heap-dump")   (builtin "sys-heap-dump"))
    ((= name "sys-heap-objects") (builtin "sys-heap-objects"))
    ((= name "sys-gc")   (builtin "sys-gc"))
    ((= name "nil?")     (builtin "nil?"))
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

;; is a binding present?
(defun env-bound? (env name)
  (cond
    ((nil? env)           nil)

    ((= (caar env) name)  t)

    (t (env-bound? (cdr env) name))))

;; set a variable in the environment
(defun env-set (env name value)
  (cons (list name value) env))

(defun env-update (env name value)
  (cond
    ((nil? env)
     nil)
    ((= (caar env) name)
     (cons
      (list name value)
      (cdr env)))
    (t
     (cons
      (car env)
      (env-update
       (cdr env)
       name
       value)))))

;; eval will return a list: (return-value updated-environment)
;; get the value.
(defun eval-value (x)
  (car x))

;; eval will return a list: (return-value updated-environment)
;; get the environment
(defun eval-env (x)
  (cadr x))

;; eval: where the magic happens.
;;
;; NOTE: This returns a list of "return value" and "[possibly updated] environment".
(defun eval (expr env)
  (if *DEBUG* (println "expr:" expr))
  (cond
    ;; integers evaluate to themselves
    ((int? expr) (list expr env))

    ;; floats are self-evaluating too.
    ((float? expr) (list expr env))

    ;; strings are self-evaluating too.
    ((str? expr) (list expr env))

    ;; symbols will be looked up
    ((symbol? expr) (list (eval-symbol expr env) env))

    ;; lists are expressions
    ((cons? expr) (eval-list expr env))

    ;; something unknown
    (t (list nil env))))

;; helper which is used by apply, do, and let.
(defun eval-body (forms env)
  (let ((result nil))
    (while forms
      (set! result (eval (car forms) env))
      (set! env (eval-env result))
      (set! forms (cdr forms)))
    (sys-gc)
    result))


;; function-call, lambda, builtin, etc.
(defun eval-call (expr env)
  (let ((fn-result (eval (car expr) env)))

    (let ((fn    (car fn-result))   ; eval returns "[RESULT ENV]" - get the [callable] result
          (env   (cadr fn-result))  ; eval returns "[RESULT ENV]" - get the environment
          (args  nil)
          (forms (cdr expr)))

      ;; evaluate arguments left-to-right
      (while forms

        (let ((result (eval (car forms) env)))
          (set! args (append args (list (car result))))
          (set! env  (cadr result)))
        (set! forms (cdr forms)))

      ;; make the call, and return the result of apply, and the env, as a list
      (list (apply fn args) env))))

;; special form: and
(defun eval-and (expr env)
  (eval-and-forms (cdr expr) env))

(defun eval-and-forms (forms env)
  (if (nil? forms)

      ;; (and) => t
      (list t env)

      (let ((result (eval (car forms) env)))

        (if (nil? (eval-value result))

            ;; first false value
            result

            ;; last value wins
            (if (nil? (cdr forms))
                result
                (eval-and-forms
                 (cdr forms)
                 (eval-env result)))))))

;; special form: cond
(defun eval-cond (expr env)
  (eval-cond-clauses (cdr expr) env))

(defun eval-cond-clauses (clauses env)
  (if (nil? clauses)
      (list nil env)

      (let ((clause (car clauses)))

        ;; Evaluate the test expression.
        (let ((result (eval (car clause) env)))

          (if (eval-value result)

              ;; Test succeeded.
              (eval-body
               (cdr clause)
               (eval-env result))

              ;; Try the next clause.
              (eval-cond-clauses
               (cdr clauses)
               (eval-env result)))))))

;; special form: defun
(defun eval-defun (expr env)
  (add-function
   (symbol-name (cadr expr))    ; name
   (symbols-names (caddr expr)) ; params
   (cdddr expr))                ; body

  ;; return the name of the defun, and the environment
  (list (cadr expr) env))

;; special form: defvar - return the value
(defun eval-defvar (expr env)
  (let ((name (symbol-name (cadr expr)))    ; get the name
        (result (eval (caddr expr) env)))   ; get the result of the expression

    ;; remember the result of eval is a list, so use "eval-value" to get the
    ;; actual value
    (set! *globals* (env-set *globals* name (eval-value result)))

    ;; return list of "value" "env"
    (list (eval-value result) (eval-env result))))

;; special form: do - run each statement in the list
(defun eval-do (expr env)
  (eval-body (cdr expr) env))

;; special form; if
(defun eval-if (expr env)
  (let ((r (eval (cadr expr) env)))
    (if (eval-value r)
        (eval (caddr expr) (eval-env r))
        (eval (cadddr expr) (eval-env r)))))

;; evaluate a lambda
(defun eval-lambda (expr env)
  ;; return a list of result + env
  (list
   (closure
    (symbols-names (cadr expr)) ; params
    (cddr expr)                 ; body
    env) ; result
   env))

;; special form: let
(defun eval-let (expr env)
  (let ((bindings (cadr expr))
        (body     (cddr expr))
        (new-env  env))

    (while bindings
      (let ((result
             (eval (cadr (car bindings))
                   new-env)))
        (set! new-env
              (env-set
               new-env
               (symbol-name (car (car bindings)))
               (eval-value result))))
      (set! bindings (cdr bindings)))
    (eval-body body new-env)))

;; evaluate a list - handling special forms via our own code
(defun eval-list (expr env)
  (let ((op (symbol-name (car expr))))
    (cond
      ((= op "and")
       (eval-and expr env))
      ((= op "cond")
       (eval-cond expr env))
      ((= op "defun")
       (eval-defun expr env))
      ((= op "defvar")
       (eval-defvar expr env))
      ((= op "defconst")
       (eval-defvar expr env))
      ((= op "do")
       (eval-do expr env))
      ((= op "if")
       (eval-if expr env))
      ((= op "lambda")
       (eval-lambda expr env))
      ((= op "let")
       (eval-let expr env))
      ((= op "or")
       (eval-or expr env))
      ((= op "quote")
       (eval-quote expr env))
      ((= op "set!")
       (eval-set expr env))
      ((= op "unless")
       (eval-unless expr env))
      ((= op "when")
       (eval-when expr env))
      ((= op "while")
       (eval-while expr env))
      (t
       (eval-call expr env)))))

;; eval set!
(defun eval-set (expr env)
  (let ((name
         (symbol-name (cadr expr))))

    (let ((result
           (eval (caddr expr) env)))

      (let ((value
             (car result))
            (env
             (cadr result)))

        (if (env-bound? env name)
            (list value
                  (env-update env name value))
            (do
             (global-set name value)
             (list value env)))))))

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

       ;; env
       (if (env-bound? env name)
           (env-get env name)

           ;; global
           (if (env-bound? *globals* name)
               (env-get *globals* name)

               ;; builtin
               (let ((fn (lookup-builtin name)))
                 (if fn
                     fn
                     ;; user-function
                     (lookup-function name)))))))))

;; special form: or
(defun eval-or (expr env)
  (eval-or-forms (cdr expr) env))

(defun eval-or-forms (forms env)
  (if (nil? forms)

      ;; (or) => nil
      (list nil env)

      (let ((result (eval (car forms) env)))

        (if (eval-value result)

            ;; first true value
            result

            ;; otherwise continue
            (eval-or-forms
             (cdr forms)
             (eval-env result))))))

;; special form: quote
(defun eval-quote (expr env)
  (list
   (cadr expr)
   env))

;; special form: unless
(defun eval-unless (expr env)
  (let ((r (eval (cadr expr) env)))
    (if (eval-value r)
        nil
        (eval-body (cdr expr) env))))

;; special form: while
(defun eval-when (expr env)
  (let ((r (eval (cadr expr) env)))
    (if (eval-value r)
        (eval-body (cdr expr) env))))

;; special form: while
(defun eval-while (expr env)
  (let ((running t)
        (result (list nil env)))
    (while running
      (let ((cond-result (eval (cadr expr) env)))

        (set! env (eval-env cond-result))
        (if (eval-value cond-result)
            (do
             (set! result
                   (eval-body (cddr expr) env))
             (set! env
                   (eval-env result)))
            (set! running nil))))
    result))


;; apply for built-in and user-functions
(defun apply (fn args)

  (cond
    ;; native builtins
    ((builtin? fn) (apply-builtin fn args))

    ;; user functions and lambdas
    ((closure? fn) (apply-closure fn args))

    ;; nothing known
    (t nil)))


(defun builtin-map (fn lst)
  (if (nil? lst)
      nil
      (cons
       (apply fn (list (car lst)))
       (builtin-map fn (cdr lst)))))

;; built-in functions all live here, which is a bit horrid.
;;
;; 99% of these are deferred to the host.  map/print/println are the only obvious exceptions.
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

      ((= name "chr")
       (chr (car args)))

      ((= name "exit")
       (exit (car args)))

      ((= name "getc")
       (getc))

      ((= name "getenv")
       (getenv (car args)))

      ((= name "length")
       (length (car args)))

      ((= name "list")
       args)

      ((= name "map")
       (builtin-map (car args) (cadr args)))

      ((= name "nat")
       (nat (car args)))

      ((= name "newline")
       (newline))

      ((= name "not")
       (not (car args)))

      ((= name "nth")
       (nth (car args) (cadr args)))

      ((= name "nth!")
       (nth! (car args) (cadr args) (caddr args)))

      ((= name "ord")
       (ord (car args)))

      ((= name "random")
       (random (car args)))

      ((= name "repeated")
       (repeated (car args) (cadr args)))

      ((= name "reverse")
       (reverse (car args)))

      ((= name "print")
        (while args
          (print (car args))
          (set! args (cdr args)))
        nil)

      ((= name "println")
        (while args
          (print (car args))
          (set! args (cdr args)))
        (print "\n")
        nil)

      ((= name "seq")
       (seq (car args)))

      ((= name "strcmp")
       (strcmp (car args) (cadr args)))

      ((= name "strlen")
       (strlen (car args)))

      ((= name "substr")
       (substr (car args) (cadr args) (caddr args)))

      ((= name "sys-heap-allocs")
       (sys-heap-allocs))

      ((= name "sys-heap-bytes")
       (sys-heap-bytes))

      ((= name "sys-heap-dump")
       (sys-heap-dump))

      ((= name "sys-heap-objects")
       (sys-heap-objects))

      ((= name "sys-gc")
       (sys-gc))

      ((= name "nil?")
       (nil? (car args)))

      (t nil))))


(defun apply-closure (closure args)
  (let ((params (cadr closure))
        (body   (caddr closure))
        (env    (cadddr closure)))

    ;; Extend the captured environment with arguments.
    (while params
      (set! env
            (env-set env
                     (car params)
                     (car args)))
      (set! params (cdr params))
      (set! args   (cdr args)))

    (let ((result (eval-body body env)))

      ;; Save the updated environment back into the closure.
      (nth! closure 3 (eval-env result))

      (eval-value result))))


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
      (symbol "quote")
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
  (reader-next)   ; consume opening "

  (let ((text ""))

    (while
        (and
         (!= (reader-peek) "")
         (!= (reader-peek) "\""))

      (if (= (reader-peek) "\\")
          (do
           (reader-next)     ; consume '\'

           (let ((ch (reader-next)))
             (set! text
                   (strcat
                    text
                    (cond
                      ((= ch "n") "\n")
                      ((= ch "t") "\t")
                      ((= ch "\"") "\"")
                      ((= ch "\\") "\\")
                      (t ch))))))
          (set! text
                (strcat
                 text
                 (reader-next)))))

    (reader-next)    ; closing "
    text))

(defun reader-read-string-old ()
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
      (if *DEBUG* (println "FORM " (car forms)))
      (set! result (eval (car forms) nil))
      (set! forms  (cdr forms)))
    result))

;; Evaluate a program comprised of expressions
(defun run-program (text)
  (reader-init text)
  (let ((forms (reader-parse-program)))
    (if *DEBUG* (println "Parsed: " forms))
    (eval-program forms)))


;; REPL code
(defun repl ()
  (println "Welcome to lisp in slisp!")
  (println "Enter :quit to exit.")
  (println "")

  (let ((run t))
    (while run
      (print "> ")
      (let ((line (read-line-sexp)))
        (if (= line ":quit")
            (set! run nil)
            (let ((result
                   (repl-execute-line line)))
              (println (car result))))))))

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
         (if (and (!= name "--repl")
                  (!= name "--main"))
             (do
              (println "Loading .. " name)
              (let ((handle (fopen name "r"))  ; open
                    (data   (fread handle))    ; read
                    (res    (fclose handle)))  ; close
                (if data
                    (run-program data))))))
       (cdr args))

  ;; Should we auto-run (defun main) ..?
  (if (member? args "--main")
      (repl-execute-line "(main)"))

  ;; Is this REPL mode?  Then run it
  (if (member? args "--repl")
      (do
       (repl)
       (exit 0)))
  )
