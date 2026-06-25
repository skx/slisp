package lexer

import (
	"testing"
)

func TestWhitespace(t *testing.T) {

	l := New("         \t\n\r")
	out, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error tokenizing")
	}
	if len(out) != 0 {
		t.Fatalf("unexpected lexer result %d - %v", len(out), out)
	}
}

func TestComment(t *testing.T) {
	l := New(`
;; Test comment
"Test" ;; comment at the end`)
	out, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error tokenizing")
	}
	if len(out) != 1 {
		t.Fatalf("unexpected lexer result %d - %v", len(out), out)
	}
}

func TestBasic(t *testing.T) {
	l := New(`
(if nil
  (message "fatal error")
 (message "This is fine."))`)

	out, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error tokenizing")
	}
	if len(out) != 12 {
		t.Fatalf("unexpected lexer result %d - %v", len(out), out)
	}

}
