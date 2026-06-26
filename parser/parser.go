// Package parser contains our AST definitions, which are turned into
// code by our compiler.
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
		defun, err := p.parseDefun()
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
	if p.next() != s {
		return false
	}
	return true
}

// parseDefun parses a single function definition, containing an arbitrary number
// of expressions within the body.
func (p *Parser) parseDefun() (*Defun, error) {

	if !p.expectNext("(") {
		return nil, fmt.Errorf("expected '(' before opening defun")
	}

	tok := p.next()
	if tok != "defun" {
		return nil, fmt.Errorf("expected defun, but got %v", tok)
	}

	name := p.next()

	if !p.expectNext("(") {
		return nil, fmt.Errorf("expected '(' before defun arguments")
	}

	var params []string
	for p.peek() != ")" && p.peek() != "" {
		params = append(params, p.next())
	}
	if !p.expectNext(")") {
		return nil, fmt.Errorf("expected ')' after defun arguments")
	}

	// body goes here
	body := []Expr{}

	// allow multiple expressions
	for p.peek() != "" {
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

	return &Defun{
		Name:   name,
		Params: params,
		Exprs:  body,
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
		return &Char{Value: byte(c)}, nil
	}

	// string
	if strings.HasPrefix(t, "\"") {
		return &String{Value: strings.Trim(t, "\"")}, nil
	}

	// integer
	if n, err := strconv.ParseInt(t, 0, 64); err == nil {
		return &Int{Value: n}, nil
	}

	// nil?
	if t == "nil" {
		return &Nil{}, nil
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

	head, err := p.parseExpr()
	if err != nil {
		return head, err
	}

	if sym, ok := head.(*Symbol); ok {
		switch sym.Name {

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
			for p.peek() != "" {
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
			for p.peek() != "" {
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
				return nil, fmt.Errorf("expected ')' to close set!")
			}

			return &Set{
				Name: name,
				Expr: expr,
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
