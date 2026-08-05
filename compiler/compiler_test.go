package compiler

import (
	"strings"
	"testing"
)

func TestBasic(t *testing.T) {

	c := New(`
(defun foo (a b)
  "This is a demo"
  (* a b))

(defun counter ()
  "Counter returns a function which will return an incrementing number every time it is called."
  (let ((n 0))
    (lambda ()
      (do
	(set! n (+ n 1))
	n))))

(defun maths ()
  (print (= 0 0))
  (print (! 0))
  (print (<= 0 0))
  (print (< 0 0))
  (print (> 0 0))
  (print (>= 0 0))
  (print (+ 0 0))
  (print (- 0 0))
  (print (* 0 0))
  (print (/ 0 0))
  (print (% 0 0))
  (print (cons? 0))
  (print (char? 0))
  (print (int? 0))
  (print (lambda? 0))
  (print (nil? 0))
  (print (str?  0)))

(defun main ()
  (let ((f (counter)))
    (println (f))
    (println (f))
  )
  (if 1 (print "OK") (print "fail"))
  (if 1 (print "OK"))

  (let ((x 1))
     ;; x = 1
     (printint x)
     (newline)

     ;; mutate
     (set! x 42)

     ;; confirm it worked
     (printint x)
     (newline))

  (foo 32 11)
  (print (cons 1 (cons 2 (cons 3 nil))))
  (print ( (lambda (x) 3) 3))
  (do
    (print 1)
    (print #\x)
    (print 2))
)
		`)

	out, err := c.Compile()
	if err != nil {
		t.Fatalf("failed to compile %s", err)
	}
	if !strings.Contains(out, "call fn_main") {
		t.Fatalf("compilation looks bogus")
	}

}

func TestErrors(t *testing.T) {
	tests := []string{
		`(defun main () ( (set! foo 3)))`,
		`(defun main () ( (do (set! foo 3))))`,
		`(defun main () ( (set! foo 3)))`,
		`(defun main () ( (foo bar)))`,
		`(defun main () ( (if foo 1 2)))`,
		`(defun main () ( (if 1 (foo bar) 1)))`,
		`(defun main () ( (if 1 1 (foo bar))))`,
		`(defun main () ( (if 1 (foo bar))))`,
		`(defun main () ( (let ((a foo)) 1)))`,
		`(defun main () ( (let ((x 3)) y)))`,
		`(defun main () ((let ((x 3)) (set! x y))))`,
	}

	for _, tst := range tests {
		c := New(tst)

		_, err := c.Compile()
		if err == nil {
			t.Fatalf("expected error, got none %s", tst)
		}
	}
}

// TestQuote confirms that quote/quasiquote/unquote-splicing, appearing
// in ordinary (non-macro) code, compile successfully.
func TestQuote(t *testing.T) {
	c := New(`
(defun main ()
  (print 'hello)
  (print '(1 2 3))
  (print '())
  (print '((1 2) (3 4)))
  (let ((n 1))
    (print ` + "`(a ,(+ n 1) ,@(cons 3 (cons 4 nil)) b))))" + `
`)

	out, err := c.Compile()
	if err != nil {
		t.Fatalf("failed to compile %s", err)
	}
	if !strings.Contains(out, "call fn_main") {
		t.Fatalf("compilation looks bogus")
	}
}

// TestQuoteInert confirms that unquote/unquote-splicing appearing within
// a plain (non-quasi) quote are inert: they become literal data, rather
// than being evaluated.
func TestQuoteInert(t *testing.T) {
	c := New(`(defun main () (print '(1 ,foo 2)) (print '(1 ,@foo)))`)

	out, err := c.Compile()
	if err != nil {
		t.Fatalf("failed to compile %s", err)
	}
	if !strings.Contains(out, "call fn_main") {
		t.Fatalf("compilation looks bogus")
	}
}

// TestQuoteErrors confirms that misuse of unquote/unquote-splicing is
// rejected.
func TestQuoteErrors(t *testing.T) {
	tests := []string{
		// unquote outside of a quasiquote.
		`(defun main () (print ,foo))`,
		// unquote-splicing outside of a quasiquote.
		`(defun main () (print ,@foo))`,
		// unquote-splicing directly, rather than as a list-element.
		`(defun main () (print ` + "`,@foo))" + `)`,
	}

	for _, tst := range tests {
		c := New(tst)
		_, err := c.Compile()
		if err == nil {
			t.Fatalf("expected error, got none: %s", tst)
		}
	}
}

// TestMacro confirms that simple, and variadic, macros expand and
// compile correctly - including recursive expansion (a macro whose
// expansion calls another macro).
func TestMacro(t *testing.T) {
	c := New(`
(defmacro cond (&clauses)
  (if (nil? clauses)
      nil
      ` + "`(if ,(car (car clauses))" + `
           (do ,@(cdr (car clauses)))
           (cond ,@(cdr clauses)))))

(defmacro my-if (c then else)
  ` + "`(cond (,c ,then) (t ,else)))" + `

(defmacro my-unless (c &body)
  ` + "`(if ,c nil (do ,@body)))" + `

(defmacro my-when-not (c &body)
  ` + "`(my-unless ,c ,@body))" + `

(defun main ()
  (print (my-if 1 "yes" "no"))
  (my-unless nil (print "a") (print "b"))
  (my-when-not nil (print "c")))
`)

	out, err := c.Compile()
	if err != nil {
		t.Fatalf("failed to compile %s", err)
	}
	if !strings.Contains(out, "call fn_main") {
		t.Fatalf("compilation looks bogus")
	}
}

// TestMacroErrors confirms a range of malformed macro definitions/calls
// are rejected with an error, rather than compiling to bogus code or
// panicking.
func TestMacroErrors(t *testing.T) {
	tests := []string{
		// wrong number of arguments.
		`(defmacro m (a b) a) (defun main () (m 1))`,
		`(defmacro m (a) a) (defun main () (m 1 2))`,
		`(defmacro m (a &b) a) (defun main () (m))`,

		// unbound symbol referenced in the macro body.
		`(defmacro m (a) b) (defun main () (m 1))`,

		// arbitrary computation isn't supported - only bound
		// parameters, literals, and quote/quasiquote templates.
		`(defmacro m (a) (+ a a)) (defun main () (m 1))`,

		// unquote-splicing used somewhere other than a list-element.
		`(defmacro m (a) ` + "`,@a)" + ` (defun main () (m 1))`,

		// infinite recursive expansion.
		`(defmacro m (a) ` + "`(m ,a))" + ` (defun main () (m 1))`,
	}

	for _, tst := range tests {
		c := New(tst)
		_, err := c.Compile()
		if err == nil {
			t.Fatalf("expected error, got none: %s", tst)
		}
	}
}
