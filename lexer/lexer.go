// Package lexer contains our trivial lexer code.
//
// As lisp is so minimal this is as basic as could be, we only have to worry about
// comments, strings, and nothing else.
//
// We build up characters into an internal reader and just stop when
// we see "(", ")", "newline", etc.  Character literals, comments
// and strings are peeled off ahead of that.
//
// TODO: Better.
package lexer

import (
	"fmt"
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
			val := cur.String()
			if strings.HasPrefix(val, ":") {
				val = val[1:]
				val = "\"" + val + "\""
			}
			out = append(out, val)
			cur.Reset()
		}
	}

	inComment := false
	inString := false

	for i := 0; i < len(l.input); i++ {
		ch := l.input[i]

		// Note that we do no processing of "\n" to newline, etc.
		//
		// That is deferred to nasm.
		if inString {
			if ch == '"' {
				cur.WriteByte(ch)
				out = append(out, cur.String())
				cur.Reset()
				inString = false
				continue
			}
			cur.WriteByte(ch)
			continue
		}

		// comment start at ";" and end at the end of the line
		if inComment {
			if ch == '\n' {
				inComment = false
			}
			continue
		}

		// A bit ugly, but we test #\xx specially here
		// because otherwise "#\(" wouldn't be possible
		// as we regard "(" and ")" as separators.
		if ch == '#' &&
			i+1 < len(l.input) &&
			l.input[i+1] == '\\' {

			flush()

			cur.WriteByte('#')
			cur.WriteByte('\\')
			i += 2

			if i >= len(l.input) {
				return nil, fmt.Errorf("unterminated character literal")
			}

			cur.WriteByte(l.input[i])

			if l.input[i] == '\\' {
				i++
				if i >= len(l.input) {
					return nil, fmt.Errorf("unterminated escape")
				}
				cur.WriteByte(l.input[i])
			}

			out = append(out, cur.String())
			cur.Reset()
			continue
		}

		// obvious stuff
		switch ch {

		case '(', ')':
			flush()
			out = append(out, string(ch))

		case '"':
			flush()
			cur.WriteByte(ch)
			inString = true

		case ' ', '\n', '\r', '\t':
			flush()

		case ';':
			flush()
			inComment = true

		default:
			cur.WriteByte(ch)
		}
	}

	flush()
	return out, nil
}
