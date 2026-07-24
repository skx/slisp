;; A minimal lisp interpreter, meant to be compiled with slisp.
;;
;; Features match those of the parent compiler, so we have strings, floats,
;; integers, lambdas, characters, etc.
;;
;; In some ways the interpreter is more advanced than the compiler as it
;; has proper symbols and the (quote..) special form.  The downside is
;; that it is slower in the execution of programs.
;;
;; Note that we store built-in primitives, variables, and functions
;; in three different namespaces.  We do allow `alias!` to remap
;; a function, and of course defining the same function name a second
;; time will also overwrite the previous version.
;;


;;
;; lisp-reader.lisp contains our reader/parser code.
;;
;; The source of this lives in ../packages/lisp-reader.lisp, however
;; it is also bundled into the runtime and available that way.
;;
(require lisp-reader)


;;
;; tree.lisp contains a simple AVL-tree library.
;;
;; The source of this lives in ../packages/tree.lisp, however
;; it is also bundled into the runtime and available that way.
;;
(require tree)


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

(defun character? (x)
  (and (cons? x)
       (= (car x) "character")))

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

;; we tag characters with "character", but the content is actually a string.
;; get the first character and return it.
(defun character-value (x)
  (car (explode (cadr x)))) ; horrid

(defun builtin-fn (x)
  (cadr x))

;; get the names of parameters
(defun symbols-names (lst)
  (if (nil? lst)
      nil
      (cons
       (symbol-name (car lst))
       (symbols-names (cdr lst)))))



;; Contains references to built-in functions
(defvar *builtins* nil)

;; global storage for user-functions
(defvar *functions*  nil)

;; global storage for global variables
(defvar *globals* nil)

(defun global-get (name)
  (tree:get *globals* name))

(defun global-set (name value)
  (set! *globals* (tree:put *globals* name value)))



;; Register a new built-in
(defun register-builtin (name fn)
  (set! *builtins* (tree:put *builtins* name (builtin fn))))

;; Lookup a builtin function.
(defun lookup-builtin (name)
  (tree:get *builtins* name))

;; add a function
(defun add-function (name params body)
  (set! *functions*
        (tree:put
         *functions*
         name
         (closure params body nil)))
  ;; return the name
  name)

;; get a function, by name
;;
;; Check for the self-hosted/inception version of the entry first.
(defun lookup-function (name)
  (let ((nested
          (tree:get
             (global-get "*functions*")
             name)))
    (if nested
        nested
        (tree:get *functions* name))))

;; lookup a binding from the environment
(defun env-get (env name)
  (tree:get env name))

(defun env-bound? (env name)
  (tree:bound? env name))

;; set a variable in the environment
(defun env-set (env name value)
  (set! env (tree:put env name value)))

(defun env-update (env name value)
  (set! env (tree:put env name value)))


;; Is the given argument name variadic (i.e prefixed with "&")
(defun variadic-arg? (name)
  (if (> (strlen name) 0)
      (= (substr name 0 1) "&")))

;; Get the name of the argument, having removed the leading &-prefix
(defun variadic-name (name)
  (substr name 1 (- (strlen name) 1)))

;; eval will return a list: (return-value updated-environment)
;; get the value.
(defun eval-value (x)
  (car x))

;; eval will return a list: (return-value updated-environment)
;; get the environment
(defun eval-env (x)
  (cadr x))

(defun init-builtins ()
  (set! *builtins* nil)
  (register-builtin "char?" (lambda (args) (char? (car args))))
  (register-builtin "cons?" (lambda (args) (cons? (car args))))
  (register-builtin "float?" (lambda (args) (float? (car args))))
  (register-builtin "int?" (lambda (args) (int? (car args))))
  (register-builtin "lambda?" (lambda (args) (lambda? (car args))))
  (register-builtin "list" (lambda (args) args))
  (register-builtin "nil?" (lambda (args) (nil? (car args))))
  (register-builtin "printfloat" (lambda (args) (printfloat (car args))))
  (register-builtin "printint" (lambda (args) (printint (car args))))
  (register-builtin "printstr" (lambda (args) (printstr (car args))))
  (register-builtin "str?" (lambda (args) (str? (car args))))
  (register-builtin "string" (lambda (args) (string (car args))))
  (register-builtin "sys-gc" (lambda (args) (sys-gc)))
  (register-builtin "sys-heap-allocs" (lambda (args) (sys-heap-allocs)))
  (register-builtin "sys-heap-bytes" (lambda (args) (sys-heap-bytes)))
  (register-builtin "sys-heap-data" (lambda (args) (sys-heap-data)))
  (register-builtin "sys-heap-dump" (lambda (args) (sys-heap-dump)))
  (register-builtin "sys-heap-objects" (lambda (args) (sys-heap-objects)))
  (register-builtin "sys_%" (lambda (args) (sys_% (car args) (cadr args))))
  (register-builtin "sys_<" (lambda (args) (sys_< (car args) (cadr args))))
  (register-builtin "sys_<=" (lambda (args) (sys_<= (car args) (cadr args))))
  (register-builtin "sys_=" (lambda (args) (sys_= (car args) (cadr args))))
  (register-builtin "sys_>" (lambda (args) (sys_> (car args) (cadr args))))
  (register-builtin "sys_>=" (lambda (args) (sys_>= (car args) (cadr args))))
  (register-builtin "sys_car" (lambda (args) (sys_car (car args))))
  (register-builtin "sys_cdr" (lambda (args) (sys_cdr (car args))))
  (register-builtin "sys_chr" (lambda (args) (sys_chr (car args))))
  (register-builtin "sys_cons" (lambda (args) (sys_cons (car args) (cadr args))))
  (register-builtin "sys_divide" (lambda (args) (sys_divide (car args) (cadr args))))
  (register-builtin "sys_entries" (lambda (args) (sys_entries (car args))))
  (register-builtin "sys_environment" (lambda (args) (sys_environment)))
  (register-builtin "sys_exit" (lambda (args) (sys_exit (car args))))
  (register-builtin "sys_explode" (lambda (args) (sys_explode (car args))))
  (register-builtin "sys_fclose" (lambda (args) (sys_fclose (car args))))
  (register-builtin "sys_fopen" (lambda (args) (sys_fopen (car args) (cadr args))))
  (register-builtin "sys_fread" (lambda (args) (sys_fread (car args))))
  (register-builtin "sys_fwrite" (lambda (args)  (sys_fwrite (car args) (cadr args) (caddr args))))
  (register-builtin "sys_getc" (lambda (args) (sys_getc)))
  (register-builtin "sys_implode" (lambda (args) (sys_implode (car args))))
  (register-builtin "sys_int" (lambda (args) (sys_int (car args))))
  (register-builtin "sys_isqrt" (lambda (args) (sys_isqrt (car args))))
  (register-builtin "sys_minus" (lambda (args) (sys_minus (car args) (cadr args))))
  (register-builtin "sys_mkdir" (lambda (args) (sys_mkdir (car args))))
  (register-builtin "sys_multiply" (lambda (args) (sys_multiply (car args) (cadr args))))
  (register-builtin "sys_newline" (lambda (args) (sys_newline)))
  (register-builtin "sys_not" (lambda (args) (sys_not (car args))))
  (register-builtin "sys_now" (lambda (args) (sys_now)))
  (register-builtin "sys_nth!" (lambda (args) (sys_nth! (car args) (cadr args) (caddr args))))
  (register-builtin "sys_nth" (lambda (args) (sys_nth (car args) (cadr args))))
  (register-builtin "sys_ord" (lambda (args) (sys_ord (car args))))
  (register-builtin "sys_package" (lambda (args) (sys_package (car args))))
  (register-builtin "sys_packages" (lambda (args) (sys_packages )))
  (register-builtin "sys_plus" (lambda (args) (sys_plus (car args) (cadr args))))
  (register-builtin "sys_putc" (lambda (args) (sys_putc (car args))))
  (register-builtin "sys_random" (lambda (args) (sys_random (car args))))
  (register-builtin "sys_rmdir" (lambda (args) (sys_rmdir (car args))))
  (register-builtin "sys_run" (lambda (args) (sys_run (car args) (cadr args))))
  (register-builtin "sys_split" (lambda (args) (sys_split (car args) (cadr args))))
  (register-builtin "sys_sqrt" (lambda (args) (sys_sqrt (car args))))
  (register-builtin "sys_stat" (lambda (args) (sys_stat (car args))))
  (register-builtin "sys_stdlib" (lambda (args) (sys_stdlib)))
  (register-builtin "sys_strcat" (lambda (args) (sys_strcat (car args) (cadr args))))
  (register-builtin "sys_strcmp" (lambda (args) (sys_strcmp (car args) (cadr args))))
  (register-builtin "sys_strlen" (lambda (args) (sys_strlen (car args))))
  (register-builtin "sys_substr" (lambda (args) (sys_substr (car args) (cadr args) (caddr args))))
  (register-builtin "sys_unlink" (lambda (args) (sys_unlink (car args))))
)

;; eval: where the magic happens.
;;
;; NOTE: This returns a list of "return value" and "[possibly updated] environment".
(defun eval (expr env)
  (cond
    ;; integers evaluate to themselves
    ((int? expr) (list expr env))

    ;; floats are self-evaluating too.
    ((float? expr) (list expr env))

    ;; strings are self-evaluating too.
    ((str? expr) (list expr env))

    ;; characters are self-evaluating too - but in this case they look like strings
    ((character? expr) (list (character-value expr) env))

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
    result))


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

      ;; Get the name of the thing we're calling
      (let ((call-name
        (if (symbol? (car expr))
            (symbol-name (car expr))
            "<lambda>")))

        ;; make the call
        (list (apply fn args call-name) env)))))


;; special form: alias!
;;
;; Two forms:
;; "(alias! foo    bar)" will receive ((symbol alias!) (symbol foo) (symbol bar))
;; "(alias! "foo" "bar") will be ((symbol alias!) foo bar)
;;
;; The former form is obvious and what we expect, but the latter form is what our compiler
;; will generate.
(defun eval-alias (expr env)
  (if (cons? (caddr expr))
      (eval-alias-references expr env)                  ; default two symbosl
      (eval-alias-references (list                      ; strings - build up a list
                              (symbol "alias!")
                              (symbol (cadr expr))
                              (symbol (caddr expr)))
                             env)))

(defun eval-alias-references (expr env)
  (let ((old (symbol-name (cadr expr)))
        (new (symbol-name (caddr expr)))
        (fn  (lookup-function new)))
    (if fn
        (do
         (set! *functions* (tree:put *functions* old fn))
         (list old env))
        (do
         (println "Unknown function " new)
         (list nil env)))))

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
      ((= op "alias!")   (eval-alias expr env))
      ((= op "and")      (eval-and expr env))
      ((= op "cond")     (eval-cond expr env))
      ((= op "defconst") (eval-defvar expr env))
      ((= op "defun")    (eval-defun expr env))
      ((= op "defvar")   (eval-defvar expr env))
      ((= op "do")       (eval-do expr env))
      ((= op "if")       (eval-if expr env))
      ((= op "lambda")   (eval-lambda expr env))
      ((= op "let")      (eval-let expr env))
      ((= op "or")       (eval-or expr env))
      ((= op "quote")    (eval-quote expr env))
      ((= op "require")  (eval-require expr env))
      ((= op "set!")     (eval-set expr env))
      ((= op "unless")   (eval-unless expr env))
      ((= op "when")     (eval-when expr env))
      ((= op "while")    (eval-while expr env))

      ;; default
      (t  (eval-call expr env)))))

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

    (cond
      ;; boolean constants
      ((= name "t")       t)
      ((= name "nil")     nil)

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
                     (let ((user (lookup-function name)))
                       (if user
                           user
                         (do (println "Unknown function " name) nil)))))))))))

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
  (list (cadr expr) env))


(defun require-path (file)
  "Find the given file on LISP_PATH, if possible"
  (let ((path (getenv "LISP_PATH")))
    (if path
        (let ((split (split-all path #\:))
              (res (filter split (lambda (dir) (exists? (join (list dir "/" file)))))))
          (if res
              (join (list (car res) "/" file))))
        file)))

(defun require-filename (arg env)
  (cond
    ;; literal string
    ((str? arg)
      arg)

    ;; symbol
    ((symbol? arg)
      (let ((name (symbol-name arg)))
        (if (or (env-bound? env name)
                (env-bound? *globals* name))
            (eval-symbol arg env)
            name)))

    ;; expression
    (t
      (eval-value (eval arg env)))))

;; special form: require
;;
;; Load a file from the embedded asset, if we can.  Otherwise we
;; append ".lisp" to files missing it and search LISP_PATH for them.
(defun eval-require (expr env)
  (let ((filename (require-filename (cadr expr) env))
        (data nil))
    (set! data (package filename))
    (if data
        (run-program data)
        (do
          (if (strstr filename ".")
              (set! filename (require-path filename))
              (set! filename (require-path (strcat filename ".lisp"))))
          (if filename
              (if (exists? filename)
                  (execute-file filename)
                  (println "File not found " filename)))))
    (list nil env)))


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
(defun apply (fn args name)
  (cond
    ;; native builtins
    ((builtin? fn) ((builtin-fn fn) args))

    ;; user functions and lambdas
    ((closure? fn) (apply-closure fn args name))

    ;; nothing known
    (t nil)))


(defun apply-closure (closure args call-name)
  (let ((params (cadr closure))
        (body   (caddr closure))
        (env    (cadddr closure)))

    (while params
      (if (variadic-arg? (car params))
          ;; Bind all remaining arguments.
          (do
            (set! env
                  (env-set env
                           (variadic-name (car params))
                           args))
            (set! args nil)
            (set! params nil))

          ;; Ordinary argument - Too few arguments?
          (if (nil? args)
              (do
               (println call-name ": too few arguments supplied")
               (set! env (env-set env (car params) nil)) ; missing arg is nil
                (set! params nil))
              (do
               (set! env
                     (env-set env
                              (car params)
                              (car args)))
               (set! params (cdr params))
                (set! args   (cdr args))))))

    ;; Too many arguments?
    (if args (println call-name ": too many arguments supplied"))

    (let ((result (eval-body body env)))
      (nth! closure 3 (eval-env result))
      (eval-value result))))

;; given a list of expressions, evaluate them all
(defun eval-program (forms)
  (let ((result nil))
    (while forms
      (set! result (eval (car forms) nil))
      (set! forms  (cdr forms)))
    result))


;; Evaluate a program comprised of expressions
(defun run-program (text)
  (init-builtins)
  (reader:init text)
  (let ((forms (reader:parse-program)))
    (eval-program forms)))



;;
;; REPL code
;;
(defun repl ()
  (println "Welcome to lisp in slisp!")
  (println "Enter :quit to exit.")
  (println "")

  (init-builtins)

  (let ((run t))
    (while run
      (print "> ")
      (let ((line (read-line-sexp)))
        (if (= line ":quit")
            (set! run nil)
            (let ((result (repl-execute-line line)))
              (println (car result))))))))


(defun repl-execute-line (text)
  (reader:init text)
  (eval-program (reader:parse-program)))


;;
;; Load File code
;;
(defun execute-file(name)
  (println "Loading .. " name)
  (let ((handle (fopen name "r"))  ; open
        (data   (fread handle))    ; read
        (res    (fclose handle)))  ; close
    (if data
        (run-program data)
        (println "Failed to read file"))))


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

  ;; Load the standard library
  (let ((before (now))
        (x (run-program (stdlib)))
        (after (now)))
    ;; Loading time will vary, so we should exclude from tests
    (if (not (getenv "TEST"))
        (println "Loaded stdlib.lisp in \e[1m" (- after before) "ms\e[0m.")))

  ;; We process each named file (skipping --repl)
  (map (lambda (name)
         (if (and (!= name "--repl")
                  (!= name "--main"))
             (execute-file name)))
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
