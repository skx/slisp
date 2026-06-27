package parser

import (
	"testing"
)

func TestParseValid(t *testing.T) {

	src := `
(defun main ()

  (if 1 (print "OK") (print "fail"))
  (if 1 (print "OK"))
  (let ((x 1))
   (set! x 2))
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
	}

	for _, txt := range tests {
		p := New(txt)
		_, err := p.Parse()
		if err == nil {
			t.Fatalf("expected error parsing %s - got none", txt)
		}
	}

}
