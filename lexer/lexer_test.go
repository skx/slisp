package lexer

import (
	"testing"
)

// TestString makes various tests against strings
func TestString(t *testing.T) {

	// happy path
	l := New(`"Hello, world!"`)

	str, err := l.scanStringLiteral()
	if err != nil {
		t.Fatalf("expected no error, got one %s", err)
	}
	if str != "\"Hello, world!\"" {
		t.Fatalf("got wrong result: %s", str)
	}

	// happy path with quotes
	l = New(`" Hello \"Steve\" "`)

	str, err = l.scanStringLiteral()
	if err != nil {
		t.Fatalf("expected no error, got one %s", err)
	}
	if str != `" Hello \"Steve\" "` {
		t.Fatalf("got wrong result: %s", str)
	}

	// unterminated string
	l = New(`"Hello, world!`)

	_, err = l.scanStringLiteral()
	if err == nil {
		t.Fatalf("expected [unterminated string] error, got none")
	}

	// unterminated escape
	l = New(`"Hello, world!\`)

	_, err = l.scanStringLiteral()
	if err == nil {
		t.Fatalf("expected [unterminated escape] error, got none")
	}

	// not a string (can't happen)
	l = New(`meh`)
	_, err = l.scanStringLiteral()
	if err == nil {
		t.Fatalf("expected [not a string] error, got none")
	}
}

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
	// valid program
	l := New(`
(println 3.1)
(println "This is " pi )
(println "I fake symbols " :symbol)

(if nil
  (message "fatal error")
 (message "This is fine."))`)

	out, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error tokenizing")
	}
	if len(out) != 26 {
		t.Fatalf("unexpected lexer result %d - %v", len(out), out)
	}

	// broken program - string error
	l = New(`
(println "This is never complete`)

	_, err = l.Tokenize()
	if err == nil {
		t.Fatalf("expected error, saw none")
	}
}

func TestCharacterLiteral(t *testing.T) {

	// broken
	test := []string{
		`#\`,
		`#\\`,
		`#x`,
		`#\cake`,
	}

	for _, tst := range test {
		l := New(tst)

		_, err := l.Tokenize()
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	}

	// valid
	test = []string{
		`#\(`,
		`#\)`,
		`#\a`,
		`#\B`,
		`#\\n`,
		`#\Newline`,
		`#\Space`,
		`#\Tab`,
		`#\Return`,
	}

	for _, tst := range test {
		l := New(tst)

		out, err := l.Tokenize()
		if err != nil {
			t.Fatalf("unexpected error parsing %s : %s", tst, err)
		}
		if len(out) != 1 {
			t.Fatalf("unexpected lexer result %d - %v", len(out), out)
		}
	}

	// not a character literal (can't happen)
	l := New(`meh`)
	_, err := l.scanCharacterLiteral()
	if err == nil {
		t.Fatalf("expected [not a character] error, got none")
	}

}

func TestEOF(t *testing.T) {
	l := New(``)
	i := 0
	for i < 10 {
		l.peek()
		l.next()
		i++
	}

}
