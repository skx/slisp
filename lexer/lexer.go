// Package lexer contains our trivial lexer code.
//
// As lisp is so minimal this is as basic as could be, we only have to worry about
// comments, strings, and nothing else.
package lexer

import (
	"strings"
)

// Lexer holds our state.
type Lexer struct {
	// input contains the string we're going to lex into tokens.
	input string
}

// New is our constructor, which creates a new lexer from a given input source.
func New(src string) *Lexer {
	return &Lexer{
		input: src,
	}
}

// Tokenize is the method that returns an array of tokens from the given
// source.
func (l *Lexer) Tokenize() ([]string, error) {
	var out []string
	var cur strings.Builder

	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}

	inComment := false
	inString := false

	for _, ch := range l.input {

		// Note that we do no processing of "\n" to newline, etc.
		//
		// That is deferred to nasm.
		if inString {
			if ch == '"' {
				cur.WriteRune(ch)
				out = append(out, cur.String())
				cur.Reset()
				inString = false
				continue
			}
			cur.WriteRune(ch)
			continue
		}

		// comment start at ";" and end at the end of the line
		if inComment {
			if ch == '\n' {
				inComment = false
			}
			continue
		}

		// obvious stuff
		switch ch {
		case '(', ')':
			flush()
			out = append(out, string(ch))
		case '"':
			flush()
			cur.WriteRune(ch)
			inString = true

		case ' ', '\n', '\r', '\t':
			flush()

		case ';':
			flush()
			inComment = true
		default:
			cur.WriteRune(ch)
		}
	}

	flush()
	return out, nil
}
