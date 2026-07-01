package parser

import (
	"testing"
)

func TestParseValid(t *testing.T) {

	src := `
(defun main ()
  (if t (print "OK") (print "fail"))
  (if 1 (print "OK") (print "fail"))
  (if 1 (print "OK"))
  (let ((x 1))
   (set! x 2)
   (while (> x 0)
     (println x)
     (set! x (- x 1))))

  (print (list 1 2 3 ))
  (print ( (lambda (x) 3) 3))
  (do
    (print 1)
    (print 2))
)
`

	p := New(src)
	_, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error parsing valid program; %v", err)
	}
}

func TestIssue68(t *testing.T) {

	src := `
(defun empty())

(defun main ()
  (let ((binding nil)))
  (do)
  (print (lambda ()))
  (list)
)
`

	p := New(src)
	_, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error parsing valid program; %v", err)
	}
}

func TestBroken(t *testing.T) {

	tests := []string{

		// defun
		`(defun (a ) `,
		`(defun (a b c`,
		`(defun (a `,
		`(defun (`,
		`(defun `,
		`(`,

		// if
		`(if 1 (print "OK") (print "fail")`,
		`(if 1 (print "OK") (print "fail"`,
		`(if 1 (print "OK") `,
		`(if 1 (print "OK")`,
		`(if 1 (print "OK"`,
		`(if 1 (print `,
		`(if 1 (`,
		`(if 1`,
		`(if`,

		// do
		`(len (do (print 1) ((`,
		`(len (do (print 1 ((`,
		`(len (do (print ((`,
		`(len (do ((`,
		`(len (do((`,

		// while
		`(while (< 1 1) (do (print ok))`,
		`(while (< 1 1) (do (print ok)`,
		`(while (< 1 1) (do (print`,
		`(while (< 1 1) (do (`,
		`(while (< 1 1) (`,
		`(while (< 1 1`,
		`(while (<`,
		`(while (`,

		`(foo `,
	}

	for _, txt := range tests {
		p := New(txt)
		_, err := p.Parse()
		if err == nil {
			t.Fatalf("expected error parsing %s - got none", txt)
		}
	}

}

func TestEmptyList(t *testing.T) {
	p := New("(defun main() (print ()))")
	_, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error parsing valid program; %v", err)
	}
}

func TestFloat(t *testing.T) {
	p := New("(defun main() (print 3.1))")
	out, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error parsing valid program; %v", err)
	}

	for _, x := range out[0].Exprs {

		call, ok := x.(*Call)
		if !ok {
			t.Fatalf("expression isn't a call")
		}

		arg := call.Args[0]
		f, ok := arg.(*Float)
		if !ok {
			t.Fatalf("argument isn't a float: %v", arg)
		}
		if f.Value != 3.1 {
			t.Fatalf("wrong floating point value %f", f)
		}
	}

}
