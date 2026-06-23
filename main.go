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
	"fmt"
	"os"
	"strconv"
	"strings"
)

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

type StringLiteral struct {
	Label string
	Value string
}
type Symbol struct {
	Name string
}

type Nil struct {
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
	Body     Expr
}

type Set struct {
	Name string
	Expr Expr
}

//
// Lexer
//

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

type Parser struct {
	tokens []string
	pos    int
}

func (p *Parser) peek() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	return p.tokens[p.pos]
}

func (p *Parser) next() string {
	t := p.peek()
	p.pos++
	return t
}

func (p *Parser) expect(s string) {
	if p.next() != s {
		panic("expected " + s)
	}
}

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

			body := p.parseExpr()
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
		default:
			var args []Expr

			for p.peek() != ")" {
				args = append(args, p.parseExpr())
			}

			p.expect(")")

			return &Call{
				Fn:   &Symbol{Name: sym.Name},
				Args: args,
			}
		}
	}

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

func NewEnv(parent *Env) *Env {
	return &Env{
		parent: parent,
		slots:  map[string]int{},
	}
}

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

func asmName(name string) string {
	switch name {
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

func (g *Generator) emitExpr(e Expr, env *Env) {
	switch n := e.(type) {
	case *Call:
		if symbol, ok := n.Fn.(*Symbol); ok {
			if symbol.Name == "<=" {
				g.emitExpr(n.Args[0], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    push rax")
				g.emitExpr(n.Args[1], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    pop rbx")
				g.emitln("    call lt_equals")
				return
			}

			if symbol.Name == "<" {
				g.emitExpr(n.Args[0], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    push rax")
				g.emitExpr(n.Args[1], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    pop rbx")
				g.emitln("    call lt")
				return
			}

			if symbol.Name == ">" {
				g.emitExpr(n.Args[0], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    push rax")
				g.emitExpr(n.Args[1], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    pop rbx")
				g.emitln("    call gt")
				return
			}

			if symbol.Name == ">=" {
				g.emitExpr(n.Args[0], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    push rax")
				g.emitExpr(n.Args[1], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    pop rbx")
				g.emitln("    call gt_equals")
				return
			}

			if symbol.Name == "+" {
				g.emitExpr(n.Args[0], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    push rax")
				g.emitExpr(n.Args[1], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    pop rbx")
				g.emitln("    add rax, rbx")
				g.emitln("    TAG_INTEGER_REG rax")
				return
			}

			if symbol.Name == "-" {
				g.emitExpr(n.Args[0], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    push rax")
				g.emitExpr(n.Args[1], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    pop rbx")
				g.emitln("    sub rbx, rax")
				g.emitln("    mov rax, rbx")
				g.emitln("    TAG_INTEGER_REG rax")
				return
			}

			if symbol.Name == "*" {
				g.emitExpr(n.Args[0], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    push rax")
				g.emitExpr(n.Args[1], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    pop rbx")
				g.emitln("    imul rbx, rax")
				g.emitln("    mov rax, rbx")
				g.emitln("    TAG_INTEGER_REG rax")
				return
			}

			if symbol.Name == "/" {
				g.emitExpr(n.Args[0], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    push rax")
				g.emitExpr(n.Args[1], env)
				g.emitln("    UNTAG_REG rax")
				g.emitln("    pop rbx")      // rax and rbx have args
				g.emitln("    mov rcx, rax") // meh
				g.emitln("    mov rax, rbx")
				g.emitln("    xor rdx, rdx")
				g.emitln("    idiv rcx")
				g.emitln("    TAG_INTEGER_REG rax")
				return
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
		child := NewEnv(env)

		nextSlot := len(child.slots)

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

		g.emitExpr(n.Body, child)

	case *Nil:
		g.emitln("    mov rax, 0       ; NIL")
		g.emitln("    TAG_NIL_REG rax  ; Tagged")

	case *String:
		lbl := g.label("str")
		g.strings = append(
			g.strings,
			StringLiteral{
				Label: lbl,
				Value: n.Value,
			},
		)
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

		panic(fmt.Sprintf("%T %V\n", n, n))
	}
}

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

// emitRuntime outputs our standard library;
// newline, print, and printstr.
func (g *Generator) emitRuntime() {
	stdlib := `
section .text

;; Is the given value an integer?
intp:
    mov rax, rdi
    and rax, 7
    cmp rax, 0
    jnz .nil
    mov rax, 1
    TAG_INTEGER_REG rax
    ret
.nil:
    mov rax, 0
    TAG_NIL_REG rax
    ret

;; Is the given value a string?
strp:
    mov rax, rdi
    and rax, 7
    cmp rax, 1
    jnz .nil
    mov rax, 1
    TAG_INTEGER_REG rax
    ret
.nil:
    mov rax, 0
    TAG_NIL_REG rax
    ret

;; Is the given value a cons?
consp:
    mov rax, rdi
    and rax, 7
    cmp rax, 2
    jnz .nil
    mov rax, 1
    TAG_INTEGER_REG rax
    ret
.nil:
    mov rax, 0
    TAG_NIL_REG rax
    ret

;; is the given value a lambda?
lambdap:
    mov rax, rdi
    and rax, 7
    cmp rax, 3
    jnz .nil
    mov rax, 1
    TAG_INTEGER_REG rax
    ret
.nil:
    mov rax, 0
    TAG_NIL_REG rax
    ret

;; Is the given value nil?
nilp:
    mov rax, rdi
    and rax, 7
    cmp rax, 7
    jnz .nil
    mov rax, 1
    TAG_INTEGER_REG rax
    ret
.nil:
    mov rax, 0
    TAG_NIL_REG rax
    ret

;; <=
lt_equals:
    cmp rbx, rax
    jle .true
    mov rax, 0
    TAG_NIL_REG rax
    ret
.true:
    mov rax, 1
    TAG_INTEGER_REG rax
    ret

;; >=
gt_equals:
    cmp rbx, rax
    jge .true
    mov rax, 0
    TAG_NIL_REG rax
    ret
.true:
    mov rax, 1
    TAG_INTEGER_REG rax
    ret

;; <
lt:
    cmp rbx, rax
    jl .true
    mov rax, 0
    TAG_NIL_REG rax
    ret
.true:
    mov rax, 1
    TAG_INTEGER_REG rax
    ret

;; >
gt:
    cmp rbx, rax
    jg .true
    mov rax, 0
    TAG_NIL_REG rax
    ret
.true:
    mov rax, 1
    TAG_INTEGER_REG rax
    ret

;; Print an integer
printint:
    push rbp
    mov rbp, rsp
    sub rsp, 64
    UNTAG_REG rax
    lea rsi, [rsp+63]     ;; pointer to end of buffer - terminated with NULL
    mov byte [rsi], 0
    mov rcx, 0
convert_loop:             ;; build up ASCII via divisions
    mov rbx, 10
    xor rdx, rdx
    div rbx
    add dl, '0'
    dec rsi
    mov [rsi], dl
    inc rcx
    test rax, rax          ;; keep going until we're zero
    jnz convert_loop

    mov rax, 1             ;; now write the buffer to STDOUT
    mov rdi, 1
    mov rdx, rcx
    syscall
    leave
    ret

;; terminate execution with the given exit-code
exit:
    UNTAG_REG rdi
    mov rax, 60     ; sys_exit
    syscall
    ret


;; print a newline.
newline:
    mov rax, 1      ; SYS_write
    mov rdi, 1      ; stdout
    mov rsi, newline_str
    mov rdx, 1
    syscall
    ret

;; Write the given ASCII character to stdout
putc:
    mov rax, rdi
    UNTAG_REG rax
    mov [putc_buffer],  al  ; store character
    mov rax, 1      ; SYS_write
    mov rdi, 1      ; stdout
    mov rsi, putc_buffer
    mov rdx, 1
    syscall
    ret

;; Print the given string.
printstr:
    UNTAG_REG rdi
    push rdi
    mov rdx, 0                ; length counter
.printstr_len_loop:
    cmp byte [rdi + rdx], 0
    je .printstr_found_len
    inc rdx
    jmp .printstr_len_loop
.printstr_found_len:
    pop rsi
    mov rax, 1      ; write
    mov rdi, 1      ; stdout
    syscall

    ret


;; Allocate space for a new cons area, return it in RAX
;; 16-bytes; first has the CAR, second the CDR.
cons:
    mov rax, [heap_ptr]      ; get the value of the heap pointer
    add qword [heap_ptr], 16 ; bump it by 2x ptrs
    mov [rax], rdi           ; store first item
    mov [rax+8], rsi         ; second item
    TAG_CONS_REG rax         ; return the tagged allocation
    ret

;; Get the first item in the cons.
car:
    mov rbx, rdi
    and rbx, 7
    cmp rbx, 2
    jne type_error
    UNTAG_REG rdi
    mov rax, [rdi]
    ret

;; Get the second item in the cons.
cdr:
    mov rbx, rdi
    and rbx, 7
    cmp rbx, 2
    jne type_error
    UNTAG_REG rdi
    mov rax, [rdi + 8]
    ret

type_error:
    jmp exit

section .data

;; buffer for "\n", used by (newline)
align 8
newline_str:
    db 10

;; buffer for storing a single character, used by (putc x)
align 8
putc_buffer:
    db 10

;; zero section
section .bss

;; heap storage
align 16
heap:
    resb 1048576

;; offset used of our heap area.
heap_ptr:
    resq 1
	`
	g.emitln(stdlib)
}

func (g *Generator) Generate(defs []*Defun) string {

	// Write out our header.
	header := `

%macro TAG_INTEGER_REG 1
    sal %1, 3
%endmacro

%macro TAG_STRING_REG 1
    sal %1, 3
    or %1, 1
%endmacro

%macro TAG_CONS_REG 1
    sal %1, 3
    or %1, 2
%endmacro

%macro TAG_LAMBDA_REG 1
    sal %1, 3
    or %1, 3
%endmacro

%macro TAG_NIL_REG 1
    sal %1, 3
    or %1, 7
%endmacro

%macro UNTAG_REG 1
    and %1, -8
    sar %1, 3
%endmacro



global _start

section .text
`
	g.emitln(header)

	// Now the user-defined functions
	for _, d := range defs {
		g.emitDefun(d)
		g.emitln("")
	}

	// Now user-defined lambdas
	for _, l := range g.lambdas {
		g.emitLambda(l)
		g.emitln("")
	}

	// Add our entry-point
	entry := `
_start:
    mov rax, heap        ; setup our heap pointer
    mov [heap_ptr], rax  ; cons cells are heap-allocated

    call main
    mov rdi, rax
    shr rdi, 1
    mov rax, 60
    syscall
`
	g.emitln(entry)

	// Then the string-table for user-defined strings
	g.emitln("section .data")
	for _, s := range g.strings {
		g.emitln("align 8")
		g.emitln(s.Label + ":")
		g.emitln(fmt.Sprintf("     db \"%s\", 0", s.Value))
	}

	// Finally add the runtime functions
	g.emitRuntime()
	return g.text.String()
}

// main
func main() {

	if len(os.Args) != 2 {
		fmt.Println("usage: slisp file.lisp")
		os.Exit(1)
	}

	// Read the file-contents
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("failed to read input %s: %s\n", os.Args[1], err)
		return
	}

	// Parse into functions
	defs := parseProgram(string(data))

	// Generate the code, and print it
	g := &Generator{}
	txt := g.Generate(defs)

	fmt.Print(txt)
}
