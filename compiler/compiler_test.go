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

(defun main ()
  (let ((f (counter)))
    (println (f))
    (println (f))
  )
  (if 1 (print "OK") (print "fail"))
  (if 1 (print "OK"))
  (let ((x 1))
   (set! x 2))
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
