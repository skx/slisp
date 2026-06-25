// Package parser contains our AST definitions, which are turned into
// code by our compiler.
//
// Most of this package is very minimal stuff, as lisp is very low on
// syntax.
package parser

import (
	"strconv"
	"strings"

	"github.com/skx/slisp/lexer"
)

// Parser holds our parse-state
type Parser struct {
	tokens []string
	lex    *lexer.Lexer
	pos    int
}

// New is our constructor.
func New(src string) *Parser {
	p := &Parser{
		lex: lexer.New(src),
	}
	return p
}

// Parse processes the input which was given in our constructor, and returns
// all the top-level defuns found.
//
// We don't allow top-level expressions in our language.
func (p *Parser) Parse() ([]*Defun, error) {

	var defs []*Defun

	// Tokenize the input
	var err error
	p.tokens, err = p.lex.Tokenize()
	if err != nil {
		return defs, err
	}

	// Now parse the input
	for p.pos < len(p.tokens) {
		defs = append(defs, p.parseDefun())
	}

	return defs, nil
}

// peek returns the next token, without consuming it.
func (p *Parser) peek() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	return p.tokens[p.pos]
}

// next returns the next token.
func (p *Parser) next() string {
	t := p.peek()
	p.pos++
	return t
}

// expect confirms the next token is what is specified, if it isn't this
// will panic.
func (p *Parser) expect(s string) {
	if p.next() != s {
		panic("expected " + s)
	}
}

// parseDefun parses a single function definition, containing an arbitrary number
// of expressions within the body.
func (p *Parser) parseDefun() *Defun {
	p.expect("(")
	if p.next() != "defun" {
		panic("expected defun")
	}

	name := p.next()

	p.expect("(")

	var params []string
	for p.peek() != ")" {
		params = append(params, p.next())
	}
	p.expect(")")

	// body goes here
	body := []Expr{}

	// allow multiple expressions
	for p.peek() != "" {
		// get the expression
		expr := p.parseExpr()

		// If there are no expressions
		if len(body) == 0 {
			// And the first expression is a string
			// we just ignore it and continue, around this
			// loop again.
			switch expr.(type) {
			case *String:
				continue
			}
		}
		body = append(body, expr)

		// stop if we see a close
		if p.peek() == ")" {
			break
		}
	}

	// and ensure we do see that close
	p.expect(")")

	return &Defun{
		Name:   name,
		Params: params,
		Exprs:  body,
	}
}

// buildList is used to turn "(list 1 2 3)" into "(cons 1 (cons 2 (cons 3 nil)))"
func (p *Parser) buildList(args []Expr) Expr {
	result := Expr(&Nil{})

	for i := len(args) - 1; i >= 0; i-- {
		result = &Call{
			Fn: &Symbol{Name: "cons"},
			Args: []Expr{
				args[i],
				result,
			},
		}
	}

	return result
}

// parseExpr parses a single expression, and returns the appropriate AST node.
func (p *Parser) parseExpr() Expr {
	t := p.peek()

	if t == "(" {
		return p.parseList()
	}

	p.next()

	// char
	if strings.HasPrefix(t, "#\\") {
		x := strings.TrimPrefix(t, "#\\")
		c := x[0]
		if c == '\\' && len(x) > 1 {
			switch x[1] {
			case 'a':
				c = '\a'
			case 'b':
				c = '\b'
			case 'r':
				c = '\r'
			case 't':
				c = '\t'
			case 'n':
				c = '\n'

			}
		}
		return &Char{Value: byte(c)}
	}

	// string
	if strings.HasPrefix(t, "\"") {
		return &String{Value: strings.Trim(t, "\"")}
	}

	// integer
	if n, err := strconv.ParseInt(t, 0, 64); err == nil {
		return &Int{Value: n}
	}

	// nil?
	if t == "nil" {
		return &Nil{}
	}

	// symbol
	return &Symbol{Name: t}
}

// parseList parses a list, handling any special forms, but otherwise
// converting "(foo bar baz)" into the AST node representing a call
// to function "foo" with bar/baz arguments.
func (p *Parser) parseList() Expr {
	p.expect("(")

	head := p.parseExpr()

	if sym, ok := head.(*Symbol); ok {
		switch sym.Name {

		case "do":

			var exprs []Expr

			for p.peek() != ")" && p.peek() != "" {
				exprs = append(exprs, p.parseExpr())
			}

			p.expect(")")

			return &Do{
				Exprs: exprs,
			}

		case "if":
			cond := p.parseExpr()
			thenExpr := p.parseExpr()
			var elseExpr Expr
			if p.peek() != ")" {
				elseExpr = p.parseExpr()
			}
			p.expect(")")

			return &If{
				Cond: cond,
				Then: thenExpr,
				Else: elseExpr,
			}

		case "lambda":
			p.expect("(")

			var params []string
			for p.peek() != ")" && p.peek() != "" {
				params = append(params, p.next())
			}
			p.expect(")")

			// body goes here
			body := []Expr{}

			// allow multiple expressions
			for p.peek() != "" {
				expr := p.parseExpr()
				body = append(body, expr)

				// stop if we see a close
				if p.peek() == ")" {
					break
				}
			}

			// and ensure we do see that close
			p.expect(")")

			return &Lambda{
				Params: params,
				Exprs:  body,
			}

		case "let":
			p.expect("(")

			var binds []Binding

			for p.peek() == "(" && p.peek() != "" {
				p.expect("(")
				name := p.next()
				expr := p.parseExpr()
				p.expect(")")

				binds = append(binds, Binding{
					Name: name,
					Expr: expr,
				})
			}

			p.expect(")")

			// body goes here
			body := []Expr{}

			// allow multiple expressions
			for p.peek() != "" {
				expr := p.parseExpr()
				body = append(body, expr)

				// stop if we see a close
				if p.peek() == ")" {
					break
				}
			}
			// ensure we do see that close
			p.expect(")")

			return &Let{
				Bindings: binds,
				Body:     body,
			}

		case "list":
			var args []Expr

			for p.peek() != ")" && p.peek() != "" {
				args = append(args, p.parseExpr())
			}

			p.expect(")")

			return p.buildList(args)

		case "set!":
			name := p.next()
			expr := p.parseExpr()

			p.expect(")")

			return &Set{
				Name: name,
				Expr: expr,
			}
		}
	}

	// Not a special form.
	//
	// Just handle it as a Call expression with any arguments
	var args []Expr

	for p.peek() != ")" && p.peek() != "" {
		args = append(args, p.parseExpr())
	}

	p.expect(")")

	return &Call{
		Fn:   head,
		Args: args,
	}

}
