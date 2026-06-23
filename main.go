// trivial lisp compiler which generates nasm-style assembly.
//
// Has support for integers, strings, printing and simple maths.
//
// Integer operations:
// [X] -
// [X] +
// [X] *
// [X] /
// [X] <=
// [X] <
// [X] >
// [X] >=
//
// Special forms
// [X] DEFUN
// [X] IF
// [X] LET
// [X] INT?    // true if value is integer.
// [X] STR?    // true if value is string.
// [X] CONS?   // true if value is cons
// [X] LAMBDA? // true if value is lambda
// [X] NIL?    // true if value is nil
//
// Standard library:
// [X] PRINTINT
// [X] PRINTSTR
// [X] EXIT
// [X] NEWLINE
// [X] PUTC
//
// The lower three bits of values is used for type storage, with macros used for getting/setting
// them, to avoid user-error.  Hopefully:
//
//	000:  INT
//	001:  STRING
//      010:  CONS
//      011:  LAMBDA
//      100:  ...
//      101:  ...
//      110:  ...
//      111:  NIL

// ABI uses the Sys V style, so max six arguments:
// arg0 -> rdi
// arg1 -> rsi
// arg2 -> rdx
// arg3 -> rcx
// arg4 -> r8
// arg5 -> r9
package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
)

//go:embed stdlib.lisp
var stdlibLisp string

//go:embed template.tmpl
var tmplTxt string

//
// AST
//

type Expr interface{}

// Types
type Int struct {
	Value int64
}

type String struct {
	Value string
}

type Symbol struct {
	Name string
}

type Nil struct {
}

// Used for inline strings
type StringLiteral struct {
	Label string
	Value string
}

// specials

type Binding struct {
	Name string
	Expr Expr
}

type Call struct {
	Fn   Expr
	Args []Expr
}

type Defun struct {
	Name   string
	Params []string
	Exprs  []Expr
}

type Do struct {
	Exprs []Expr
}

type If struct {
	Cond Expr
	Then Expr
	Else Expr
}

type Lambda struct {
	// name is auto-generated when we encounter the lambda
	name   string
	Params []string
	Exprs  []Expr
}

type Let struct {
	Bindings []Binding
	Body     []Expr
}

type Set struct {
	Name string
	Expr Expr
}

//
// Lexer
//

// tokenize is a trivial tokenizer which can handle strings, comments, and
// basic splitting.
func tokenize(src string) []string {
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

	for _, ch := range src {

		// naive - no processing of "\n" to newline, etc.
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
	return out
}

//
// Parser
//

// Parser holds our parse-state
type Parser struct {
	tokens []string
	pos    int
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

// parseProgram uses the Parser and returns a series of functions "(defun .." from it.
//
// We don't allow top-level expressions in our language.
func parseProgram(src string) []*Defun {
	p := &Parser{
		tokens: tokenize(src),
	}

	var defs []*Defun

	for p.pos < len(p.tokens) {
		defs = append(defs, p.parseDefun())
	}

	return defs
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
	for {
		expr := p.parseExpr()
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

	// string
	if strings.HasPrefix(t, "\"") {
		return &String{Value: strings.Trim(t, "\"")}
	}

	// integer
	if n, err := strconv.ParseInt(t, 10, 64); err == nil {
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

			for p.peek() != ")" {
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
			for p.peek() != ")" {
				params = append(params, p.next())
			}
			p.expect(")")

			// body goes here
			body := []Expr{}

			// allow multiple expressions
			for {
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

			for p.peek() == "(" {
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
			for {
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

			for p.peek() != ")" {
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

	for p.peek() != ")" {
		args = append(args, p.parseExpr())
	}

	p.expect(")")

	return &Call{
		Fn:   head,
		Args: args,
	}

}

//
// Environment
//

type Env struct {
	parent *Env
	slots  map[string]int
}

// NewEnv creates a new environment, with an optional parent.
func NewEnv(parent *Env) *Env {
	return &Env{
		parent: parent,
		slots:  map[string]int{},
	}
}

// Lookup returns the slot-index of the given variable-name.
//
// If not found in the current scope the parent(s) will be searched, recursively.
func (e *Env) Lookup(name string) (int, bool) {
	if v, ok := e.slots[name]; ok {
		return v, true
	}

	if e.parent != nil {
		return e.parent.Lookup(name)
	}

	return 0, false
}

//
// Code Generator
//

type Generator struct {
	// text stores the text we emit as we compile
	text strings.Builder

	// labelID is used to give unique labels to if/lambda/etc
	labelID int

	// strings holds the strings we've encountered, we need to
	// emit those with their labels later.
	strings []StringLiteral

	// lambdas holds the lambdas we've encountered and we need
	// to emit those later too.
	lambdas []*Lambda
}

func (g *Generator) label(prefix string) string {
	s := fmt.Sprintf("%s_%d", prefix, g.labelID)
	g.labelID++
	return s
}

func (g *Generator) emitln(s string) {
	g.text.WriteString(s)
	g.text.WriteString("\n")
}

// asmName converts the given label into something nasm will
// accept.  It doesn't like special characters inside label names.
func asmName(name string) string {
	switch name {

	// comparisons
	case "=", "==":
		return "equals"
	case "!":
		return "not"
	case "<=":
		return "lt_equals"
	case "<":
		return "lt"
	case ">":
		return "gt"
	case ">=":
		return "gt_equals"

	// maths
	case "+":
		return "integer_plus"
	case "-":
		return "integer_minus"
	case "*":
		return "integer_multiply"
	case "/":
		return "integer_divide"

	// type checks
	case "cons?":
		return "consp"
	case "int?":
		return "intp"
	case "lambda?":
		return "lambdap"
	case "nil?":
		return "nilp"
	case "set!":
		return "set"
	case "str?":
		return "strp"
	}

	return name
}

// emitExpr emits the code for each of our expression AST types.
func (g *Generator) emitExpr(e Expr, env *Env) {
	switch n := e.(type) {

	case *Call:
		if symbol, ok := n.Fn.(*Symbol); ok {

			regs := []string{
				"rdi",
				"rsi",
				"rdx",
				"rcx",
				"r8",
				"r9",
			}

			for _, a := range n.Args {
				g.emitExpr(a, env)
				g.emitln("    push rax")
			}

			for i := len(n.Args) - 1; i >= 0; i-- {
				g.emitln(fmt.Sprintf(
					"    pop %s",
					regs[i],
				))
			}

			// lambda?
			if offset, ok := env.Lookup(symbol.Name); ok {

				g.emitln(fmt.Sprintf(
					"    mov rax,[rbp-%d]",
					offset,
				))

				g.emitln("    UNTAG_REG rax")
				g.emitln("    call rax")

				return
			} else {
				// defun
				g.emitln("    call " + asmName(symbol.Name))
				return
			}
		}

		regs := []string{
			"rdi",
			"rsi",
			"rdx",
			"rcx",
			"r8",
			"r9",
		}

		for _, a := range n.Args {
			g.emitExpr(a, env)
			g.emitln("    push rax")
		}

		for i := len(n.Args) - 1; i >= 0; i-- {
			g.emitln(fmt.Sprintf(
				"    pop %s",
				regs[i],
			))
		}

		// evaluate callable expression
		g.emitExpr(n.Fn, env)

		g.emitln("    UNTAG_REG rax")
		g.emitln("    call rax")

	case *Do:
		for _, expr := range n.Exprs {
			g.emitExpr(expr, env)
		}

	case *Int:
		g.emitln(fmt.Sprintf("    mov rax, %d", n.Value))
		g.emitln("   TAG_INTEGER_REG rax")

	case *If:
		elseLbl := g.label("else")
		endLbl := g.label("endif")

		g.emitExpr(n.Cond, env)

		g.emitln("    and rax, 7    ; get type bits")
		g.emitln("    cmp rax, 7    ; is this a nil?")
		g.emitln("    jz " + elseLbl)

		g.emitExpr(n.Then, env)

		g.emitln("    jmp " + endLbl)

		g.emitln(elseLbl + ":")

		// else branch is optional
		if n.Else != nil {
			g.emitExpr(n.Else, env)
		}
		g.emitln(endLbl + ":")

	case *Lambda:
		// create a unique name for this lambda
		name := fmt.Sprintf("lambda_%d", g.labelID)
		g.labelID++

		// load the address - it will be compiled eventually.
		g.emitln(fmt.Sprintf("    lea rax, %s", name))
		g.emitln("    TAG_LAMBDA_REG rax")

		// save away the lambda in the list of lambdas we
		// know about, because we do need to compile it .. later
		n.name = name
		g.lambdas = append(g.lambdas, n)

	case *Let:
		// create a new child environment
		child := NewEnv(env)

		nextSlot := len(child.slots)

		// populate the new environment
		for _, b := range n.Bindings {

			g.emitExpr(b.Expr, env)

			offset := (nextSlot + 1) * 8

			child.slots[b.Name] = offset

			g.emitln(fmt.Sprintf(
				"    mov [rbp-%d], rax",
				offset,
			))

			nextSlot++
		}

		// compile each expression within the body
		for _, expr := range n.Body {
			g.emitExpr(expr, child)
		}

	case *Nil:
		g.emitln("    mov rax, 0       ; NIL")
		g.emitln("    TAG_NIL_REG rax  ; Tagged")

	case *String:
		// create a label
		lbl := g.label("str")

		// save the string, because we're gonna put it into the
		// generated code, later.
		g.strings = append(
			g.strings,
			StringLiteral{
				Label: lbl,
				Value: n.Value,
			},
		)

		// load the address of the label and tag.
		g.emitln(fmt.Sprintf("    lea rax, %s", lbl))
		g.emitln("    TAG_STRING_REG rax")

	case *Set:
		offset, ok := env.Lookup(n.Name)
		if !ok {
			panic("unknown variable: " + n.Name)
		}

		g.emitExpr(n.Expr, env)

		g.emitln(fmt.Sprintf(
			"    mov [rbp-%d], rax",
			offset,
		))

	case *Symbol:
		offset, ok := env.Lookup(n.Name)
		if !ok {
			panic("unknown symbol: " + n.Name)
		}

		g.emitln(fmt.Sprintf(
			"    mov rax, [rbp-%d]",
			offset,
		))

	default:
		panic(fmt.Sprintf("emitExpr: Unhandled node type:%T value:%V\n", n, n))
	}
}

// emitDefun emits the body for the given function definition "(defun ..)".
//
// TODO: this is basically a copy/paste of emitLambda
func (g *Generator) emitDefun(fn *Defun) {

	g.emitln(fn.Name + ":")

	g.emitln("    push rbp")
	g.emitln("    mov rbp, rsp")
	g.emitln("    sub rsp, 256 ;; guess at space for locals")

	env := NewEnv(nil)

	regs := []string{
		"rdi",
		"rsi",
		"rdx",
		"rcx",
		"r8",
		"r9",
	}

	for i, p := range fn.Params {
		offset := (i + 1) * 8

		env.slots[p] = offset

		g.emitln(fmt.Sprintf(
			"    mov [rbp-%d], %s",
			offset,
			regs[i],
		))
	}

	for _, xpr := range fn.Exprs {
		g.emitExpr(xpr, env)
	}

	g.emitln("    leave")
	g.emitln("    ret")
}

// emitLambda emits the body for the given lambda definition "(lambda ..)".
//
// TODO: this is basically a copy/paste of emitDefun.
func (g *Generator) emitLambda(l *Lambda) {

	g.emitln(l.name + ":")

	g.emitln("    push rbp")
	g.emitln("    mov rbp, rsp")
	g.emitln("    sub rsp, 256 ;; guess at space for locals")

	env := NewEnv(nil)

	regs := []string{
		"rdi",
		"rsi",
		"rdx",
		"rcx",
		"r8",
		"r9",
	}

	for i, p := range l.Params {
		offset := (i + 1) * 8

		env.slots[p] = offset

		g.emitln(fmt.Sprintf(
			"    mov [rbp-%d], %s",
			offset,
			regs[i],
		))
	}

	for _, xpr := range l.Exprs {
		g.emitExpr(xpr, env)
	}

	g.emitln("    leave")
	g.emitln("    ret")
}

func (g *Generator) Generate(defs []*Defun) string {

	defuns := ""

	// Now the user-defined functions
	for _, d := range defs {
		g.emitDefun(d)
		g.emitln("")
	}

	defuns = g.text.String()
	g.text.Reset()

	// Now user-defined lambdas
	lambdas := ""
	for _, l := range g.lambdas {
		g.emitLambda(l)
		g.emitln("")
	}
	lambdas = g.text.String()
	g.text.Reset()

	// Then the string-table for user-defined strings
	stringTable := ""
	g.emitln("section .data")
	for _, s := range g.strings {
		g.emitln("align 8")
		g.emitln(s.Label + ":")
		g.emitln(fmt.Sprintf("     db `%s`, 0", s.Value))
	}
	stringTable = g.text.String()
	g.text.Reset()

	type Generated struct {
		Defuns      string
		Lambdas     string
		StringTable string
	}
	x := &Generated{
		Defuns:      defuns,
		Lambdas:     lambdas,
		StringTable: stringTable,
	}

	buf := bytes.Buffer{}
	t1 := template.New("t1")
	t1 = template.Must(t1.Parse(tmplTxt))
	err := t1.Execute(&buf, x)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

// main
func main() {

	// CLI flags
	stdlib := flag.Bool("stdlib", true, "Prepend our Lisp standard library to user-programs")
	flag.Parse()

	// Do we have a file?
	if len(flag.Args()) != 1 {
		fmt.Println("usage: slisp [-stdlib=false] file.lisp")
		os.Exit(1)
	}

	// Read the file-contents
	data, err := os.ReadFile(flag.Args()[0])
	if err != nil {
		fmt.Printf("failed to read input %s: %s\n", os.Args[1], err)
		return
	}

	prg := string(data)
	if *stdlib {
		prg = stdlibLisp + "\n" + prg
	}

	// Parse into functions
	defs := parseProgram(prg)

	// Generate the code, and print it
	g := &Generator{}
	txt := g.Generate(defs)

	fmt.Print(txt)
}
