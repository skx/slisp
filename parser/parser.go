// Package parser contains our AST definitions, and the code necessary
// to populate them from our input.
//
// Most of this package is very minimal stuff, as lisp is very low on
// syntax.
package parser

import (
	"fmt"
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
// all the top-level things we've found.
func (p *Parser) Parse() ([]TopLevel, error) {

	var defs []TopLevel

	// Tokenize the input
	var err error
	p.tokens, err = p.lex.Tokenize()
	if err != nil {
		return defs, err
	}

	// Now parse the input
	for p.pos < len(p.tokens) {
		defun, err := p.parseTopLevel()
		if err != nil {
			return defs, err
		}
		defs = append(defs, defun)
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

// expectNext confirms the next token is what is specified, if it isn't this
// will panic.
func (p *Parser) expectNext(s string) bool {
	return p.next() == s
}

// Parse parses the given input into a series of top-level expressions.
//
// top-level expressions are those that implement the "TopLevel" interface, and
// currently that means either "defconst", "defun" or "defvar".
func (p *Parser) parseTopLevel() (TopLevel, error) {

	// Everything starts with "("
	if !p.expectNext("(") {
		return nil, fmt.Errorf("expected '(' before opening top-level expression")
	}

	tok := p.next()

	switch tok {
	case "defconst":
		return p.parseGlobal(tok)
	case "defun":
		return p.parseDefun()
	case "defvar":
		return p.parseGlobal(tok)
	}

	return nil, fmt.Errorf("illegal top-level statement (%s ..)", tok)
}

// parseGlobal parses a global variable declaration, via either "defconst" or
// "defvar".
func (p *Parser) parseGlobal(tok string) (TopLevel, error) {
	name := p.next()

	// get the expression
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	if !p.expectNext(")") {
		return nil, fmt.Errorf("expected ')' after variable definition")
	}

	// defvar is not constant.
	// defconst .. is.
	constant := tok == "defconst"

	return Global{
		Const: constant,
		Name:  name,
		Value: expr,
	}, nil
}

// parseDefun parses a single function definition, containing an arbitrary number
// of expressions within the body.
func (p *Parser) parseDefun() (TopLevel, error) {

	// Get the name
	name := p.next()

	if !p.expectNext("(") {
		return nil, fmt.Errorf("expected '(' before defun arguments")
	}

	var params []string
	variadic := false
	for p.peek() != ")" && p.peek() != "" {
		params = append(params, p.next())
	}
	if !p.expectNext(")") {
		return nil, fmt.Errorf("expected ')' after defun arguments")
	}

	// Updated parameters with "&" removed.
	tmp := []string{}

	// If the last, and only the last, argument has a "&" prefix
	// then it should be removed and the function noted as having
	// variadic arguments.
	for i, param := range params {
		if after, ok := strings.CutPrefix(param, "&"); ok {
			if i == len(params)-1 {
				param = after
				variadic = true
			} else {
				return nil, fmt.Errorf("only the last parameter may have a &-prefix, saw it on %s: %s", name, param)
			}
		}
		tmp = append(tmp, param)
	}

	// body goes here
	body := []Expr{}

	// allow multiple expressions
	for p.peek() != "" && p.peek() != ")" {
		// get the expression
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

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
	if !p.expectNext(")") {
		return nil, fmt.Errorf("expected ')' after defun body")
	}

	return Defun{
		Name:     name,
		Params:   tmp,
		Exprs:    body,
		Variadic: variadic,
	}, nil
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
func (p *Parser) parseExpr() (Expr, error) {
	t := p.peek()

	if t == "(" {
		return p.parseList()
	}

	p.next()

	// char
	if after, ok := strings.CutPrefix(t, "#\\"); ok {
		x := after
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
		return &Char{Value: byte(c)}, nil
	}

	// string
	if strings.HasPrefix(t, "\"") {
		return &String{Value: strings.Trim(t, "\"")}, nil
	}

	// integer?
	if n, err := strconv.ParseInt(t, 0, 64); err == nil {
		return &Int{Value: n}, nil
	}

	// float?
	if f, err2 := strconv.ParseFloat(t, 64); err2 == nil {
		return &Float{Value: f}, nil
	}

	// nil?
	if t == "nil" {
		return &Nil{}, nil
	}

	// true?
	if t == "t" {
		return &Int{Value: 1}, nil
	}

	// symbol
	return &Symbol{Name: t}, nil
}

// parseList parses a list, handling any special forms, but otherwise
// converting "(foo bar baz)" into the AST node representing a call
// to function "foo" with bar/baz arguments.
func (p *Parser) parseList() (Expr, error) {

	if !p.expectNext("(") {
		return nil, fmt.Errorf("expected '(' for list opening")
	}

	// empty list?  That's a nil, baby
	if p.peek() == ")" {
		p.next()
		return &Nil{}, nil
	}

	head, err := p.parseExpr()
	if err != nil {
		return head, err
	}

	if sym, ok := head.(*Symbol); ok {
		switch sym.Name {

		case "cond":
			var cases []CondCase

			for p.peek() == "(" && p.peek() != "" {

				if !p.expectNext("(") {
					return nil, fmt.Errorf("expected '(' to open cond-case")
				}

				// condition
				cond, err := p.parseExpr()
				if err != nil {
					return nil, err
				}

				var exprs []Expr

				// arbitrary number of expressions
				for p.peek() != ")" && p.peek() != "" {
					x, err := p.parseExpr()
					if err != nil {
						return nil, err
					}
					exprs = append(exprs, x)
				}

				if !p.expectNext(")") {
					return nil, fmt.Errorf("expected ')' to close cond-case")
				}

				cases = append(cases, CondCase{
					Case:  cond,
					Exprs: exprs,
				})
			}

			if !p.expectNext(")") {
				return nil, fmt.Errorf("expected ')' to close cond")
			}

			return &Cond{
				Cases: cases,
			}, nil

		case "do":

			var exprs []Expr

			for p.peek() != ")" && p.peek() != "" {
				x, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				exprs = append(exprs, x)
			}

			if !p.expectNext(")") {
				return nil, fmt.Errorf("expected ')' after do-expressions")
			}

			return &Do{
				Exprs: exprs,
			}, nil

		case "if":
			cond, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			thenExpr, err2 := p.parseExpr()
			if err2 != nil {
				return nil, err2
			}
			var elseExpr Expr
			if p.peek() != ")" {
				var err error
				elseExpr, err = p.parseExpr()
				if err != nil {
					return nil, err
				}
			}
			if !p.expectNext(")") {
				return nil, fmt.Errorf("expected ')' to close IF")
			}

			return &If{
				Cond: cond,
				Then: thenExpr,
				Else: elseExpr,
			}, nil

		case "lambda":
			if !p.expectNext("(") {
				return nil, fmt.Errorf("expected '(' to open lambda parameters")
			}

			var params []string
			for p.peek() != ")" && p.peek() != "" {
				params = append(params, p.next())
			}
			if !p.expectNext(")") {
				return nil, fmt.Errorf("expected ')' to close lambda parameters")
			}

			// body goes here
			body := []Expr{}

			// allow multiple expressions
			for p.peek() != "" && p.peek() != ")" {
				expr, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				body = append(body, expr)

				// stop if we see a close
				if p.peek() == ")" {
					break
				}
			}

			// and ensure we do see that close
			if !p.expectNext(")") {
				return nil, fmt.Errorf("expected ')' to close lambda body")
			}

			// Lambda is basically a Defun with extra "Captures".
			// We don't populate those at parse time, so we'll just
			// populate the defun-things here.
			return &Lambda{
				Defun: Defun{
					Params: params,
					Exprs:  body,
				},
			}, nil

		case "let":
			if !p.expectNext("(") {
				return nil, fmt.Errorf("expected '(' to open let-bindings")
			}

			var binds []Binding

			for p.peek() == "(" && p.peek() != "" {

				if !p.expectNext("(") {
					return nil, fmt.Errorf("expected '(' to open let-binding")
				}

				name := p.next()
				expr, err := p.parseExpr()
				if err != nil {
					return nil, err
				}

				if !p.expectNext(")") {
					return nil, fmt.Errorf("expected ')' to close let-binding")
				}

				binds = append(binds, Binding{
					Name: name,
					Expr: expr,
				})
			}

			if !p.expectNext(")") {
				return nil, fmt.Errorf("expected ')' to close binding list")
			}

			// body goes here
			body := []Expr{}

			// allow multiple expressions
			for p.peek() != "" && p.peek() != ")" {
				expr, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				body = append(body, expr)

				// stop if we see a close
				if p.peek() == ")" {
					break
				}
			}
			// ensure we do see that close
			if !p.expectNext(")") {
				return nil, fmt.Errorf("expected ')' to close let")
			}

			return &Let{
				Bindings: binds,
				Body:     body,
			}, nil

		case "list":
			var args []Expr

			for p.peek() != ")" && p.peek() != "" {
				x, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, x)
			}

			if !p.expectNext(")") {
				return nil, fmt.Errorf("expected ')' to close list")
			}

			lst := p.buildList(args)
			return lst, nil

		case "set!":
			name := p.next()
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

			if !p.expectNext(")") {
				return nil, fmt.Errorf("expected ')' to close set! expression")
			}

			return &Set{
				Name: name,
				Expr: expr,
			}, nil

		case "while":
			cond, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

			var exprs []Expr

			for p.peek() != ")" && p.peek() != "" {
				x, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				exprs = append(exprs, x)
			}

			if !p.expectNext(")") {
				return nil, fmt.Errorf("expected ')' after while-expressions")
			}

			return &While{Cond: cond,
				Exprs: exprs,
			}, nil
		}
	}

	// Not a special form.
	//
	// Just handle it as a Call expression with any arguments
	var args []Expr

	for p.peek() != ")" && p.peek() != "" {
		x, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, x)
	}

	if !p.expectNext(")") {
		return nil, fmt.Errorf("expected ')' to close calling arguments")
	}

	return &Call{
		Fn:   head,
		Args: args,
	}, nil
}
