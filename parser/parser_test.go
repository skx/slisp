package parser

import (
	"testing"
)

func TestParseValid(t *testing.T) {

	src := `
(defconst foo 3)
(defvar steve 43)

(defun main ()
  (print #\\a #\\b #\\t #\\r #\\n )
  (if t (print "OK") (print "fail"))
  (if 1 (print "OK") (print "fail"))
  (if 1 (print "OK"))
  (let ((x 1))
   (cond
     ((int? x) (printint x))
     ((float? x) (printfloat x))
     (t        (print "steve")))

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
	out, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error parsing valid program; %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("expected three expressions")
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

		// cond
		`(cond x 3 4`,
		`(cond (((3  3 4`,
		`(cond )`,

		// defconst
		`(defconst 3 33 3)`,
		`(defconst 3 33 3`,
		`(defconst 3 `,
		`(defconst `,

		// defun
		`(defun (a ) `,
		`(defun (a b c (`,
		`(defun (a `,
		`(defun (`,
		`(defun `,
		`x (`,
		`(defun foo (a &b c) 1)`,
		`(defun foo (a b &c) `,

		// defvar
		`(defvar 3 33 3)`,
		`(defvar 3 33 3`,
		`(defvar 3 `,
		`(defvar `,

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

		// let
		`(let ((3 x x)))`,
		`(let ( 3  x x)))`,

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

		`(foo a 3 (`,
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

func TestQuoting(t *testing.T) {

	src := `
(defmacro my-if (c t e)
  ` + "`(cond (,c ,t) (t ,e)))" + `

(defun main ()
  (print 'a)
  (print '(1 2 3))
  (print ` + "`(1 ,(+ 1 1) ,@(list 3 4)))" + `
  (my-if 1 2 3))
`

	p := New(src)
	out, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error parsing valid program; %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected two top-level expressions, got %d", len(out))
	}

	macro, ok := out[0].(Defmacro)
	if !ok {
		t.Fatalf("expected first top-level item to be a Defmacro, got %T", out[0])
	}
	if macro.Name != "my-if" {
		t.Fatalf("unexpected macro name %q", macro.Name)
	}
	if len(macro.Params) != 3 {
		t.Fatalf("expected 3 macro parameters, got %d", len(macro.Params))
	}
	if _, ok := macro.Exprs[0].(*Quasiquote); !ok {
		t.Fatalf("expected macro body to be a Quasiquote, got %T", macro.Exprs[0])
	}

	def, ok := out[1].(Defun)
	if !ok {
		t.Fatalf("expected second top-level item to be a Defun")
	}

	call, ok := def.Exprs[0].(*Call)
	if !ok {
		t.Fatalf("expected first expression to be a call to print")
	}
	if _, ok := call.Args[0].(*Quote); !ok {
		t.Fatalf("expected first argument to be a Quote, got %T", call.Args[0])
	}
}

func TestBrokenQuoting(t *testing.T) {
	tests := []string{
		`(defmacro (a) `,
		`(defmacro foo (a `,
		`(defmacro foo (a) `,
		`(defmacro foo (a &b c) 1)`,
	}

	for _, txt := range tests {
		p := New(txt)
		_, err := p.Parse()
		if err == nil {
			t.Fatalf("expected error parsing %s - got none", txt)
		}
	}
}

func TestFloat(t *testing.T) {
	p := New("(defun main() (print 3.1))")
	out, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error parsing valid program; %v", err)
	}

	d, ok := out[0].(Defun)
	if !ok {
		t.Fatalf("top level result wasn't a defun")
	}

	for _, x := range d.Exprs {

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
