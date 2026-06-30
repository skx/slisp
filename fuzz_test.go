package main

import (
	"github.com/skx/slisp/parser"
	"strings"
	"testing"
)

// FuzzProject runs the fuzz-testing against our parser and compiler.
//
// We mostly catch errors with the lexer and parser here, the compiler itself
// will generate and return text for the AST we produce.
func FuzzProject(f *testing.F) {

	// Known errors we might see
	known := []string{
		// #\x
		"unterminated character literal",
		"unterminated escape",

		//
		"expected '('",
		"expected ')'",

		// defun
		"expected defun",

		// variables
		"unknown variable",
	}

	//
	// Some examples to seed the fuzz corpus
	//
	testcases := []string{
		// comments
		`;; comment
(defun main() (print 3) ;; comment
)`,

		// simple maths
		"(defun main() (print (= 0 0)))",
		"(defun main() (print (! 0)))",
		"(defun main() (print (<= 0 0)))",
		"(defun main() (print (< 0 0)))",
		"(defun main() (print (> 0 0)))",
		"(defun main() (print (>= 0 0)))",
		"(defun main() (print (+ 0 0)))",
		"(defun main() (print (- 0 0)))",
		"(defun main() (print (* 0 0)))",
		"(defun main() (print (/ 0 0)))",
		"(defun main() (print (% 0 0)))",

		// type-checking
		"(defun main() (print (cons? 0)))",
		"(defun main() (print (char? 0)))",
		"(defun main() (print (int? 0)))",
		"(defun main() (print (lambda? 0)))",
		"(defun main() (print (nil? 0)))",
		"(defun main() (print (str?  0)))",

		// strings
		`(defun main() (print "Hello, world!"))`,

		// if
		`(defun main() (if nil (print "nil") (print "true")))`,
		`(defun main() (if nil (print "nil")))`,

		// lambda
		`(defun makeAdder (n) (lambda (x)(+ x n)))
		 (defun main () (let ((ten (makeAdder 10)))  (print (ten 90))))`,

		// let
		`(defun main() (let ((x 1)) (print x)))`,

		// set!
		`(defun main() (let ((x 1)) (set! x 3) (print x)))`,

		// list
		`(defun main() (print (list 1 2 3 #\x "steve")))`,

		// do
		`(defun main() (do (print (list 1 2 3 #\x "steve"))))`,

		// cond
		`(defun main() (cond ((1 (print "one")))))`,
		`(defun main() (let ((n 2))
		   (cond (
		     (= n 1) (print "one")
		     (= n 2) (print "two")
		     ))))`,

		// while
		`(defun main() (while nil (print "ok")))`,
		`(defun main() (let ((i 0))
				 (while (< i 10)
				   (println i)
				   (set! i (+ i 1)))))`,
	}

	//
	// Seed the fuzzer with our samples
	//
	for _, tc := range testcases {
		f.Add([]byte(tc))
	}

	//
	// Run the fuzzer.
	//
	f.Fuzz(func(t *testing.T, input []byte) {
		falsePositive := false

		// Create a parser
		p := parser.New(string(input))

		// Parse into functions
		_, err := p.Parse()
		if err != nil {

			//
			// We got an error, was it a false-positive?
			//
			for _, ignored := range known {
				if strings.Contains(err.Error(), ignored) {
					falsePositive = true
				}
			}

			//
			// If it wasn't a false positive we want to see what
			// was produced and mark it as a failure.
			//
			if !falsePositive {
				t.Fatalf("error running input: '%s': %v", input, err)
			}
		}
	})
}
