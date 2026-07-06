// Package lexer contains code to convert our input program into a series of tokens.
//
// As lisp is so minimal this package is as basic as it could be; we only really have
// to care about strings, character literals and comments.  Everything else is either
// a symbol or parenthesis characters.
package lexer

import (
	"fmt"
	"strings"
)

// Lexer holds our state.
type Lexer struct {

	// input contains the string we're going to lex into tokens.
	input string

	// pos records our current position.
	pos int
}

// New creates a new lexer object
func New(src string) *Lexer {
	return &Lexer{
		input: src,
	}
}

// Tokenize returns the tokenized version of the input string passed to our constructor.
func (l *Lexer) Tokenize() ([]string, error) {
	var out []string

	for !l.eof() {

		switch l.peek() {

		case ' ', '\t', '\n', '\r':
			l.skipWhitespace()

		case ';':
			l.skipComment()

		case '(':
			l.next()
			out = append(out, "(")

		case ')':
			l.next()
			out = append(out, ")")

		case '"':
			str, err := l.scanStringLiteral()
			if err != nil {
				return nil, err
			}
			out = append(out, str)

		case '#':
			tok, err := l.scanCharacterLiteral()
			if err != nil {
				return nil, err
			}
			out = append(out, tok)

		default:
			out = append(out, l.scanSymbol())
		}
	}

	return out, nil
}

// scanStringLiteral is called to consume a string literal, and return it.
func (l *Lexer) scanStringLiteral() (string, error) {
	var b strings.Builder

	// We've already seen '"'
	if l.next() != '"' {
		return "", fmt.Errorf("scanStringLiteral called at wrong position")
	}

	// opening quote
	b.WriteByte('"')

	for !l.eof() {

		ch := l.next()

		switch ch {

		case '\\':
			b.WriteByte(ch)

			if l.eof() {
				return "", fmt.Errorf("unterminated escape sequence in string")
			}

			b.WriteByte(l.next())

		case '"':
			b.WriteByte(ch)
			return b.String(), nil

		default:
			b.WriteByte(ch)
		}
	}

	return "", fmt.Errorf("unterminated string")
}

// scanCharacterLiteral is invoked to parse a character literal, i.e. something prefixed with #\.
func (l *Lexer) scanCharacterLiteral() (string, error) {

	// We've already seen '#'
	if l.next() != '#' {
		return "", fmt.Errorf("scanCharacterLiteral called at wrong position")
	}

	if l.eof() || l.next() != '\\' {
		return "", fmt.Errorf("expected '\\' after '#'")
	}

	var b strings.Builder
	b.WriteString("#\\")

	if l.eof() {
		return "", fmt.Errorf("unterminated character literal")
	}

	ch := l.next()
	b.WriteByte(ch)

	// Allow escaped characters such as #\\ or #\"
	if ch == '\\' {
		if l.eof() {
			return "", fmt.Errorf("unterminated escape")
		}
		b.WriteByte(l.next())
	}

	return b.String(), nil
}

// scanSymbol is invoked to parse a symbol.
func (l *Lexer) scanSymbol() string {
	var b strings.Builder

	for !l.eof() {

		switch l.peek() {
		case ' ', '\t', '\n', '\r', ';', '(', ')':
			goto done

		default:
			b.WriteByte(l.next())
		}
	}

done:

	s := b.String()

	// :-prefixed symbols become strings.
	if strings.HasPrefix(s, ":") {
		return `"` + s[1:] + `"`
	}

	return s
}

// skipWhitespace skips over any whitespace.
func (l *Lexer) skipWhitespace() {
	for !l.eof() {
		switch l.peek() {
		case ' ', '\t', '\n', '\r':
			l.next()
		default:
			return
		}
	}
}

// skipComment moves our lexer over any comment.
//
// Comments start with ";" and continue until the end of the line.
func (l *Lexer) skipComment() {
	for !l.eof() && l.peek() != '\n' {
		l.next()
	}
}

// eof tests if we've exceeded the length of our input
func (l *Lexer) eof() bool {
	return l.pos >= len(l.input)
}

// peek returns the next character which will be encountered, without consuming it or moving
// our position forward.
func (l *Lexer) peek() byte {
	if l.eof() {
		return 0
	}
	return l.input[l.pos]
}

// next advances our position and returns the next character to be processed.
func (l *Lexer) next() byte {
	ch := l.peek()
	l.pos++
	return ch
}
