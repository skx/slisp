package compiler

import (
	"strings"
	"testing"

	"github.com/skx/slisp/parser"
)

func TestBasic(t *testing.T) {

	c := New()

	p := parser.New(`
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
  (print (list 1 2 3 ))
  (print ( (lambda (x) 3) 3))
  (do
    (print 1)
    (print #\x)
    (print 2))
)
		`)

	defs, err := p.Parse()
	if err != nil {
		t.Fatalf("error parsing %s", err)
	}

	out, err := c.Compile(defs)
	if err != nil {
		t.Fatalf("failed to compile %s", err)
	}
	if !strings.Contains(out, "call fn_main") {
		t.Fatalf("compilation looks bogus")
	}

}
