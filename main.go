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
// [ ] <
// [ ] >
// [ ] >=
//
// Special forms
// [X] DEFUN
// [X] IF
// [X] LET
// [X] INT?   // true if value is integer.
// [X] STR?   // true if value is string.
//
// Standard library:
// [X] PRINTINT
// [X] PRINTSTR
// [X] EXIT
// [X] NEWLINE
// [X] PUTC
//
// The lower bit of values is used for type:
//
//	0:  INT
//	1:  STRING
//
// ABI uses the Sys V style, so max six arguments:
// arg0 -> rdi
// arg1 -> rsi
// arg2 -> rdx
// arg3 -> rcx
// arg4 -> r8
// arg5 -> r9
//
// TODO: Need to do better with typing to allow CONS, LAMBDA, & etc.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

//
// AST
//

type Expr interface{}

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

type Call struct {
	Fn   string
	Args []Expr
}

type If struct {
	Cond Expr
	Then Expr
	Else Expr
}

type Binding struct {
	Name string
	Expr Expr
}

type Defun struct {
	Name   string
	Params []string
	Body   Expr
}

type Do struct {
	Exprs []Expr
}

type Let struct {
	Bindings []Binding
	Body     Expr
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

	body := p.parseExpr()

	p.expect(")")

	return &Defun{
		Name:   name,
		Params: params,
		Body:   body,
	}
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

	// symbol
	return &Symbol{Name: t}
}

func (p *Parser) parseList() Expr {
	p.expect("(")

	head := p.next()

	switch head {

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

	default:
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
	text    strings.Builder
	labelID int
	strings []StringLiteral
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

func (g *Generator) emitExpr(e Expr, env *Env) {
	switch n := e.(type) {

	case *Int:
		g.emitln(fmt.Sprintf("    mov rax, %d", n.Value<<1))

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
		g.emitln("    or rax, 1 ; add tagging")

	case *Symbol:
		offset, ok := env.Lookup(n.Name)
		if !ok {
			panic("unknown symbol: " + n.Name)
		}

		g.emitln(fmt.Sprintf(
			"    mov rax, [rbp-%d]",
			offset,
		))

	case *If:
		elseLbl := g.label("else")
		endLbl := g.label("endif")

		g.emitExpr(n.Cond, env)

		g.emitln("    cmp rax, 0")
		g.emitln("    je " + elseLbl)

		g.emitExpr(n.Then, env)

		g.emitln("    jmp " + endLbl)

		g.emitln(elseLbl + ":")

		// else branch is optional
		if n.Else != nil {
			g.emitExpr(n.Else, env)
		}
		g.emitln(endLbl + ":")

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

	case *Call:

		if n.Fn == "<=" {
			g.emitExpr(n.Args[0], env)
			g.emitln("    sar rax, 1") // remove typing bit
			g.emitln("    push rax")

			g.emitExpr(n.Args[1], env)
			g.emitln("    sar rax, 1") // remove typing bit

			g.emitln("    pop rbx")

			g.emitln("    cmp rbx, rax")
			g.emitln("    setle al")
			g.emitln("    movzx rax, al")
			g.emitln("    sal rax, 1") // add typing
			return
		}

		if n.Fn == "+" {
			g.emitExpr(n.Args[0], env)
			g.emitln("    sar rax, 1") // remove typing bit
			g.emitln("    push rax")

			g.emitExpr(n.Args[1], env)
			g.emitln("    sar rax, 1") // remove typing bit

			g.emitln("    pop rbx")
			g.emitln("    add rax, rbx")
			g.emitln("    sal rax, 1")

			return
		}

		if n.Fn == "-" {
			g.emitExpr(n.Args[0], env)
			g.emitln("    sar rax, 1") // remove typing-bit
			g.emitln("    push rax")

			g.emitExpr(n.Args[1], env)
			g.emitln("    sar rax, 1") // remove typing-bit

			g.emitln("    pop rbx")
			g.emitln("    sub rbx, rax")
			g.emitln("    mov rax, rbx")
			g.emitln("    sal rax, 1")

			return
		}

		if n.Fn == "*" {
			g.emitExpr(n.Args[0], env)
			g.emitln("    sar rax, 1") // remove typing-bit
			g.emitln("    push rax")

			g.emitExpr(n.Args[1], env)
			g.emitln("    sar rax, 1") // remove typing-bit

			g.emitln("    pop rbx")
			g.emitln("    imul rbx, rax")
			g.emitln("    mov rax, rbx")
			g.emitln("    sal rax, 1")

			return
		}

		if n.Fn == "/" {
			g.emitExpr(n.Args[0], env)
			g.emitln("    sar rax, 1") // remove typing-bit
			g.emitln("    push rax")
			g.emitExpr(n.Args[1], env)
			g.emitln("    sar rax, 1")   // remove typing-bit
			g.emitln("    pop rbx")      // rax and rbx have args
			g.emitln("    mov rcx, rax") // meh
			g.emitln("    mov rax, rbx")
			g.emitln("    xor rdx, rdx")
			g.emitln("    idiv rcx")
			g.emitln("    sal rax, 1") // restore typing
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

		for i, a := range n.Args {
			g.emitExpr(a, env)
			g.emitln(fmt.Sprintf(
				"    mov %s, rax",
				regs[i],
			))
		}

		g.emitln("    call " + n.Fn)

	case *Do:

		for _, expr := range n.Exprs {
			g.emitExpr(expr, env)
		}
	default:
		panic("unknown node")
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

	g.emitExpr(fn.Body, env)

	g.emitln("    leave")
	g.emitln("    ret")
}

// emitRuntime outputs our standard library;
// newline, print, and printstr.
func (g *Generator) emitRuntime() {
	stdlib := `

; input:  rax = value
; output: rax = 1 if low bit is 0
;          rax = 0 if low bit is 1

int?:
    mov rax, rdi ; first arg in rdi
    test rax, 1
    setz al
    movzx rax, al
    ret

str?:
    mov rax, rdi ; first arg in rdi
    test rax, 1
    setnz al
    movzx rax, al
    ret


section .text
printint:
    push rbp
    mov rbp, rsp
    sub rsp, 64
    sar rax, 1            ;; remove typing-bit
    lea rsi, [rsp+63]     ;; pointer to end of buffer - terminated with newline.
    mov byte [rsi], 10
    mov rcx, 1
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

section .data
align 8
newline_str:
    db 10

section .text

exit:
    sar rdi, 1      ; remove typing
    mov rax, 60     ; sys_exit
    syscall
    ret

newline:
    mov rax, 1      ; SYS_write
    mov rdi, 1      ; stdout
    mov rsi, newline_str
    mov rdx, 1
    syscall
    ret


section .data
align 8
putc_buffer:
    db 10

section .text

putc:
    sar rax, 1                      ; remove type
    mov [putc_buffer],  al  ; store character
    mov rax, 1      ; SYS_write
    mov rdi, 1      ; stdout
    mov rsi, putc_buffer
    mov rdx, 1
    syscall
    ret

; RDI = pointer to null-terminated string
printstr:
    and rdi, -2               ; remove tagging bit
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
	`
	g.emitln(stdlib)
}

func (g *Generator) Generate(defs []*Defun) string {

	g.emitln("global _start")
	g.emitln("")
	g.emitln("section .text")
	g.emitln("")

	for _, d := range defs {
		g.emitDefun(d)
		g.emitln("")
	}

	g.emitln("_start:")
	g.emitln("    call main")
	g.emitln("    mov rdi, rax")
	g.emitln("    shr rdi, 1")
	g.emitln("    mov rax, 60")
	g.emitln("    syscall")

	g.emitln("section .data")
	for _, s := range g.strings {
		g.emitln("align 8")
		g.emitln(s.Label + ":")
		g.emitln(fmt.Sprintf("     db \"%s\", 0", s.Value))
	}

	g.emitln("section .text")
	g.emitRuntime()
	return g.text.String()
}

//
// main
//

func main() {

	if len(os.Args) != 2 {
		fmt.Println("usage: compiler file.lisp")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	defer f.Close()

	var src strings.Builder

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		src.WriteString(sc.Text())
		src.WriteString("\n")
	}

	defs := parseProgram(src.String())

	g := &Generator{}

	fmt.Print(g.Generate(defs))
}
